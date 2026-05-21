package config

import "time"

type AppConfig struct {
	Env         string         `mapstructure:"ENV" validate:"required,oneof=development staging production"`
	LogLevel    string         `mapstructure:"LOG_LEVEL" validate:"required,oneof=debug info warn error"`
	ServiceName string         `mapstructure:"SERVICE_NAME" validate:"required"`
	HTTP        HTTPConfig     `mapstructure:"HTTP" validate:"required"`
	Postgres    PostgresConfig `mapstructure:"POSTGRES" validate:"required"`
	Redis       RedisConfig    `mapstructure:"REDIS" validate:"required"`
	S3          S3Config       `mapstructure:"S3" validate:"required"`
	Milvus      MilvusConfig   `mapstructure:"MILVUS" validate:"required"`
	JWT         JWTConfig      `mapstructure:"JWT" validate:"required"`
}

type HTTPConfig struct {
	Port                int           `mapstructure:"PORT" validate:"required,min=1,max=65535"`
	ReadTimeout         time.Duration `mapstructure:"READ_TIMEOUT" validate:"required"`
	WriteTimeout        time.Duration `mapstructure:"WRITE_TIMEOUT" validate:"required"`
	IdleTimeout         time.Duration `mapstructure:"IDLE_TIMEOUT" validate:"required"`
	ShutdownGracePeriod time.Duration `mapstructure:"SHUTDOWN_GRACE_PERIOD" validate:"required"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"HOST" validate:"required"`
	Port            int           `mapstructure:"PORT" validate:"required,min=1,max=65535"`
	User            string        `mapstructure:"USER" validate:"required"`
	Password        string        `mapstructure:"PASSWORD" validate:"required"`
	DBName          string        `mapstructure:"DB" validate:"required"`
	SSLMode         string        `mapstructure:"SSL_MODE" validate:"required,oneof=disable require verify-ca verify-full"`
	MaxOpenConns    int           `mapstructure:"MAX_OPEN_CONNS" validate:"required,min=1"`
	MaxIdleConns    int           `mapstructure:"MAX_IDLE_CONNS" validate:"required,min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"CONN_MAX_LIFETIME" validate:"required"`
	ConnMaxIdleTime time.Duration `mapstructure:"CONN_MAX_IDLE_TIME" validate:"required"`
}

type RedisConfig struct {
	Host     string `mapstructure:"HOST" validate:"required"`
	Port     int    `mapstructure:"PORT" validate:"required,min=1,max=65535"`
	Password string `mapstructure:"PASSWORD"`
	DB       int    `mapstructure:"DB" validate:"min=0"`
}

type S3Config struct {
	Endpoint  string `mapstructure:"ENDPOINT" validate:"required"`
	AccessKey string `mapstructure:"ACCESS_KEY" validate:"required"`
	SecretKey string `mapstructure:"SECRET_KEY" validate:"required"`
	Bucket    string `mapstructure:"BUCKET" validate:"required"`
	Region    string `mapstructure:"REGION" validate:"required"`
	UseSSL    bool   `mapstructure:"USE_SSL"`
}

type MilvusConfig struct {
	Host string `mapstructure:"HOST" validate:"required"`
	Port int    `mapstructure:"PORT" validate:"required,min=1,max=65535"`
}

type JWTConfig struct {
	SigningKey      string        `mapstructure:"SIGNING_KEY" validate:"required"`
	Issuer          string        `mapstructure:"ISSUER" validate:"required"`
	KeyID           string        `mapstructure:"KEY_ID" validate:"required"`
	AccessTokenTTL  time.Duration `mapstructure:"ACCESS_TOKEN_TTL" validate:"required"`
	RefreshTokenTTL time.Duration `mapstructure:"REFRESH_TOKEN_TTL" validate:"required"`
	MfaTokenTTL     time.Duration `mapstructure:"MFA_TOKEN_TTL" validate:"required"`
}
