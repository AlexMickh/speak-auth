package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string `env:"ENV" env-default:"prod"`
	Port            int    `env:"PORT" env-default:"50061"`
	UserServiceAddr string `env:"USER_SERVICE_ADDR" env-required:"true"`
	DB              DBConfig
	Jwt             JwtConfig
	Mail            MailConfig
}

type DBConfig struct {
	Host           string `env:"DB_HOST" env-default:"localhost"`
	Port           int    `env:"DB_PORT" env-default:"5432"`
	User           string `env:"DB_USER" env-default:"postgres"`
	Password       string `env:"DB_PASSWORD" env-required:"true"`
	Name           string `env:"DB_NAME" env-default:"auth"`
	MaxPools       int    `env:"DB_MAX_POOS" env-default:"5"`
	MigrationsPath string `env:"MIGRATIONS_PATH" env-default:"./migrations"`
}

type JwtConfig struct {
	Secret     string        `env:"JWT_SECRET" env-required:"true"`
	AccessTtl  time.Duration `env:"JWT_ACCESS_TTL" env-required:"true"`
	RefreshTtl time.Duration `env:"JWT_REFRESH_TTL" env-required:"true"`
}

type MailConfig struct {
	Host     string `env:"MAIL_HOST" env-required:"true"`
	Port     int    `env:"MAIL_PORT" env-required:"true"`
	FromAddr string `env:"MAIL_FROM_ADDR" env-required:"true"`
	Password string `env:"MAIL_PASSWORD" env-required:"true"`
}

func MustLoad() *Config {
	path := fetchPath()
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	cfg := &Config{}

	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}

func fetchPath() string {
	var path string
	flag.StringVar(&path, "config", "", "path to the config")
	flag.Parse()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}

	return path
}
