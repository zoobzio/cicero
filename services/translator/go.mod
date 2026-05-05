module github.com/zoobz-io/cicero/translator

go 1.25.0

replace github.com/zoobz-io/cicero/proto => ../../proto

require (
	github.com/zoobz-io/cicero/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.79.1
)

require (
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)
