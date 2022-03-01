package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Name string
}

// 读取配置，默认配置为./conf/default.conf
func (c *Config) InitConfig() error {
	if c.Name != "" {
		viper.SetConfigFile(c.Name)
	} else {
		viper.AddConfigPath("conf")
		viper.SetConfigName("default.conf")
	}
	viper.SetConfigType("yaml")

	return viper.ReadInConfig()
}



