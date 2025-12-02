package config

import (
	"fmt"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(NewJWTConfig, NewMiddlewareConfig, NewLogConfig, NewLimiterConfig, NewRedisConfig)

type JWTConfig struct {
	JwtKey  string `yaml:"jwtKey"` //秘钥
	EncKey  string `yaml:"encKey"`
	Timeout int    `yaml:"timeout"` //过期时间
}

func NewJWTConfig() *JWTConfig {
	var cfg *JWTConfig
	err := viper.UnmarshalKey("jwt", &cfg)
	if err != nil {
		panic(fmt.Sprintf("无法解析 JWT 配置: %v", err))
	}

	return cfg
}

type MiddlewareConfig struct {
	AllowedOrigins []string `yaml:"allowedOrigins"`
}

func NewMiddlewareConfig() *MiddlewareConfig {
	var cfg *MiddlewareConfig
	err := viper.UnmarshalKey("middleware", &cfg)
	if err != nil {
		panic(fmt.Sprintf("无法解析中间件配置: %v", err))
	}

	return cfg
}

type LogConfig struct {
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
}

func NewLogConfig() *LogConfig {
	var cfg *LogConfig
	err := viper.UnmarshalKey("log", &cfg)
	if err != nil {
		panic(fmt.Sprintf("无法解析日志配置: %v", err))
	}

	return cfg
}

type LimiterConfig struct {
	Capacity     int `yaml:"capacity"`     //令牌桶容量
	FillInterval int `yaml:"fillInterval"` //放置令牌的时间间隔(每秒多少次)
	Quantum      int `yaml:"quantum"`      //每次放置的令牌数
}

func NewLimiterConfig() *LimiterConfig {
	var cfg *LimiterConfig
	err := viper.UnmarshalKey("limiter", &cfg)
	if err != nil {
		panic(fmt.Sprintf("无法解析限流器配置: %v", err))
	}

	return cfg
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func NewRedisConfig() *RedisConfig {
	var cfg *RedisConfig
	err := viper.UnmarshalKey("redis", &cfg)
	if err != nil {
		panic(fmt.Sprintf("无法解析 Redis 配置: %v", err))
	}

	return cfg
}
