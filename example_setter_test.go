package hexconf_test

import (
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"

	"gitlab.com/wintersparkle/go-hexconf"
)

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
