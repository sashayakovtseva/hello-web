package config

import (
	"errors"
	"fmt"

	"github.com/jessevdk/go-flags"
)

//nolint:lll
type (
	// AppConfig contains full configuration of the service.
	//nolint:govet
	AppConfig struct {
		DefaultFirstName      string `long:"default_fist_name" env:"DEFAULT_FIST_NAME" description:"Default fist name" default:""`
		DefaultLastNameConfig string `long:"default_last_name_config" env:"DEFAULT_LAST_NAME_CONFIG" description:"Default last name config" default:"local/default_last_name.json"`

		Consul Consul `group:"Consul options" namespace:"consul" env-namespace:"CONSUL"`
		HTTP   Server `group:"HTTP server options" namespace:"http" env-namespace:"HTTP"`
		GRPC   Server `group:"GRPC server options" namespace:"grpc" env-namespace:"GRPC"`
	}

	// Consul contains consul configuration of the service.
	// With nomad deployment this section should not be used.
	Consul struct {
		Scheme    string   `long:"scheme" env:"SCHEME" description:"Scheme to use to talk to consul agent" default:"http"`
		DC        string   `long:"dc" env:"DC" description:"Datacenter to connect to consul agent"`
		Host      string   `long:"host" env:"HOST" description:"Host to connect to consul agent" default:"127.0.0.1"`
		ExtraTags []string `long:"extra_tags" env:"EXTRA_TAGS" env-delim:"," description:"Extra tags to use during service registration"`
		Port      int      `long:"port" env:"PORT" description:"Port to connect to consul agent" default:"8500"`
		Enabled   bool     `long:"enabled" env:"ENABLED" description:"Feature toggle for automatic consul integration"`
	}

	// Server contains server configuration, regardless of the server type (http, grpc, etc).
	Server struct {
		Host string `long:"host" env:"HOST" description:"Host to listen on, default is empty (all interfaces)"`
		Port int    `long:"port" env:"PORT" description:"Port to listen on" required:"true"`
	}
)

// ErrHelp is returned when --help flag is
// used and application should not launch.
var ErrHelp = errors.New("help")

// New reads flags and envs and returns AppConfig
// that corresponds to the values read.
func New() (*AppConfig, error) {
	var config AppConfig
	if _, err := flags.Parse(&config); err != nil {
		var flagsErr *flags.Error
		if errors.As(err, &flagsErr) && flagsErr.Type == flags.ErrHelp {
			return nil, ErrHelp
		}
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}
