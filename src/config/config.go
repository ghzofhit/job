package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const DefaultSystemConfigPath = "etc/job.conf"

type Config struct {
	SystemPath string
	First      string `toml:"conf_first" env:"CONF_FIRST"`
}

func New() *Config {
	c := new(Config)
	c.SystemPath = DefaultSystemConfigPath
	c.First = "Test"
	return c
}

func (c *Config) Load(arguments []string) error {
	var path string
	f := flag.NewFlagSet("job", -1)
	f.SetOutput(ioutil.Discard)
	f.StringVar(&path, "config", "", "path to config file")
	f.Parse(arguments)
	if path != "" {
		if err := c.LoadFile(path); err != nil {
			return err
		}
	}
	if err := c.LoadEnv(); err != nil {
		return err
	}
	if err := c.LoadFlags(arguments); err != nil {
		return err
	}
	return nil
}

// Loads configuration from a file.
func (c *Config) LoadFile(path string) error {
	_, err := toml.DecodeFile(path, &c)
	return err
}

// LoadEnv loads the configuration via environment variables.
func (c *Config) LoadEnv() error {
	if err := c.loadEnv(c); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadEnv(target interface{}) error {
	value := reflect.Indirect(reflect.ValueOf(target))
	typ := value.Type()
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Retrieve environment variable.
		v := strings.TrimSpace(os.Getenv(field.Tag.Get("env")))
		if v == "" {
			continue
		}

		// Set the appropriate type.
		switch field.Type.Kind() {
		case reflect.Bool:
			value.Field(i).SetBool(v != "0" && v != "false")
		case reflect.Int:
			newValue, err := strconv.ParseInt(v, 10, 0)
			if err != nil {
				return fmt.Errorf("Parse error: %s: %s", field.Tag.Get("env"), err)
			}
			value.Field(i).SetInt(newValue)
		case reflect.String:
			value.Field(i).SetString(v)
		}
	}
	return nil
}

func (c *Config) LoadFlags(arguments []string) error {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.SetOutput(ioutil.Discard)

	f.StringVar(&c.First, "cf", c.First, "(deprecated)")
	if err := f.Parse(arguments); err != nil {
		return err
	}
	return nil
}
