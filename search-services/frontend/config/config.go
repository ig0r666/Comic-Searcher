package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel     string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPAddress  string `yaml:"address" env:"FRONTEND_ADDRESS" env-default:"localhost:84"`
	APIAddress   string `yaml:"api_address" env:"API_ADDRESS" env-default:"api:8080"`
	TemplatePath string `yaml:"template_path" env:"TEMPLATE_PATH" env-default:"/templates"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
