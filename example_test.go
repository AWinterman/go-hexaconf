package hexconf_test

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	hexconf "gitlab.com/wintersparkle/go-hexconf"
)

// ExampleRead_env demonstrates how to populate a struct with
// values from the environment, in the simplest case.
func ExampleRead_env() {
	type Config struct {
		URL  string `env:"URL"`
		User string `env:"USER"`
	}

	err := errors.Join(
		os.Setenv("URL", "http://example.com"),
		os.Setenv("USER", "sparkles"),
	)
	if err != nil {
		panic(err)
	}

	envConf := &Config{}
	err = hexconf.Read(envConf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", envConf)
	// Output:
	// &{URL:http://example.com User:sparkles}
}

// ExampleRead_env_nested demonstrates how this library will
// handle nested structs
func ExampleRead_env_nested() {
	type SubConfig struct {
		Count uint16 `env:"COUNT"`
	}
	type Config struct {
		URL       string    `env:"URL"`
		SubConfig SubConfig // any tag here would be ignored
		Count     uint16    `env:"COUNT"` // collisions are no problem
	}

	err := errors.Join(
		os.Setenv("URL", "http://example.com"),
		os.Setenv("COUNT", "32"),
	)
	if err != nil {
		panic(err)
	}

	envConf := &Config{}
	err = hexconf.Read(envConf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", envConf)
	// Output:
	// &{URL:http://example.com SubConfig:{Count:32} Count:32}
}

// ExampleRead_yaml demonstrates overlays of values from the
// environment and a YAML file.
func ExampleRead_yaml() {
	type Config struct {
		URL         string           `env:"URL"`
		Complicated []map[string]any `yaml:"complicated"`
	}

	err := errors.Join(
		os.Setenv("URL", "http://example.com"),
		os.Setenv("COUNT", "32"),
	)
	if err != nil {
		slog.Error("could not set env", "err", err)
		panic(err)
	}

	envConf := &Config{}
	err = hexconf.Read(envConf, "./example.yaml")
	if err != nil {
		slog.Error("could not read conf", "err", err)
		panic(err)
	}

	fmt.Printf("%+v", envConf)
	// Output:
	// &{URL:http://example.com Complicated:[map[a:1 b:2 c:map[a:1 b:2]]]}
}
