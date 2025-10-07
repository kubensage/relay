package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/kubensage/relay/pkg/buildinfo"
	"go.uber.org/zap"
)

// RelayConfig holds configuration parameters for the relay service.
//
// Fields:
//   - RelayAddress: TCP address where the relay's gRPC server will listen
//     for incoming agent connections. Typically, in the form "host:port".
type RelayConfig struct {
	RelayAddress string
}

// RegisterRelayFlags registers relay-specific command-line flags into the provided FlagSet.
//
// It returns a closure that, when invoked with a logger, validates the parsed flags
// and constructs a RelayConfig instance. This allows flag parsing to be performed
// in main(), while validation and construction are deferred until after flags are parsed.
//
// Required flag:
//
//	--relay-address string
//	  The address of the metrics relay gRPC server (e.g. "localhost:5000").
//
// Optional flags:
//
//	--version
//	  If set, prints the current agent version (as defined in pkg/buildinfo.Version) and exits.
//
// Parameters:
//   - fs *flag.FlagSet:
//     The flag set into which relay flags should be registered.
//
// Returns:
//   - func(logger *zap.Logger) *RelayConfig:
//     A function that validates the parsed flags, logs any fatal errors,
//     and returns a populated RelayConfig instance.
func RegisterRelayFlags(fs *flag.FlagSet) func(logger *zap.Logger) *RelayConfig {
	relayAddress := fs.String("relay-address", "localhost:50051", "TCP address where the relay will listen for gRPC traffic")
	version := fs.Bool("version", false, "Print the current version and exit")

	return func(logger *zap.Logger) *RelayConfig {
		// Handle version flag
		if *version {
			fmt.Printf("%s\n", buildinfo.Version)
			os.Exit(0)
		}

		if *relayAddress == "" {
			// Fatal is appropriate here because the relay cannot start without a listening address
			logger.Fatal("missing required flag: --relay-address")
		}

		return &RelayConfig{
			RelayAddress: *relayAddress,
		}
	}
}
