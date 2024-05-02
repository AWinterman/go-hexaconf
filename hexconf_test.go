package hexconf

import (
	"errors"
	"os"
	"testing"

	"github.com/matryer/is"
	"gopkg.in/yaml.v3"
)

type testEnv map[string]string

func (t testEnv) LookupEnv(s string) (string, bool) {
	v, ok := t[s]
	return v, ok
}

func TestRead(t *testing.T) {
	err := errors.Join(
		os.Setenv("A", "alpha"),
		os.Setenv("B", "-10"),
		os.Setenv("C", "1"),
		os.Setenv("D", "true"),
		os.Setenv("F", "16"),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		os.Unsetenv("A")
		os.Unsetenv("B")
		os.Unsetenv("C")
		os.Unsetenv("D")
		os.Unsetenv("F")
	}()

	type Conf struct {
		A string `env:"A"`
		B int    `env:"B"`
		C uint   `env:"C"`
		D bool   `env:"D"`
		E []string
		F uint8 `env:"F"`
		G uint8 `env:"G"`
	}

	conf := Conf{
		E: []string{"a", "b", "c"},
		G: 23,
	}

	b, err := yaml.Marshal(conf)
	if err != nil {
		t.Fatal(err)
	}
	path, err := os.MkdirTemp("./", "hexaconf-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(path)

	path = path + "/conf.yaml"

	err = os.WriteFile(path, b, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = Read(&conf, path)
	if err != nil {
		t.Fatal(err)
	}

	is := is.New(t)

	is.Equal(conf.A, "alpha")
	is.Equal(conf.B, -10)
	is.Equal(conf.C, uint(1))
	is.Equal(conf.D, true)
	is.Equal(conf.E, []string{"a", "b", "c"})
	is.Equal(conf.F, uint8(16))
	is.Equal(conf.G, uint8(23))
}

func TestReadEnv(t *testing.T) {
	type unsized struct {
		A string `env:"A"`
		B int    `env:"B"`
		C uint   `env:"C"`
		D bool   `env:"D"`
	}

	u := unsized{}

	err := readEnv(&u, testEnv{"A": "alpha", "B": "-10", "C": "1", "D": "true"})
	if err != nil {
		t.Fatal(err)
	}
	expected := unsized{"alpha", -10, 1, true}

	if u != expected {
		t.Fatalf("expected equality; got %+v", u)
	}
}
