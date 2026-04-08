package postgresql

import "time"

type Config struct {
	DSN             string        `yaml:"dsn"                env:"DATABASE_URI"`
	MaxOpenConns    int32         `yaml:"max_open_conns"     env-default:"25"`
	MaxIdleConns    int32         `yaml:"max_idle_conns"     env-default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"  env-default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"1m"`
}
