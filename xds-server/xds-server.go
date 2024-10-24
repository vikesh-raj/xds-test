package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"net"

	"sync"
	"sync/atomic"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	xds "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	health_api "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const grpcMaxConcurrentStreams = 1000

type Server struct {
	version  int32
	cache    cache.SnapshotCache
	config   Config
	nodeId   string
	fetches  int
	requests int
	mu       sync.Mutex
	logger   *slog.Logger
}

func (s *Server) PrintStats() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logger.Info("Stats", "fetches", s.fetches, "requests", s.requests)
}

func (s *Server) OnStreamOpen(ctx context.Context, id int64, typ string) error {
	s.logger.Info("OnStreamOpen", "id", id, "type", typ)
	return nil
}

func (s *Server) OnStreamClosed(id int64, cn *core.Node) {
	s.logger.Info("OnStreamClosed", "id", id, "node", cn.String())
}

func (s *Server) OnStreamRequest(id int64, r *discovery.DiscoveryRequest) error {
	s.logger.Info("OnStreamRequest", "id", id, "request_url", r.TypeUrl, "request", r.String())
	s.mu.Lock()
	defer s.mu.Unlock()
	s.requests++
	return nil
}

func (s *Server) OnStreamResponse(ctx context.Context, id int64, req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
	s.logger.Info("OnStreamResponse", "id", id, "request_url", req.TypeUrl,
		"response_url", resp.TypeUrl)
	s.PrintStats()
}

func (s *Server) OnFetchRequest(ctx context.Context, req *discovery.DiscoveryRequest) error {
	s.logger.Info("OnFetchRequest", "request_url", req.TypeUrl, "request", req)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fetches++
	return nil
}

func (s *Server) OnFetchResponse(req *discovery.DiscoveryRequest, resp *discovery.DiscoveryResponse) {
	s.logger.Info("OnFetchResponse", "request_url", req.TypeUrl, "response_url", resp.TypeUrl)
}

func (s *Server) OnDeltaStreamClosed(id int64, cn *core.Node) {
	s.logger.Info("OnDeltaStreamClosed", "id", id)
}

func (s *Server) OnDeltaStreamOpen(ctx context.Context, id int64, typ string) error {
	s.logger.Info("OnDeltaStreamOpen", "id", id, "type", typ)
	return nil
}

func (s *Server) OnStreamDeltaRequest(id int64, request *discovery.DeltaDiscoveryRequest) error {
	s.logger.Info("OnStreamDeltaRequest", "id", id)
	return nil
}

func (s *Server) OnStreamDeltaResponse(id int64, request *discovery.DeltaDiscoveryRequest, response *discovery.DeltaDiscoveryResponse) {
	s.logger.Info("OnStreamDeltaResponse", "id", id)
}

