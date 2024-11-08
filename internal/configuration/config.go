package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	LogLevel int    `envconfig:"ODJ_DEP_CORE_API_STATICS_LOG_LEVEL" default:"8" required:"true"`
	Port     int    `envconfig:"ODJ_DEP_CORE_API_STATICS_DB_PORT" default:"3306" required:"true"`
	Adress   string `envconfig:"ODJ_DEP_CORE_API_STATICS_ADRESS" default:":8080"`
	Host     string `envconfig:"ODJ_DEP_CORE_API_STATICS_HOST" default:"localhost"`
	Db       string `envconfig:"ODJ_DEP_CORE_API_STATICS_DB"`
}

func Get() (*Config, error) {
	var config Config

	err := envconfig.Process("", &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
