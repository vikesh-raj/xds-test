module github.com/vikesh-raj/xds-test/server

go 1.22

replace github.com/vikesh-raj/xds-test/echo => ../echo

require (
	github.com/vikesh-raj/xds-test/echo v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.30.0
	google.golang.org/grpc v1.67.1
)

require (
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241021214115-324edc3d5d38 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
)
