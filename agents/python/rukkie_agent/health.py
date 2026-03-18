from __future__ import annotations

import asyncio
from typing import TYPE_CHECKING, Any, Dict

if TYPE_CHECKING:
    from .setup import RukkieConfig


def register_health_endpoint(app: Any, framework: str, config: "RukkieConfig") -> None:
    if framework == "fastapi":
        _register_fastapi(app, config)
    elif framework == "flask":
        _register_flask(app, config)
    elif framework == "starlette":
        _register_fastapi(app, config)  # same API


def _register_fastapi(app: Any, config: "RukkieConfig") -> None:
    async def health_handler():  # type: ignore[misc]
        return await _build_response(config)

    app.add_api_route("/__rukkie/health", health_handler, methods=["GET"])


def _register_flask(app: Any, config: "RukkieConfig") -> None:
    from flask import jsonify

    @app.route("/__rukkie/health", methods=["GET"])
    def rukkie_health():  # type: ignore[misc]
        body = asyncio.run(_build_response(config))
        status = 200 if body["status"] == "ok" else 207
        return jsonify(body), status


async def _build_response(config: "RukkieConfig") -> Dict[str, Any]:
    deps: Dict[str, str] = {}

    if config.dependencies:
        async def check_one(name: str, fn: Any) -> None:
            try:
                result = fn()
                if asyncio.iscoroutine(result):
                    result = await result
                deps[name] = "connected" if result else "error"
            except Exception:
                deps[name] = "error"

        await asyncio.gather(*(check_one(n, f) for n, f in config.dependencies.items()))

    all_ok = all(v == "connected" for v in deps.values())
    status = "ok" if (not deps or all_ok) else "degraded"

    return {
        "status": status,
        "service": config.service_name,
        "dependencies": deps,
    }
