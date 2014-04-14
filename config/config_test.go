package config

import (
	"os"
	"testing"

	"github.com/BurntSushi/toml"
	. "github.com/smartystreets/goconvey/convey"
)

func TestConfigToml(t *testing.T) {
	content := `
		conf_first = "127.0.0.1:4002"
	`
	c := New()
	_, err := toml.Decode(content, &c)

	Convey("Toml can parse file right", t, func() {
		So(err, ShouldBeNil)
	})
	Convey("ShouldEqual", t, func() {
		So(c.First, ShouldEqual, "127.0.0.1:4002")
	})
}

func TestConfigEnv(t *testing.T) {
	os.Setenv("CONF_FIRST", "this.is.test")
	c := New()
	c.LoadEnv()

	Convey("Env can use", t, func() {
		So(c.First, ShouldEqual, "this.is.test")
	})
}
