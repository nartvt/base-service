package config //nolint:revive

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Profile    string           `mapstructure:"profile" json:"profile,omitempty"`
	Server     ServerConfig     `mapstructure:"server" json:"server,omitempty"`
	Log        LogConfig        `mapstructure:"log" json:"log,omitempty"`
	Database   DatabaseConfig   `mapstructure:"database" json:"database,omitempty"`
	Redis      RedisConfig      `mapstructure:"redis" json:"redis,omitempty"`
	Middleware MiddlewareConfig `mapstructure:"middleware" json:"middleware,omitempty"`
}

type LogConfig struct {
	LogLevel   slog.Level `mapstructure:"level"`
	JSONOutput bool       `mapstructure:"jsonOutput" json:"json_output,omitempty"`
	AddSource  bool       `mapstructure:"addSource" json:"add_source,omitempty"`
}

type ServerConfig struct {
	Http ServerInfo `mapstructure:"http" json:"http,omitempty"`
	Grpc ServerInfo `mapstructure:"grpc" json:"grpc,omitempty"`
}

type ServerInfo struct {
	AppName        string `mapstructure:"appName" json:"app_name,omitempty"`
	Host           string `mapstructure:"host" json:"host,omitempty"`
	Port           int    `mapstructure:"port" json:"port,omitempty"`
	EnableTLS      bool   `mapstructure:"enableTLS" json:"enable_tls,omitempty"`
	ReadTimeout    int    `mapstructure:"readTimeout" json:"read_timeout,omitempty"`
	WriteTimeout   int    `mapstructure:"writeTimeout" json:"write_timeout,omitempty"`
	ConnectTimeOut int    `mapstructure:"connectTimeOut" json:"connect_time_out,omitempty"`
}

type DatabaseConfig struct {
	DriverName         string        `mapstructure:"driverName" json:"driver_name,omitempty"`
	Host               string        `mapstructure:"host" json:"host,omitempty"`
	Port               int           `mapstructure:"port" json:"port,omitempty"`
	UserName           string        `mapstructure:"userName" json:"user_name,omitempty"`
	Password           string        `mapstructure:"password" json:"password,omitempty"`
	DBName             string        `mapstructure:"dbName" json:"db_name,omitempty"`
	SSLMode            string        `mapstructure:"sslMode" json:"ssl_mode,omitempty"`
	MaxOpenConnections int32         `mapstructure:"maxOpenConnections" json:"max_open_connections,omitempty"`
	MaxIdleConnections int32         `mapstructure:"maxIdleConnections" json:"max_idle_connections,omitempty"`
	MaxConnLifetime    time.Duration `mapstructure:"maxConnLifetime" json:"max_conn_lifetime,omitempty"`
	MaxConnIdleTime    time.Duration `mapstructure:"maxConnIdleTime" json:"max_conn_idle_time,omitempty"`
}

type MiddlewareConfig struct {
	Token     TokenConfig     `mapstructure:"token" json:"token,omitempty"`
	CORS      CORSConfig      `mapstructure:"cors" json:"cors,omitempty"`
	RateLimit RateLimitConfig `mapstructure:"rateLimit" json:"rate_limit,omitempty"`
}

type TokenConfig struct {
	PasswordSalt       string        `mapstructure:"passwordSalt" json:"password_salt,omitempty"`
	AccessTokenSecret  string        `mapstructure:"accessTokenSecret" json:"access_token_secret,omitempty"`
	AccessTokenExp     time.Duration `mapstructure:"accessTokenExp" json:"access_token_exp,omitempty"`
	RefreshTokenSecret string        `mapstructure:"refreshTokenSecret" json:"refresh_token_secret,omitempty"`
	RefreshTokenExp    time.Duration `mapstructure:"refreshTokenExp" json:"refresh_token_exp,omitempty"`
}

type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowedOrigins" json:"allowed_origins,omitempty"`
	AllowedMethods   []string `mapstructure:"allowedMethods" json:"allowed_methods,omitempty"`
	AllowedHeaders   []string `mapstructure:"allowedHeaders" json:"allowed_headers,omitempty"`
	ExposedHeaders   []string `mapstructure:"exposedHeaders" json:"exposed_headers,omitempty"`
	AllowCredentials bool     `mapstructure:"allowCredentials" json:"allow_credentials,omitempty"`
	MaxAge           int      `mapstructure:"maxAge" json:"max_age,omitempty"`
}

