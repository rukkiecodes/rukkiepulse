from typing import Any

Framework = str  # "fastapi" | "flask" | "starlette" | "unknown"


def detect_framework(app: Any) -> Framework:
    module = type(app).__module__

    if "fastapi" in module:
        return "fastapi"

    if "flask" in module:
        return "flask"

    if "starlette" in module:
        return "starlette"

    # Check by class name as fallback
    class_name = type(app).__name__.lower()
    if "fastapi" in class_name:
        return "fastapi"
    if "flask" in class_name:
        return "flask"

    return "unknown"
