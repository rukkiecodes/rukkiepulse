# rukkie-agent (Python)

Auto-instrument Python backend services for [RukkiePulse](https://rukkiepulse.netlify.app).

## Install

```bash
pip install rukkie-agent
```

## Usage

```python
# FastAPI
from fastapi import FastAPI
from rukkie_agent import init_rukkie

app = FastAPI()

init_rukkie(
    service_name="auth-service",
    api_key="rk_live_xxx",
    app=app,
    dependencies={
        "db": check_db_connection,
        "redis": check_redis_connection,
    }
)

# Flask
from flask import Flask
from rukkie_agent import init_rukkie

app = Flask(__name__)
init_rukkie(service_name="auth-service", api_key="rk_live_xxx", app=app)
```

Exposes `GET /__rukkie/health` and pushes traces to your Jaeger collector.

## Docs

[rukkiepulse.netlify.app](https://rukkiepulse.netlify.app)
