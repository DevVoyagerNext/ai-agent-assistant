package config

import (
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Admin   AdminConfig   `yaml:"admin"`
	Mysql   MysqlConfig   `yaml:"mysql"`
	Redis   RedisConfig   `yaml:"redis"`
	Session SessionConfig `yaml:"session"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type AdminConfig struct {
	InviteCode string `yaml:"invite-code"`
}

type MysqlConfig struct {
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Config       string `yaml:"config"`
	DbName       string `yaml:"db-name"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	MaxIdleConns int    `yaml:"max-idle-conns"`
	MaxOpenConns int    `yaml:"max-open-conns"`
}

type RedisConfig struct {
	DB       int    `yaml:"db"`
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
}

type SessionConfig struct {
	ExpiresTime int64 `yaml:"expires-time"`
}

func Load() Config {
	cfg := Config{
		Server: ServerConfig{Port: "8081"},
		Admin:  AdminConfig{InviteCode: "admin2026"},
		Mysql: MysqlConfig{
			Host:         "127.0.0.1",
			Port:         "3306",
			Config:       "charset=utf8mb4&parseTime=True&loc=Local",
			DbName:       "ai_study_assistant",
			Username:     "root",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		},
		Redis: RedisConfig{
			DB:   0,
			Addr: "127.0.0.1:6379",
		},
		Session: SessionConfig{
			ExpiresTime: 604800,
		},
	}

	if path := os.Getenv("ADMIN_CONFIG"); path != "" {
		_ = mergeConfig(path, &cfg)
	} else {
		for _, path := range []string{
			"config.yaml",
			filepath.Join("backend", "settings.yaml"),
			filepath.Join("..", "..", "backend", "settings.yaml"),
			filepath.Join("..", "backend", "settings.yaml"),
		} {
			if _, err := os.Stat(path); err == nil {
				_ = mergeConfig(path, &cfg)
				break
			}
		}
	}

	applyEnv(&cfg)
	return cfg
}

func mergeConfig(path string, cfg *Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, cfg)
}

func applyEnv(cfg *Config) {
	if value := os.Getenv("ADMIN_PORT"); value != "" {
		cfg.Server.Port = value
	}
	if value := os.Getenv("ADMIN_INVITE_CODE"); value != "" {
		cfg.Admin.InviteCode = value
	}
	if value := os.Getenv("MYSQL_HOST"); value != "" {
		cfg.Mysql.Host = value
	}
	if value := os.Getenv("MYSQL_PORT"); value != "" {
		cfg.Mysql.Port = value
	}
	if value := os.Getenv("MYSQL_DATABASE"); value != "" {
		cfg.Mysql.DbName = value
	}
	if value := os.Getenv("MYSQL_USERNAME"); value != "" {
		cfg.Mysql.Username = value
	}
	if value := os.Getenv("MYSQL_PASSWORD"); value != "" {
		cfg.Mysql.Password = value
	}
	if value := os.Getenv("REDIS_ADDR"); value != "" {
		cfg.Redis.Addr = value
	}
	if value := os.Getenv("REDIS_PASSWORD"); value != "" {
		cfg.Redis.Password = value
	}
	if value := os.Getenv("REDIS_DB"); value != "" {
		if db, err := strconv.Atoi(value); err == nil && db >= 0 {
			cfg.Redis.DB = db
		}
	}
	if value := os.Getenv("ADMIN_SESSION_EXPIRES_TIME"); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
			cfg.Session.ExpiresTime = seconds
		}
	}
	if value := os.Getenv("SESSION_EXPIRES_TIME"); value != "" {
		if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds > 0 {
			cfg.Session.ExpiresTime = seconds
		}
	}
}

func (m MysqlConfig) DSN() string {
	if value := os.Getenv("ADMIN_DB_DSN"); value != "" {
		return value
	}
	if value := os.Getenv("MYSQL_DSN"); value != "" {
		return value
	}
	return m.Username + ":" + m.Password + "@tcp(" + m.Host + ":" + m.Port + ")/" + m.DbName + "?" + m.Config
}
