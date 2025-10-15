package config

import (
	_ "embed"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/chaitin/WhaleHire/backend/domain"
	"github.com/chaitin/WhaleHire/backend/pkg/logger"
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
		Vector   struct {
			Enabled   bool `mapstructure:"enabled"`
			Dimension int  `mapstructure:"dimension"`
		} `mapstructure:"vector"`
	} `mapstructure:"redis"`

	GeneralAgent struct {
		LLM struct {
			ModelName string `mapstructure:"model_name"`
			BaseURL   string `mapstructure:"base_url"`
			APIKey    string `mapstructure:"api_key"`
		} `mapstructure:"llm"`
	} `mapstructure:"general_agent"`

	Embedding struct {
		ModelName   string `mapstructure:"model_name"`
		APIEndpoint string `mapstructure:"api_endpoint"`
		APIKey      string `mapstructure:"api_key"`
		Dimension   int    `mapstructure:"dimension"`
	} `mapstructure:"embedding"`

	Retriever struct {
		TopK              int     `mapstructure:"top_k"`
		DistanceThreshold float64 `mapstructure:"distance_threshold"`
	} `mapstructure:"retriever"`

	DataReport struct {
		Key string `mapstructure:"key"`
	} `mapstructure:"data_report"`

	S3 struct {
		Endpoint   string `mapstructure:"endpoint"`
		AccessKey  string `mapstructure:"access_key"`
		SecretKey  string `mapstructure:"secret_key"`
		BucketName string `mapstructure:"bucket_name"`
	}

	// 文件存储配置
	FileStorage struct {
		LocalPath    string   `mapstructure:"local_path"`
		MaxFileSize  int64    `mapstructure:"max_file_size"`
		AllowedTypes []string `mapstructure:"allowed_types"`
	} `mapstructure:"file_storage"`

	// DocumentParser 文档解析配置
	DocumentParser struct {
		APIKey     string `mapstructure:"api_key" json:"api_key"`
		BaseURL    string `mapstructure:"base_url" json:"base_url"`
		Timeout    int    `mapstructure:"timeout" json:"timeout"`
		MaxRetries int    `mapstructure:"max_retries" json:"max_retries"`
	} `mapstructure:"document_parser" json:"document_parser"`

	// Langsmith 配置
	Langsmith struct {
		APIKey string `mapstructure:"api_key" json:"api_key"`
	} `mapstructure:"langsmith" json:"langsmith"`
}

func (c *Config) GetBaseURL(req *http.Request, settings *domain.Setting) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	if proto := req.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}

	if settings != nil && settings.BaseURL != "" {
		baseurl := settings.BaseURL
		if !strings.HasPrefix(baseurl, "http") {
			baseurl = fmt.Sprintf("%s://%s", scheme, baseurl)
		}
		return strings.TrimSuffix(baseurl, "/")
	}

	if port := req.Header.Get("X-Forwarded-Port"); port != "" && port != "80" && port != "443" {
		c.Server.Port = port
	}
	baseurl := fmt.Sprintf("%s://%s", scheme, req.Host)
	if c.Server.Port != "" {
		baseurl = fmt.Sprintf("%s:%s", baseurl, c.Server.Port)
	}
	return baseurl
}

func Init() (*Config, error) {
	envPaths := []string{
		".env",
		filepath.Join("..", ".env"),
		filepath.Join("..", "..", ".env"),
	}
	for _, path := range envPaths {
		if err := godotenv.Load(path); err == nil {
			break
		}
	}
	v := viper.New()
	v.SetEnvPrefix("WHALEHIRE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	// 设置默认值
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
	v.SetDefault("redis.vector.enabled", true)
	v.SetDefault("redis.vector.dimension", 4096)
	v.SetDefault("data_report.key", "")
	v.SetDefault("general_agent.llm.model_name", "deepseek-chat")
	v.SetDefault("general_agent.llm.base_url", "https://api.deepseek.com/v1")
	v.SetDefault("general_agent.llm.api_key", "")

	v.SetDefault("embedding.model_name", "bge-m3")
	v.SetDefault("embedding.api_endpoint", "https://model-square.app.baizhi.cloud/v1")
	v.SetDefault("embedding.api_key", "")
	v.SetDefault("embedding.dimension", 4096)

	v.SetDefault("retriever.top_k", 3)
	v.SetDefault("retriever.distance_threshold", 0.8)
	v.SetDefault("s3.endpoint", "whalehire-minio:9000")
	v.SetDefault("s3.access_key", "s3whale-hire")
	v.SetDefault("s3.secret_key", "")
	v.SetDefault("s3.bucket_name", "static-file")

	// 文件存储默认配置
	v.SetDefault("file_storage.local_path", "./uploads")
	v.SetDefault("file_storage.max_file_size", 10485760) // 10MB
	v.SetDefault("file_storage.allowed_types", []string{".pdf", ".docx", ".doc"})

	// 文件内容提取默认配置
	v.SetDefault("document_parser.api_key", "")
	v.SetDefault("document_parser.base_url", "https://api.moonshot.cn/v1")
	v.SetDefault("document_parser.timeout", 30)
	v.SetDefault("document_parser.max_retries", 3)

	// Langsmith 默认配置
	v.SetDefault("langsmith.api_key", "")

	// 打印从环境变量中读取的所有配置值
	fmt.Println("从环境变量读取的配置值:")
	fmt.Println("Database Master:", v.GetString("database.master"))
	fmt.Println("Database Slave:", v.GetString("database.slave"))
	fmt.Println("Embedding API Key:", v.GetString("embedding.api_key"))
	fmt.Println("Document Parser API Key:", v.GetString("document_parser.api_key"))
	fmt.Println("Langsmith API Key:", v.GetString("langsmith.api_key"))

	c := Config{}
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
