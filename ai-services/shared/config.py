from __future__ import annotations

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    env: str = "development"
    log_level: str = "info"

    postgres_host: str = "localhost"
    postgres_port: int = 5432
    postgres_user: str = "enterprise"
    postgres_password: str = "enterprise"
    postgres_db: str = "enterprise_ai"

    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_password: str = ""
    redis_db: int = 0

    s3_endpoint: str = "http://localhost:9000"
    s3_access_key: str = "minio"
    s3_secret_key: str = "minio123"
    s3_bucket: str = "enterprise-docs"
    s3_region: str = "us-east-1"

    milvus_host: str = "localhost"
    milvus_port: int = 19530

    class Config:
        env_file = ".env"
        extra = "ignore"


def load_settings() -> Settings:
    return Settings()
