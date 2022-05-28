package config

import (
	"fmt"
	"os"
	"testing"
)

func initConfiguration(t *testing.T) *Config {
	dir, _ := os.Getwd()
	var filename = dir + "/../../test/configs/config.toml"
	c, err := NewConfig(filename)
	if err != nil {
		t.Error("init configuration error err:", err)
		return nil
	}
	return c
}

func Test_GetLog(t *testing.T) {
	c := initConfiguration(t)

	c.Handler.Get("log")

	v := c.Handler.GetStringMap("log")
	t.Log(fmt.Sprintf("%v", v))

}
