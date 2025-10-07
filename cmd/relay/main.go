package main

import (
	"context"
	"flag"
	"net"
	"os/signal"
	"syscall"

	gocli "github.com/kubensage/common/cli"
	golog "github.com/kubensage/common/log"
	"github.com/kubensage/relay/pkg/cli"
	grpc2 "github.com/kubensage/relay/pkg/grpc"
	"github.com/kubensage/relay/proto/gen"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const appName = "relay"

// main is the entrypoint of the relay process.
//
// It performs the following steps:
//  1. Registers and parses logging and relay configuration flags.
//  2. Initializes the logger and relay configuration.
//  3. Sets up a gRPC server listening on the configured address.
//  4. Handles graceful shutdown on SIGINT or SIGTERM.
func main() {
	// Register CLI flags for logging and relay configuration
	logCfgFn := gocli.RegisterLogStdFlags(flag.CommandLine)
	relayCfgFn := cli.RegisterRelayFlags(flag.CommandLine)

	flag.Parse()

	// Initialize logger and configuration
	logCfg := logCfgFn()
	logger := golog.SetupStdLogger(logCfg)
	relayCfg := relayCfgFn(logger)

	// Print startup configuration at INFO level
	golog.LogStartupInfo(logger, appName, logCfg, relayCfg)

	// Set up context that cancels on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start TCP listener
	listener, err := net.Listen("tcp", relayCfg.RelayAddress)
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	// Initialize gRPC server and register service
	grpcServer := grpc.NewServer()
	gen.RegisterMetricsServiceServer(grpcServer, grpc2.NewMetricsServer(logger))
	logger.Info("gRPC server listening", zap.String("address", relayCfg.RelayAddress))

	// Run gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			// Fatal is appropriate: serving loop should not exit unexpectedly
			logger.Fatal("failed to serve", zap.Error(err))
		}
	}()

	// Wait for termination signal
	<-ctx.Done()
	logger.Info("received termination signal, shutting down...")

	// Gracefully stop gRPC server
	grpcServer.GracefulStop()
	logger.Info("gRPC server stopped gracefully")
}