type RateLimitConfig struct {
	Enabled        bool          `mapstructure:"enabled" json:"enabled,omitempty"`
	Max            int           `mapstructure:"max" json:"max,omitempty"`                         // Max requests per window
	Expiration     time.Duration `mapstructure:"expiration" json:"expiration,omitempty"`           // Time window (e.g., 1 minute)
	SkipFailedReq  bool          `mapstructure:"skipFailedReq" json:"skip_failed_req,omitempty"`   // Skip failed requests
	SkipSuccessReq bool          `mapstructure:"skipSuccessReq" json:"skip_success_req,omitempty"` // Skip successful requests
	LimitReached   string        `mapstructure:"limitReached" json:"limit_reached,omitempty"`      // Custom message when limit is reached

	// Auth-specific rate limits (stricter)
	AuthEnabled    bool          `mapstructure:"authEnabled" json:"auth_enabled,omitempty"`
	AuthMax        int           `mapstructure:"authMax" json:"auth_max,omitempty"`
	AuthExpiration time.Duration `mapstructure:"authExpiration" json:"auth_expiration,omitempty"`
}

type RedisConfig struct {
	Host         string        `mapstructure:"host" json:"host,omitempty"`
	Port         int           `mapstructure:"port" json:"port,omitempty"`
	Password     string        `mapstructure:"password" json:"password,omitempty"`
	DB           int           `mapstructure:"db" json:"db,omitempty"`
	MaxIdle      int           `mapstructure:"maxIdle" json:"max_idle,omitempty"`
	DialTimeout  time.Duration `mapstructure:"dialTimeout" json:"dial_timeout,omitempty"`
	ReadTimeout  time.Duration `mapstructure:"readTimeout" json:"read_timeout,omitempty"`
	WriteTimeout time.Duration `mapstructure:"writeTimeout" json:"write_timeout,omitempty"`
}

func loadConfig(configfile string) (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigFile(configfile)
	v.SetConfigType("yaml")

	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	v.SetEnvKeyReplacer(strings.NewReplacer("_", "."))

	if err := v.ReadInConfig(); err != nil {
		log.Println("config file not found. Using exists environment variables")
		return nil, err
	}
	overrideconfig(v)
	return v, nil
}

func overrideconfig(v *viper.Viper) {
	for _, key := range v.AllKeys() {
		envKey := "APP_" + strings.ReplaceAll(strings.ToUpper(key), ".", "_")
		envValue := os.Getenv(envKey)
		if envValue != "" {
			v.Set(key, envValue)
		}

	}
}

func LoadConfig(pathToFile string, env string, config any) error {
	slog.Info(fmt.Sprintf("Load config %s", env))
	configPath := pathToFile + "/" + "application"
	if len(env) > 0 && env != "local" {
		configPath = configPath + "-" + env
	}

	pwd, err := os.Getwd()
	if err == nil && len(pwd) > 0 {
		configPath = pwd + "/" + configPath
	}

	confFile := fmt.Sprintf("%s.yaml", configPath)
	v, err := loadConfig(confFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := v.Unmarshal(&config); err != nil {
		return fmt.Errorf("unable to decode into struct, %v", err)
	}
	return nil
}

func (r *DatabaseConfig) BuildConnectionStringPostgres() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		r.Host, r.Port, r.UserName, r.Password, r.DBName, func() string {
			if len(r.SSLMode) == 0 {
				return "disable"
			}
			return r.SSLMode
		}(),
	)
}

func (r *RedisConfig) BuildRedisConnectionString() string {
	return fmt.Sprintf("redis://%s:%s@%s:%d", "", "", r.Host, r.Port)
}

func (c *CORSConfig) GetOriginsString() string {
	if len(c.AllowedOrigins) == 0 {
		return "*"
	}
	return strings.Join(c.AllowedOrigins, ", ")
}

func (c *CORSConfig) GetMethodsString() string {
	if len(c.AllowedMethods) == 0 {
		return "GET, POST, PUT, DELETE, OPTIONS"
	}
	return strings.Join(c.AllowedMethods, ", ")
}

func (c *CORSConfig) GetHeadersString() string {
	if len(c.AllowedHeaders) == 0 {
		return "Origin, Content-Type, Accept, Authorization"
	}
	return strings.Join(c.AllowedHeaders, ", ")
}

func (c *CORSConfig) GetExposedHeadersString() string {
	return strings.Join(c.ExposedHeaders, ", ")
}