func (s *Server) healthCheck(service *Service, ep string) bool {
	ep_logger := s.logger.With("name", service.Name, "endpoint", ep)
	conn, err := grpc.NewClient(ep, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ep_logger.Error("Unable to connect to Health Endpoint", "error", err)
		return false
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client := health_api.NewHealthClient(conn)
	request := health_api.HealthCheckRequest{}
	request.Service = service.FullName
	response, err := client.Check(ctx, &request)
	if err != nil {
		ep_logger.Error("Health Check failed", "error", err)
		return false
	}
	ep_logger.Info("Health response", "response", response.String())
	return response.Status == health_api.HealthCheckResponse_SERVING
}

func (s *Server) scanService(service *Service) map[resource.Type]types.Resource {
	name := service.Name
	var lb_endpoints []*endpoint.LbEndpoint

	output := make(map[resource.Type]types.Resource)

	for i := range service.Endpoints {
		ep := service.Endpoints[i]
		splits := strings.SplitN(ep, ":", 2)
		if len(splits) == 1 {
			s.logger.Error("Ignoring bad endpoint", "service", name,
				"endpoint", ep)
			continue
		}
		hostname := splits[0]
		if hostname == "" {
			hostname = "localhost"
		}
		port, err := strconv.Atoi(splits[1])
		if err != nil {
			s.logger.Error("Invalid port", "service", name, "endpoint", ep)
			continue
		}

		hst := &core.Address{Address: &core.Address_SocketAddress{
			SocketAddress: &core.SocketAddress{
				Address:  hostname,
				Protocol: core.SocketAddress_TCP,
				PortSpecifier: &core.SocketAddress_PortValue{
					PortValue: uint32(port),
				},
			},
		}}

		health_status := core.HealthStatus_UNKNOWN
		if s.healthCheck(service, ep) {
			health_status = core.HealthStatus_HEALTHY
		} else {
			health_status = core.HealthStatus_UNHEALTHY
		}

		epp := &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: hst,
				}},
			HealthStatus: health_status,
		}
		lb_endpoints = append(lb_endpoints, epp)
	}

	clusterName := name + "-cluster"
	routeConfigName := name + "-route"
	virtualHostName := name + "-vs"

	output[resource.EndpointType] = &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			Locality: &core.Locality{
				Region: service.Zone,
				Zone:   service.Zone,
			},
			Priority:            0,
			LoadBalancingWeight: &wrapperspb.UInt32Value{Value: uint32(1000)},
			LbEndpoints:         lb_endpoints,
		}},
	}

	output[resource.ClusterType] = &cluster.Cluster{
		Name:                 clusterName,
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{},
			},
		},
	}

	output[resource.RouteType] = &route.RouteConfiguration{
		Name:             routeConfigName,
		ValidateClusters: &wrapperspb.BoolValue{Value: true},
		VirtualHosts: []*route.VirtualHost{{
			Name:    virtualHostName,
			Domains: []string{name},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			},
			},
		}},
	}

	dummy_router, _ := anypb.New(&router.Router{})
	dummy_api_listener, _ := anypb.New(&hcm.HttpConnectionManager{
		CodecType: hcm.HttpConnectionManager_AUTO,
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				RouteConfigName: routeConfigName,
				ConfigSource: &core.ConfigSource{
					ResourceApiVersion: core.ApiVersion_V3,
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: dummy_router,
			},
		}},
	})
	output[resource.ListenerType] = &listener.Listener{
		Name: name,
		ApiListener: &listener.ApiListener{
			ApiListener: dummy_api_listener,
		},
	}

	return output
}

func (s *Server) scan() {
	resources := make(map[resource.Type][]types.Resource)
	for i := range s.config.Services {
		res := s.scanService(&s.config.Services[i])
		for typ, resource := range res {
			resources[typ] = append(resources[typ], resource)
		}
	}
	atomic.AddInt32(&s.version, 1)
	s.logger.Info("creating snapshot Version " + fmt.Sprint(s.version))
	snap, err := cache.NewSnapshot(fmt.Sprint(s.version), resources)
	if err != nil {
		s.logger.Error("Unable to create snapshot", "error", err)
		return
	}

	err = s.cache.SetSnapshot(context.Background(), s.nodeId, snap)
	if err != nil {
		s.logger.Error("Unable to set snapshot", "error", err)
	}
}

func main() {
	config_file := flag.String("config", "config.yaml", "Config file")
	port := flag.Uint("port", 50050, "xDS Listen port")
	flag.Parse()

	ctx := context.Background()
	config, err := ReadConfig(*config_file)
	if err != nil {
		fmt.Println("Unable to open config file : ", *config_file, ", Error : ", err.Error())
		return
	}

	s := &Server{
		config:   config,
		cache:    cache.NewSnapshotCache(true, cache.IDHash{}, nil),
		nodeId:   config.Name,
		fetches:  0,
		requests: 0,
		logger:   slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}

	// Initial Scan
	s.logger.Info("Running Initial Scan..")
	s.scan()

	s.logger.Info("Starting xds server, with services : ", "num_services", len(config.Services), "node", s.nodeId)
	srv := xds.NewServer(ctx, s.cache, s)

	grpcServer := grpc.NewServer(grpc.MaxConcurrentStreams(grpcMaxConcurrentStreams))
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		s.logger.Error("failed to listen", "error", err, "port", *port)
		os.Exit(1)
		return
	}
	discovery.RegisterAggregatedDiscoveryServiceServer(grpcServer, srv)
	go func() {
		if err = grpcServer.Serve(lis); err != nil {
			s.logger.Error("GRPC Serve error", "error", err)
		}
	}()

	ticker := time.NewTicker(time.Duration(s.config.ScanInterval) * time.Second)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case <-ticker.C:
			s.scan()
		case <-c:
			ticker.Stop()
			grpcServer.GracefulStop()
			return
		}
	}
}
