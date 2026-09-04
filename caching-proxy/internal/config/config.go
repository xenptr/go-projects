package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

const appName = "caching-proxy"

// Config holds the parsed CLI configuration.
type Config struct {
	Port   int
	Origin string
	Clear  bool
}

// Parse parses os.Args[1:] and returns a validated Config.
// It prints usage to stderr and exits on --help/-h, and returns an error
// for invalid input so the caller can decide how to handle it.
func Parse() (Config, error) {
	return parseFlags(os.Args[1:])
}

func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)

	// Route usage output through our custom printer so it goes to stderr.
	fs.Usage = func() { printUsage(fs) }

	portFlag := fs.Int("port", 0, "Port on which the caching proxy will listen (required unless --clear-cache)")
	originFlag := fs.String("origin", "", "URL of the origin server to forward requests to (required unless --clear-cache)")
	clearFlag := fs.Bool("clear-cache", false, "Clear the cache and exit")

	if err := fs.Parse(args); err != nil {
		// flag.ErrHelp means the user passed --help; that's not a real error.
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		return Config{}, fmt.Errorf("parsing flags: %w", err)
	}

	cfg := Config{
		Port:   *portFlag,
		Origin: *originFlag,
		Clear:  *clearFlag,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validate checks that the combination of flags makes sense.
func (c Config) validate() error {
	// --clear-cache is a standalone operation; no other flags are needed.
	if c.Clear {
		return nil
	}

	var errs []string

	if c.Port == 0 {
		errs = append(errs, "--port is required")
	} else if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Sprintf("--port must be between 1 and 65535, got %d", c.Port))
	}

	if c.Origin == "" {
		errs = append(errs, "--origin is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

// PrintUsage writes usage information to stderr.
func PrintUsage() {
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)
	// Register flags just so their defaults appear in the output.
	fs.Int("port", 0, "Port on which the caching proxy will listen (required unless --clear-cache)")
	fs.String("origin", "", "URL of the origin server to forward requests to (required unless --clear-cache)")
	fs.Bool("clear-cache", false, "Clear the cache and exit")
	printUsage(fs)
}

func printUsage(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", appName)
	fmt.Fprintf(os.Stderr, "A caching reverse proxy that forwards requests to an origin server\n")
	fmt.Fprintf(os.Stderr, "and caches responses for subsequent identical requests.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fs.SetOutput(os.Stderr)
	fs.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  Start the proxy on port 3000 forwarding to http://example.com:\n")
	fmt.Fprintf(os.Stderr, "    %s --port 3000 --origin http://example.com\n\n", appName)
	fmt.Fprintf(os.Stderr, "  Clear the cache:\n")
	fmt.Fprintf(os.Stderr, "    %s --clear-cache\n", appName)
}
