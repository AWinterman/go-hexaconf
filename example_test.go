package hexconf_test

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"gitlab.com/wintersparkle/go-hexconf"
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

	conf := &Config{}
	err = hexconf.Read(conf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", conf)
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

	conf := &Config{}
	err = hexconf.Read(conf)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v", conf)
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

	config := &Config{}

	err = hexconf.Read(config, "./example.yaml")
	if err != nil {
		slog.Error("could not read conf", "err", err)
		panic(err)
	}

	fmt.Printf("%+v", config)
	// Output:
	// &{URL:http://example.com Complicated:[map[a:1 b:2 c:map[a:1 b:2]]]}
}

// NetFlagValue is an example of a type that implements Setter
type NetFlagValue struct {
	url.URL
}

// Set implements flag.Value
func (v *NetFlagValue) Set(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return err
	}

	v.URL = *url
	return nil
}

// ExampleRead_setter demonstrates the package's support for Setter values
// clients with complicated env var parsing logic can use this to support their
// use case.
//
// Setter also provides interop the builtin flag package.
func ExampleRead_setter() {
	type Config struct {
		URL *NetFlagValue `env:"URL"`
	}

	err := os.Setenv("URL", "amqp://super:secret@localhost:5672")
	if err != nil {
		fmt.Println("could not set env", "err", err)
	}
	conf := &Config{URL: &NetFlagValue{}}

	flag.Var(conf.URL, "url", "an example url")

	err = hexconf.Read(conf)
	// flag.Parse() would populate based on command line flags
	if err != nil {
		slog.Error("could not read conf", "err", err)
		panic(err)
	}

	fmt.Println(conf.URL.Redacted())
	// Output:
	// amqp://super:xxxxx@localhost:5672
}
