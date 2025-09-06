package config

import (
	_ "embed"
	"strings"

	"github.com/spf13/viper"

	"github.com/ptonlix/whalehire/backend/pkg/logger"
)

type Config struct {
	Debug bool `mapstructure:"debug"`

	ReadOnly bool `mapstructure:"read_only"`

	Logger *logger.Config `mapstructure:"logger"`

	Server struct {
		Addr string `mapstructure:"addr"`
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`

	Admin struct {
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		Limit    int    `mapstructure:"limit"`
	} `mapstructure:"admin"`

	Session struct {
		ExpireDay int `mapstructure:"expire_day"`
	} `mapstructure:"session"`

	Database struct {
		Master          string `mapstructure:"master"`
		Slave           string `mapstructure:"slave"`
		MaxOpenConns    int    `mapstructure:"max_open_conns"`
		MaxIdleConns    int    `mapstructure:"max_idle_conns"`
		ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	} `mapstructure:"database"`

	Redis struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Pass     string `mapstructure:"pass"`
		DB       int    `mapstructure:"db"`
		IdleConn int    `mapstructure:"idle_conn"`
	} `mapstructure:"redis"`

	Embedding struct {
		ModelName   string `mapstructure:"model_name"`
		APIEndpoint string `mapstructure:"api_endpoint"`
		APIKey      string `mapstructure:"api_key"`
	} `mapstructure:"embedding"`

	DataReport struct {
		Key string `mapstructure:"key"`
	} `mapstructure:"data_report"`
}

func Init() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvPrefix("WHALEHIRE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	v.SetDefault("debug", false)
	v.SetDefault("read_only", false)
	v.SetDefault("logger.level", "info")
	v.SetDefault("server.addr", ":8888")
	v.SetDefault("server.port", "")
	v.SetDefault("admin.user", "admin")
	v.SetDefault("admin.password", "")
	v.SetDefault("admin.limit", 100)
	v.SetDefault("session.expire_day", 30)
	v.SetDefault("database.master", "")
	v.SetDefault("database.slave", "")
	v.SetDefault("database.max_open_conns", 50)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", 30)
	v.SetDefault("redis.host", "whalehire-redis")
	v.SetDefault("redis.port", "6379")
	v.SetDefault("redis.pass", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.idle_conn", 20)
	v.SetDefault("data_report.key", "")
	v.SetDefault("embedding.model_name", "qwen3-embedding-0.6b")
	v.SetDefault("embedding.api_endpoint", "https://aiapi.chaitin.net/v1/embeddings")
	v.SetDefault("embedding.api_key", "")

	c := Config{}
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
