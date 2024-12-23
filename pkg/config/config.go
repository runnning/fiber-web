package config

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	NSQ      NSQConfig      `mapstructure:"nsq"`
	App      AppConfig      `mapstructure:"app"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type DatabaseConfig struct {
	MultiDB   bool                `mapstructure:"multi_db"`  // 是否启用多库模式
	Databases map[string]DBConfig `mapstructure:"databases"` // 多库配置
	Default   DBConfig            `mapstructure:"default"`   // 单库配置
}

type DBConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	MultiInstance bool                           `mapstructure:"multi_instance"` // 是否启用多实例
	Instances     map[string]RedisInstanceConfig `mapstructure:"instances"`      // 多实例配置
	Default       RedisInstanceConfig            `mapstructure:"default"`        // 单实例配置
}

type RedisInstanceConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
}

type NSQConfig struct {
	NSQD struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"nsqd"`
	Lookupd struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"lookupd"`
}

type AppConfig struct {
	Env      string `yaml:"env"`
	Name     string `yaml:"name"`
	Language string `yaml:"language"`
}

type JWTConfig struct {
	SecretKey          string        `mapstructure:"secret_key"`
	AccessTokenExpiry  time.Duration `mapstructure:"access_token_expiry"`
	RefreshTokenExpiry time.Duration `mapstructure:"refresh_token_expiry"`
}

func Load() (*Config, error) {
	configName := os.Getenv("CONFIG_NAME")
	if configName == "" {
		configName = "config.local" // 默认使用本地配置
	}

	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")

	// 首先检查环境变量中是否指定了配置文件路径
	configPath := os.Getenv("CONFIG_PATH")
	if configPath != "" {
		// 如果环境变量指定了配置文件路径，就使用该路径
		viper.AddConfigPath(configPath)
	}

	// 获取可执行文件所在目录
	executable, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(executable)
		// 添加相对于可执行文件的配置路径
		viper.AddConfigPath(filepath.Join(execDir, "../config"))     // 回退到上级目录的config
		viper.AddConfigPath(filepath.Join(execDir, "../cmd/config")) // 回退到上级目录的cmd/config
	}

	// 添加默认的配置文件搜索路径
	viper.AddConfigPath("../config")     // 相对于当前目录的上级config目录
	viper.AddConfigPath("../cmd/config") // 对于当前目录的上级cmd/config目录
	viper.AddConfigPath("/app/config")   // Docker环境路径
	viper.AddConfigPath(".")             // 当前目录

	viper.AutomaticEnv()

	// 设置默认值保持不变
	viper.SetDefault("server.address", ":3000")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("app.env", "development")
	viper.SetDefault("app.name", "fiber-web")
	viper.SetDefault("redis.instances.default.host", "localhost")
	viper.SetDefault("redis.instances.default.port", 6379)
	viper.SetDefault("redis.instances.default.db", 0)
	viper.SetDefault("redis.instances.default.pool_size", 10)
	viper.SetDefault("redis.instances.default.min_idle_conns", 5)
	viper.SetDefault("redis.instances.default.max_retries", 3)
	viper.SetDefault("nsq.nsqd.host", "localhost")
	viper.SetDefault("nsq.nsqd.port", 4150)
	viper.SetDefault("nsq.lookupd.host", "localhost")
	viper.SetDefault("nsq.lookupd.port", 4161)
	viper.SetDefault("jwt.secret_key", "secret")
	viper.SetDefault("jwt.access_token_expiry", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_expiry", 7*24*time.Hour)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", time.Hour)

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("警告: 未找到配置文件，使用默认配置: %v", err)
		} else {
			return nil, err
		}
	} else {
		log.Printf("使用配置文件: %s", viper.ConfigFileUsed())
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	log.Printf("具体配置信息:%+v", config)

	return &config, nil
}
