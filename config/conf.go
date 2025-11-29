package config

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(NewJWTConfig, NewMiddlewareConfig, NewLogConfig)

type JWTConfig struct {
	JwtKey  string `yaml:"jwtKey"` //秘钥
	EncKey  string `yaml:"encKey"`
	Timeout int    `yaml:"timeout"` //过期时间
}

func NewJWTConfig() JWTConfig {
	return JWTConfig{
		JwtKey:  viper.GetString("jwt.jwtKey"),
		EncKey:  viper.GetString("jwt.encKey"),
		Timeout: viper.GetInt("jwt.timeout"),
	}
}

type MiddlewareConfig struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

func NewMiddlewareConfig() *MiddlewareConfig {
	return &MiddlewareConfig{
		AllowedOrigins: viper.GetStringSlice("middleware.allowedOrigins"),
	}
}

type LogConfig struct {
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
}

func NewLogConfig() *LogConfig {
	return &LogConfig{
		File:       viper.GetString("log.file"),
		MaxSize:    viper.GetInt("log.maxSize"),
		MaxBackups: viper.GetInt("log.maxBackups"),
		MaxAge:     viper.GetInt("log.maxAge"),
		Compress:   viper.GetBool("log.compress"),
	}
}
