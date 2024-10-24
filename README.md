# Sample XDS Server

Adopted from [this](https://github.com/salrashid123/grpc_xds)

This is sample XDS server Application without Kubernetes and Envoy.
The list of servers to load balance are specified in the `config.yaml` file.

## Steps

* Read configuration `YAML` file for list of services and their endpoints
* Do Initial scan and get the health status using the health API
* Create xDS resources and initialize cache for xDS
* Start the xDS GRPC Server
* Start a timer with scan interval
* Upon timer expiry re-scan the endpoints for their health status and update the xDS resources

## Demo

1. Run the xDS server `just run-xds-server`
2. Start the servers in multiple terminals `just run-server-1`, `just run-server-2`, `just run-server-3`
3. xDS server should show healthy status for the servers which are running.
4. Start the client `just run-client-with-xds`
5. Observe that requests are routed to all three servers
6. Bring down any of the server, and the requests to them should disappear after the next scan
7. Bring down the the xds-server. The client should continue to send the requests to the last status.