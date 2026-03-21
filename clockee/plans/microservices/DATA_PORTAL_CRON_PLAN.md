# Data Portal — Vercel Cron: Scheduled Deletion Executor

## Background

The data-portal runs on Vercel serverless functions. `scripts/cron.py` (APScheduler,
daily 02:00) cannot run in this environment. Instead, Vercel Cron Jobs call an HTTP
endpoint on a schedule. This plan covers wiring that endpoint up.

---

## What Needs to Be Built

### 1. Cron endpoint — `app/routers/cron.py`

```
POST /cron/run-deletions
```

- Protected by a `CRON_SECRET` header (`Authorization: Bearer <secret>`)
- Queries `data_deletion_requests` for all rows where:
  - `status = 'pending'`
  - `scheduled_delete_at <= NOW()`
- Calls `deletion_service.execute_deletion(pool, request_id)` for each
- Returns `{ "processed": N, "errors": [...] }`

```python
# app/routers/cron.py
from fastapi import APIRouter, Depends, Header, HTTPException, Request
from app.database import get_pool
from app.services import deletion_service
from app.config import CRON_SECRET

router = APIRouter()

@router.post("/run-deletions")
async def run_deletions(
    authorization: str = Header(...),
    pool=Depends(get_pool),
):
    if authorization != f"Bearer {CRON_SECRET}":
        raise HTTPException(status_code=401, detail="Unauthorized")

    async with pool.acquire() as conn:
        due = await conn.fetch(
            """SELECT id FROM data_deletion_requests
               WHERE status = 'pending' AND scheduled_delete_at <= NOW()"""
        )

    processed, errors = 0, []
    for row in due:
        try:
            await deletion_service.execute_deletion(pool, row["id"])
            processed += 1
        except Exception as e:
            errors.append({"request_id": row["id"], "error": str(e)})

    return {"processed": processed, "errors": errors}
```

### 2. Register the router in `app/main.py`

```python
from app.routers import cron as cron_router
app.include_router(cron_router.router, prefix="/cron", tags=["cron"])
```

### 3. New env var — `CRON_SECRET`

Add to `app/config.py`:
```python
CRON_SECRET = os.getenv("CRON_SECRET", "")
```

Set in Vercel dashboard: a long random string (e.g. `openssl rand -hex 32`).

### 4. Vercel Cron Job config — `vercel.json`

Add a `crons` block to the existing `vercel.json`:

```json
{
  "version": 2,
  "builds": [ ... ],
  "routes": [ ... ],
  "crons": [
    {
      "path": "/cron/run-deletions",
      "schedule": "0 2 * * *"
    }
  ]
}
```

> Vercel Cron Jobs send a GET request by default on Hobby plans, but POST on Pro.
> On Hobby: change the endpoint to `GET /cron/run-deletions` and verify by the
> `x-vercel-signature` header instead of a Bearer token (see Vercel docs).
> On Pro: POST + Bearer token as described above works directly.

### 5. Vercel automatically injects the `CRON_SECRET`

Vercel sends a `Authorization: Bearer <CRON_SECRET>` header on Pro plans when
you configure the secret in the dashboard. On Hobby plans, use
`x-vercel-signature` HMAC verification instead.

---

## Files to Create / Modify

| File | Change |
|------|--------|
| `app/routers/cron.py` | New — the endpoint |
| `app/main.py` | Add `include_router(cron_router.router, prefix="/cron")` |
| `app/config.py` | Add `CRON_SECRET = os.getenv("CRON_SECRET", "")` |
| `vercel.json` | Add `"crons"` block |

---

## Environment Variables

| Variable | Where | Value |
|----------|-------|-------|
| `CRON_SECRET` | Vercel dashboard | Random 32-byte hex string |

---

## Testing Locally

```bash
# Trigger manually with curl
curl -X POST http://localhost:4002/cron/run-deletions \
  -H "Authorization: Bearer your-cron-secret"
```

---

## Notes

- `execute_deletion` already handles the correct deletion order (marks request
  completed before deleting user to avoid FK violations).
- Errors for individual requests are caught and logged in the response — one
  failed deletion does not block others.
- The endpoint is idempotent: re-running it will find no more `pending` rows
  past their date and return `{ "processed": 0 }`.
