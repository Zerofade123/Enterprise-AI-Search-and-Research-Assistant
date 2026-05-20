from __future__ import annotations

from ai_services.shared.config import load_settings


def test_load_settings_defaults() -> None:
    settings = load_settings()
    assert settings.env == "development"
    assert settings.postgres_port == 5432
