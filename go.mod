module github.com/kubensage/relay

go 1.24.4

// replace github.com/kubensage/common => /home/kubensage/common

require (
	github.com/google/uuid v1.6.0
	github.com/kubensage/common v0.0.2
	go.uber.org/zap v1.27.0
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.10
)

require (
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/net v0.44.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251006185510-65f7160b3a87 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
)
