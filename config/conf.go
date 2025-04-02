package config

import (
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(NewJWTConfig, NewMiddlewareConfig)

type JWTConfig struct {
	SecretKey string `yaml:"secretKey"` //秘钥
	Timeout   int    `yaml:"timeout"`   //过期时间
}

func NewJWTConfig() JWTConfig {
	return JWTConfig{
		SecretKey: viper.GetString("jwt.secretKey"),
		Timeout:   viper.GetInt("jwt.timeout"),
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
