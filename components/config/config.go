package config

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

//@see https://github.com/spf13/viper
//
// usage:
//	if use file config  like this:
//	cf, err :=  NewConfig("../../test/configs/prod/config.toml")
//	if err != nil {
//
//	}
//	cf.Handler.Get("log")

const (
	//配置类型
	TypeFile = "file"
	TypeData = "data"
)

//图片配置信息
type Config struct {
	Setting Setting
	//句柄
	Handler *viper.Viper
}

//配置设置
type Setting struct {
	//file string
	Type string
	//文件类型 json toml yaml, hcl,
	Format string
	//文件地址
	Filename string
}

//解析配置文件
// t  JSON, TOML, YAML, HCL, INI 文件类型
// filename 文件
func NewConfig(filename string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filename)
	err := v.ReadInConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("new config error! filename:%s error:%s", filename, err.Error()))
	}

	c := new(Config)
	setting := Setting{}
	//读取文件拓展名
	if ext := filepath.Ext(filename); len(ext) > 0 {
		setting.Format = ext[1:]
	}
	setting.Filename = filename
	setting.Type = TypeFile
	c.Setting = setting
	c.Handler = v
	return c, nil
}

//解析配置数据
func NewConfigData(format string, data []byte) (*Config, error) {
	v := viper.New()
	v.SetConfigType(format)

	if err := v.ReadConfig(bytes.NewBuffer(data)); err != nil {
		return nil, errors.New(fmt.Sprintf("new config data fail err: %s", err.Error()))
	}
	c := new(Config)
	setting := Setting{}
	setting.Type = TypeData
	setting.Format = format
	c.Setting = setting
	c.Handler = v

	return c, nil
}

//监听配置变化从新读取
func (c *Config) WatchReConfig() {
	c.Handler.WatchConfig()
	c.Handler.OnConfigChange(func(in fsnotify.Event) {
		if in.Op.String() == "CREATE" {
			_ = c.Handler.ReadInConfig()
		}
	})
}
