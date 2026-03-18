from __future__ import annotations

import time
from typing import TYPE_CHECKING, Any

from opentelemetry import trace
from opentelemetry.trace import SpanKind, StatusCode

if TYPE_CHECKING:
    from .setup import RukkieConfig


def inject_middleware(app: Any, framework: str, config: "RukkieConfig") -> None:
    if framework == "fastapi":
        _inject_fastapi(app, config)
    elif framework in ("flask", "unknown"):
        _inject_flask(app, config)
    # starlette: handled by FastAPI instrumentation or standalone starlette middleware


def _inject_fastapi(app: Any, config: "RukkieConfig") -> None:
    try:
        from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
        FastAPIInstrumentor.instrument_app(app)
    except ImportError:
        # Fallback: manual Starlette middleware
        from starlette.middleware.base import BaseHTTPMiddleware
        app.add_middleware(RukkieASGIMiddleware, service_name=config.service_name)


def _inject_flask(app: Any, config: "RukkieConfig") -> None:
    try:
        from opentelemetry.instrumentation.flask import FlaskInstrumentor
        FlaskInstrumentor().instrument_app(app)
    except ImportError:
        # Fallback: manual before/after request hooks
        tracer = trace.get_tracer(config.service_name)
        _state: dict = {}

        @app.before_request
        def before_request():  # type: ignore[misc]
            from flask import request
            span = tracer.start_span(
                f"{request.method} {request.path}",
                kind=SpanKind.SERVER,
            )
            _state["span"] = span
            _state["start"] = time.time()

        @app.after_request
        def after_request(response):  # type: ignore[misc]
            span = _state.pop("span", None)
            if span:
                duration_ms = int((time.time() - _state.pop("start", time.time())) * 1000)
                span.set_attribute("http.status_code", response.status_code)
                span.set_attribute("http.duration_ms", duration_ms)
                if response.status_code >= 500:
                    span.set_status(StatusCode.ERROR)
                span.end()
            return response


class RukkieASGIMiddleware:
    """Minimal ASGI middleware fallback when OTel FastAPI instrumentation is unavailable."""

    def __init__(self, app: Any, service_name: str) -> None:
        self.app = app
        self.tracer = trace.get_tracer(service_name)

    async def __call__(self, scope: Any, receive: Any, send: Any) -> None:
        if scope["type"] != "http":
            await self.app(scope, receive, send)
            return

        method = scope.get("method", "GET")
        path = scope.get("path", "/")
        status_code = [200]

        async def send_wrapper(message: Any) -> None:
            if message["type"] == "http.response.start":
                status_code[0] = message.get("status", 200)
            await send(message)

        with self.tracer.start_as_current_span(
            f"{method} {path}", kind=SpanKind.SERVER
        ) as span:
            try:
                await self.app(scope, receive, send_wrapper)
                span.set_attribute("http.status_code", status_code[0])
                if status_code[0] >= 500:
                    span.set_status(StatusCode.ERROR)
            except Exception as exc:
                span.record_exception(exc)
                span.set_status(StatusCode.ERROR, str(exc))
                raise
