package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

// ServerConfig matches config/*.yml under server:.
type ServerConfig struct {
	Port int `mapstructure:"PORT"`
}

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	// Env is the application environment label from YAML (e.g. Dev, Qa, Prod).
	Env string `mapstructure:"env"`
}

var loaded *Config

// Load reads config/{APP_ENV}.yml (default qa) and unmarshals into Config.
// Subsequent calls to ServerPort and Env use this result.
func Load(ctx context.Context) (*Config, error) {
	_ = ctx

	env := getEnv()
	viper.Reset()
	viper.SetConfigName(env)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")
	viper.AddConfigPath("../config")
	viper.AddConfigPath("../../config")
	if taskRoot := os.Getenv("LAMBDA_TASK_ROOT"); taskRoot != "" {
		viper.AddConfigPath(filepath.Join(taskRoot, "config"))
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	// Allow PORT or SERVER_PORT to override server.PORT
	_ = viper.BindEnv("server.PORT", "PORT", "SERVER_PORT")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	loaded = &cfg
	return &cfg, nil
}

// Env returns the env label from YAML (e.g. Dev, Qa, Prod), or APP_ENV / "qa" if unset.
func Env() string {
	if loaded != nil && loaded.Env != "" {
		return loaded.Env
	}
	return getEnv()
}

// ServerPort returns the HTTP listen port from loaded config, or 8080 if unset.
func ServerPort() int {
	if loaded != nil && loaded.Server.Port > 0 {
		return loaded.Server.Port
	}
	return 8080
}

// ListenAddr returns the address string for local HTTP (e.g. ":8080").
// When not running in Lambda, YAML port 80 is mapped to 8080 to avoid bind errors / root.
func ListenAddr() string {
	port := ServerPort()
	if os.Getenv("AWS_LAMBDA_RUNTIME_API") == "" && port == 80 {
		port = 8080
	}
	return ":" + strconv.Itoa(port)
}

// ListenAddrFrom returns ":port" for the given config (useful if you keep a *Config locally).
func ListenAddrFrom(c *Config) string {
	if c == nil || c.Server.Port <= 0 {
		return ":8080"
	}
	return ":" + strconv.Itoa(c.Server.Port)
}

func GetConfig() *viper.Viper {
	return viper.GetViper()
}

func getEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "qa"
	}
	return env
}
