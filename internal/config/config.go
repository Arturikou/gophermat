package config

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env           string              `yaml:"env" env-default:"local"`
	LogLevel      string              `yaml:"log_level" env-default:"info"`
	HTTPServer    HTTPServerConfig    `yaml:"http_server"`
	DB            DBConfig            `yaml:"db"`
	Auth          AuthConfig          `yaml:"auth"`
	AccrualSystem AccrualSystemConfig `yaml:"accrual_system"`
	AccrualWorker AccrualWorkerConfig `yaml:"accrual_worker"`
}

type AuthConfig struct {
	SecretKey string `yaml:"secret_key" env:"AUTH_SECRET_KEY"`
}

type DBConfig struct {
	DSN             string        `yaml:"dsn"               env:"DATABASE_URI"`
	MaxOpenConns    int32         `yaml:"max_open_conns"    env-default:"25"`
	MaxIdleConns    int32         `yaml:"max_idle_conns"    env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"1m"`
}

type HTTPServerConfig struct {
	Address      string        `yaml:"address" env:"RUN_ADDRESS" env-default:"localhost:8080"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"2s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"2s"`
}

type AccrualSystemConfig struct {
	Address string        `yaml:"address" env:"ACCRUAL_SYSTEM_ADDRESS"`
	Timeout time.Duration `yaml:"timeout" env-default:"5s"`
}

type AccrualWorkerConfig struct {
	WorkerCount    int `yaml:"worker_count" env-default:"5"`
	MaxConcurrency int `yaml:"max_concurrency"    env-default:"10"`
}

func MustLoad() *Config {
	serverAddress := flag.String("a", "", "server address")
	dbAddress := flag.String("d", "", "db address")
	accrualSystemAddress := flag.String("r", "", "accrual system address")
	flag.Parse()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file not found: %s", configPath)
	}

	cfg := Config{}
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	if *serverAddress != "" {
		cfg.HTTPServer.Address = *serverAddress
	}

	if *dbAddress != "" {
		cfg.DB.DSN = *dbAddress
	}

	if *accrualSystemAddress != "" {
		cfg.AccrualSystem.Address = *accrualSystemAddress
	}

	return &cfg
}
