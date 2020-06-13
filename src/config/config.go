package config

import (
	"github.com/astaxie/beego/logs"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ServCfg struct {
	Listen  *ListenCfg  `yaml:"listen"`
	Connect *ConnectCfg `yaml:"connect"`
	FileUrl string      `yaml:"file_url"`
}

type ListenCfg struct {
	Stat    string `yaml:"stat"`
	Plat    string `yaml:"plat"`
	PlatTLS string `yaml:"plat_tls"`
}

type ConnectCfg struct {
	Mysql string `yaml:"mysql"`
	Plat  string `yaml:"plat"`
}

func Init() {
	buf, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		logs.Error("Config read error:", err)
		return
	}
	err = yaml.Unmarshal(buf, &g_servCfg)
	if err != nil {
		logs.Error("Config parse error:", err)
		return
	}
}

func Get() *ServCfg {
	return g_servCfg
}

var g_servCfg *ServCfg
