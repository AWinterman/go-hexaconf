// Package hexconf provides a simple way to read configuration from environment variables or from yaml files.
package hexconf

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"

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

	err := readEnv(into, LookupEnv(os.LookupEnv))
	if err != nil {
		return fmt.Errorf("%w reading environment variables", err)
	}

	return nil
}

type LookupEnv func(string) (string, bool)

type EnvLooker interface {
	LookupEnv(string) (string, bool)
}

// LookupEnv implements EnvLooker
func (l LookupEnv) LookupEnv(s string) (string, bool) {
	return l(s)
}

func readEnv(into any, env EnvLooker) error {
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

func setFields(rv reflect.Value, env EnvLooker) error {
	for i := 0; i < rv.NumField(); i++ {
		rf := rv.Field(i)
		tags := rv.Type().Field(i).Tag
		set := setter(rf.Kind(), env)
		v, ok := env.LookupEnv(tags.Get("env"))
		if !ok {
			continue
		}
		err := set(rf, v)
		if err != nil {
			return fmt.Errorf("%w setting %q to %q", err, rf.Type().Field(i).Name, v)
		}

	}
	return nil
}

func intSetter(rv reflect.Value, v string) error {
	integer, err := strconv.ParseInt(v, 0, 0)
	if err != nil {
		return fmt.Errorf("%w converting %q to int", err, v)
	}
	rv.SetInt(integer)
	return nil
}

func uintSetter(rv reflect.Value, v string) error {
	integer, err := strconv.ParseUint(v, 0, 0)
	if err != nil {
		return fmt.Errorf("%w converting %q to int", err, v)
	}
	rv.SetUint(integer)
	return nil
}

func setter(kind reflect.Kind, env EnvLooker) func(reflect.Value, string) error {
	return map[reflect.Kind]func(reflect.Value, string) error{
		reflect.String: func(rv reflect.Value, v string) error {
			rv.SetString(v)
			return nil
		},
		reflect.Int:    intSetter,
		reflect.Int8:   intSetter,
		reflect.Int16:  intSetter,
		reflect.Int32:  intSetter,
		reflect.Int64:  intSetter,
		reflect.Uint:   uintSetter,
		reflect.Uint8:  uintSetter,
		reflect.Uint16: uintSetter,
		reflect.Uint32: uintSetter,
		reflect.Uint64: uintSetter,
		reflect.Bool: func(rv reflect.Value, v string) error {
			b, err := strconv.ParseBool(v)
			if err != nil {
				return err
			}
			rv.SetBool(b)
			return nil
		},
		reflect.Invalid:       unsupportedType,
		reflect.Uintptr:       unsupportedType,
		reflect.Float32:       skip,
		reflect.Float64:       skip,
		reflect.Complex64:     skip,
		reflect.Complex128:    skip,
		reflect.Array:         skip,
		reflect.Chan:          skip,
		reflect.Func:          skip,
		reflect.Interface:     skip,
		reflect.Map:           skip,
		reflect.Pointer:       skip,
		reflect.Slice:         skip,
		reflect.Struct:        skip,
		reflect.UnsafePointer: skip,
	}[kind]
}

func skip(reflect.Value, string) error {
	return nil
}

func unsupportedType(rv reflect.Value, _ string) error {
	return fmt.Errorf("unsupported type %s", rv.Type().Kind())
}
