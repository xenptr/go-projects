package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const appName = "caching-proxy"

// Config holds both CLI flags and environment-loaded values.
type Config struct {
	// CLI flags
	Port   int
	Origin string
	Clear  bool

	// Loaded from environment
	RedisHost     string
	RedisPort     string
	RedisUsername string
	RedisPassword string
}

// Load reads the .env file (if present) then parses CLI flags, returning a
// fully populated and validated Config.
func Load() (Config, error) {
	// Load .env if it exists; silently ignore if it doesn't.
	_ = godotenv.Load()

	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		return Config{}, err
	}

	cfg.RedisHost = os.Getenv("REDIS_HOST")
	cfg.RedisPort = os.Getenv("REDIS_PORT")
	cfg.RedisUsername = os.Getenv("REDIS_USERNAME")
	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")

	if err := cfg.validateEnv(); err != nil {
		return Config{}, err
	}

	return cfg, nil
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

	if err := cfg.validateFlags(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validateFlags checks CLI flag combinations.
func (c Config) validateFlags() error {
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

// validateEnv checks that required environment variables are present.
// Skipped when --clear-cache is set since no Redis connection is needed.
func (c Config) validateEnv() error {
	if c.Clear {
		return nil
	}

	var errs []string

	if c.RedisHost == "" {
		errs = append(errs, "REDIS_HOST is required")
	}
	if c.RedisPort == "" {
		errs = append(errs, "REDIS_PORT is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, ", "))
	}

	return nil
}

// PrintUsage writes usage information to stderr.
func PrintUsage() {
	fs := flag.NewFlagSet(appName, flag.ContinueOnError)
	fs.Int("port", 0, "Port on which the caching proxy will listen (required unless --clear-cache)")
	fs.String("origin", "", "URL of the origin server to forward requests to (required unless --clear-cache)")
	fs.Bool("clear-cache", false, "Clear the cache and exit")
	printUsage(fs)
}

func printUsage(fs *flag.FlagSet) {
	fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", appName)
	// Register flags just so their defaults appear in the output.
	fmt.Fprintf(os.Stderr, "A caching reverse proxy that forwards requests to an origin server\n")
	fmt.Fprintf(os.Stderr, "and caches responses for subsequent identical requests.\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	fs.SetOutput(os.Stderr)
	fs.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nEnvironment variables (can be set in .env):\n")
	fmt.Fprintf(os.Stderr, "  REDIS_HOST      Redis server hostname (required)\n")
	fmt.Fprintf(os.Stderr, "  REDIS_PORT      Redis server port (required)\n")
	fmt.Fprintf(os.Stderr, "  REDIS_USERNAME  Redis ACL username (optional)\n")
	fmt.Fprintf(os.Stderr, "  REDIS_PASSWORD  Redis password (optional)\n")
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s --port 3000 --origin http://example.com\n", appName)
	fmt.Fprintf(os.Stderr, "  %s --clear-cache\n", appName)
}
