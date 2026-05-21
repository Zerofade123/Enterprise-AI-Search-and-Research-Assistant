package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/Zerofade123/Enterprise-AI-Search-and-Research-Assistant/backend/internal/platform/validation"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load(serviceName string) (*AppConfig, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	applyDefaults(v, serviceName)

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config unmarshal: %w", err)
	}

	if err := validation.ValidateStruct(cfg); err != nil {
		return nil, err
	}

	if cfg.JWT.AccessTokenTTL <= 0 || cfg.JWT.RefreshTokenTTL <= 0 || cfg.JWT.MfaTokenTTL <= 0 {
		return nil, fmt.Errorf("invalid jwt ttl configuration")
	}
	if cfg.Postgres.MaxIdleConns > cfg.Postgres.MaxOpenConns {
		return nil, fmt.Errorf("postgres max idle conns cannot exceed max open conns")
	}

	return cfg, nil
}

func applyDefaults(v *viper.Viper, serviceName string) {
	v.SetDefault("ENV", "development")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("SERVICE_NAME", serviceName)

	v.SetDefault("HTTP.PORT", 8080)
	v.SetDefault("HTTP.READ_TIMEOUT", 10*time.Second)
	v.SetDefault("HTTP.WRITE_TIMEOUT", 10*time.Second)
	v.SetDefault("HTTP.IDLE_TIMEOUT", 60*time.Second)
	v.SetDefault("HTTP.SHUTDOWN_GRACE_PERIOD", 15*time.Second)

	v.SetDefault("POSTGRES.HOST", "localhost")
	v.SetDefault("POSTGRES.PORT", 5432)
	v.SetDefault("POSTGRES.USER", "enterprise")
	v.SetDefault("POSTGRES.PASSWORD", "enterprise")
	v.SetDefault("POSTGRES.DB", "enterprise_ai")
	v.SetDefault("POSTGRES.SSL_MODE", "disable")
	v.SetDefault("POSTGRES.MAX_OPEN_CONNS", 25)
	v.SetDefault("POSTGRES.MAX_IDLE_CONNS", 5)
	v.SetDefault("POSTGRES.CONN_MAX_LIFETIME", 30*time.Minute)
	v.SetDefault("POSTGRES.CONN_MAX_IDLE_TIME", 5*time.Minute)

	v.SetDefault("REDIS.HOST", "localhost")
	v.SetDefault("REDIS.PORT", 6379)
	v.SetDefault("REDIS.PASSWORD", "")
	v.SetDefault("REDIS.DB", 0)

	v.SetDefault("S3.ENDPOINT", "http://localhost:9000")
	v.SetDefault("S3.ACCESS_KEY", "minio")
	v.SetDefault("S3.SECRET_KEY", "minio123")
	v.SetDefault("S3.BUCKET", "enterprise-docs")
	v.SetDefault("S3.REGION", "us-east-1")
	v.SetDefault("S3.USE_SSL", false)

	v.SetDefault("MILVUS.HOST", "localhost")
	v.SetDefault("MILVUS.PORT", 19530)

	v.SetDefault("JWT.SIGNING_KEY", "")
	v.SetDefault("JWT.ISSUER", "enterprise-ai-auth")
	v.SetDefault("JWT.KEY_ID", "default")
	v.SetDefault("JWT.ACCESS_TOKEN_TTL", 1*time.Hour)
	v.SetDefault("JWT.REFRESH_TOKEN_TTL", 7*24*time.Hour)
	v.SetDefault("JWT.MFA_TOKEN_TTL", 5*time.Minute)
}
