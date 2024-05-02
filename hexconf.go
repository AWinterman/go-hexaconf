// Package hexconf provides a simple way to read configuration from environment variables or from yaml files.
package hexconf

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// Read reads configuration from the provided files, and populates the provided
// struct, sequentially applying each file. If a file is not found, it is skipped.
//
// After reading from the files, Read will populate the struct with values from
// environment variables, if they exist.
func Read(into any, files ...string) error {
	for _, file := range files {
		b, err := os.ReadFile(file)
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return fmt.Errorf("%w reading %q", err, file)
		}
		err = yaml.Unmarshal(b, into)
		if err != nil {
			return fmt.Errorf("%w unmarshaling contents of %q", err, file)
		}
	}

	err := readEnv(into, lookupEnv(os.LookupEnv))
	if err != nil {
		return fmt.Errorf("%w reading environment variables", err)
	}

	return nil
}

// Setter is implemented by types can self-deserialize values.
// Any type that implements flag.Value also implements Setter.
type Setter interface {
	Set(value string) error
}

// lookupEnv is a function that looks up an environment variable by name.
// It is provided to allow for convenient doubling of os env
type lookupEnv func(string) (string, bool)

// envLooker is an interface that allows for flexible environment variable lookup.
type envLooker interface {
	LookupEnv(string) (string, bool)
}

// LookupEnv implements EnvLooker
func (l lookupEnv) LookupEnv(s string) (string, bool) {
	return l(s)
}

// readEnv reads environment variables and sets the fields of the provided struct.
func readEnv(into any, env envLooker) error {
	rv := reflect.ValueOf(into)

	if rv.Type().Kind() == reflect.Interface {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Ptr {
		return errors.New("expected a pointer to a struct")
	}

	rv = rv.Elem()

	return setFields(rv, env)
}
