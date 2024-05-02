package hexconf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// setFields iterates the fields of the provided struct, extracting
func setFields(rv reflect.Value, env envLooker) error {
	for i := 0; i < rv.NumField(); i++ {
		rf := rv.Field(i)
		typ := rv.Type().Field(i)

		tags := typ.Tag
		envVar, hasTag := tags.Lookup("env")

		if typ.Type.Kind() == reflect.Struct {
			err := setFields(rf, env)
			if err != nil {
				return err
			}
		}

		if !hasTag {
			continue
		}

		v, ok := env.LookupEnv(envVar)
		if !ok {
			continue
		}

		set := getSetterFor(rf.Kind(), env)

		err := set(rf, v)
		if err != nil {
			return fmt.Errorf("%w setting %q to %q", err, typ.Name, v)
		}

	}
	return nil
}

// getSetterFor returns a function that can set the value of a field of the provided kind.
func getSetterFor(kind reflect.Kind, env envLooker) func(reflect.Value, string) error {

	fns := map[reflect.Kind]func(reflect.Value, string) error{
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
		reflect.Float32:       parseFloat,
		reflect.Float64:       parseFloat,
		reflect.Complex64:     skip,
		reflect.Complex128:    skip,
		reflect.Array:         skip,
		reflect.Chan:          skip,
		reflect.Func:          skip,
		reflect.Interface:     deref(env),
		reflect.Map:           skip,
		reflect.Pointer:       deref(env),
		reflect.Slice:         getSliceSetter(env),
		reflect.Struct:        structSetter(env),
		reflect.UnsafePointer: skip,
	}
	f, ok := fns[kind]
	if !ok {
		return unsupportedType
	}
	return f
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

func parseFloat(rv reflect.Value, v string) error {
	flt, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("%w converting %q to int", err, v)
	}
	rv.SetFloat(flt)
	return nil
}

func structSetter(env envLooker) func(reflect.Value, string) error {
	return func(rv reflect.Value, v string) error {
		ok, err := trySetterInterface(rv, v)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}
		return setFields(rv, env)
	}
}

func getSliceSetter(env envLooker) func(reflect.Value, string) error {
	return func(rv reflect.Value, v string) error {
		typ := rv.Type()
		var sl reflect.Value

		firstKind := typ.Elem().Kind()

		if firstKind == reflect.Uint8 {
			sl = reflect.ValueOf([]byte(v))
		} else if strings.TrimSpace(v) != "" {
			sl = reflect.MakeSlice(typ, 1, 1)
			set := getSetterFor(firstKind, env)

			err := set(sl.Index(0), v)
			if err != nil {
				return err
			}
		}

		rv.Set(sl)
		return nil
	}
}

func skip(rv reflect.Value, v string) error {
	return nil
}

func unsupportedType(rv reflect.Value, _ string) error {
	return fmt.Errorf("unsupported type %s", rv.Kind())
}

func deref(env envLooker) func(reflect.Value, string) error {
	return func(rv reflect.Value, v string) error {
		elem := rv.Elem()
		if rv.Kind() == reflect.Invalid {
			return unsupportedType(rv, v)
		}
		return getSetterFor(elem.Kind(), env)(elem, v)
	}
}

func trySetterInterface(field reflect.Value, value string) (bool, error) {
	if !field.CanInterface() {
		return false, nil
	}

	var s Setter
	var ok bool

	// is it directly a Setter?
	s, ok = field.Interface().(Setter)

	if !ok && field.CanAddr() {
		// is it a pointer to it a Setter?
		s, _ = (field.Addr().Interface()).(Setter)
	}

	err := s.Set(value)
	if err != nil {
		return false, err
	}

	return true, nil
}
