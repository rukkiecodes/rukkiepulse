from dataclasses import dataclass, field
from typing import Any, Callable, Dict, Optional

from .otel import setup_otel
from .detect import detect_framework
from .middleware import inject_middleware
from .health import register_health_endpoint


@dataclass
class RukkieConfig:
    service_name: str
    api_key: str
    collector_url: str = "http://localhost:4317"
    dependencies: Dict[str, Callable[[], bool]] = field(default_factory=dict)


def init_rukkie(
    service_name: str,
    api_key: str,
    app: Optional[Any] = None,
    collector_url: str = "http://localhost:4317",
    dependencies: Optional[Dict[str, Callable[[], bool]]] = None,
) -> None:
    """
    Initialize the RukkiePulse agent.

    Call this as early as possible in your application startup,
    before registering routes.

    Examples::

        # FastAPI
        from fastapi import FastAPI
        from rukkie_agent import init_rukkie

        app = FastAPI()
        init_rukkie("auth-service", "rk_live_xxx", app=app)

        # Flask
        from flask import Flask
        from rukkie_agent import init_rukkie

        app = Flask(__name__)
        init_rukkie("auth-service", "rk_live_xxx", app=app)
    """
    config = RukkieConfig(
        service_name=service_name,
        api_key=api_key,
        collector_url=collector_url,
        dependencies=dependencies or {},
    )

    # 1. Boot OTel SDK first
    setup_otel(config)

    if app is None:
        return

    framework = detect_framework(app)

    # 2. Inject request tracing middleware
    inject_middleware(app, framework, config)

    # 3. Register /__rukkie/health endpoint
    register_health_endpoint(app, framework, config)
