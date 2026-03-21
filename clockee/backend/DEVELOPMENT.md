# Clockee Backend — Development Guide

## Prerequisites

| Tool | Min version | Notes |
|------|-------------|-------|
| Node.js | 18+ | |
| Python | 3.11+ | For the four Python micro-services |
| PostgreSQL | 14+ | Local or remote (set `DATABASE_URL`) |

---

## Running all services together

From the **`backend/`** directory:

```bash
npm install       # install concurrently (one-time)
npm run dev       # starts all 7 services with hot-reload
```

| Label | Service | Port |
|-------|---------|------|
| `server` | Main API (Node/Express) | 3000 |
| `password-reset` | Password reset service | auto |
| `attendance` | Students attendance service | auto |
| `super-admin` | Super-admin panel (Python/FastAPI) | 4000 |
| `data-portal` | Data portal (Python/FastAPI) | 4002 |
| `reports` | Reports service (Python/FastAPI) | 4003 |
| `auditing` | Staff auditing (Python/FastAPI) | 4004 |

---

## Running individual services

### Main server (Node.js / Express)

```bash
cd backend/server
npm install
npm run dev
```

### Password-reset service (Node.js)

```bash
cd backend/services/password-reset
npm install
npm run dev
```

### Students attendance service (Node.js / Apollo GraphQL)

```bash
cd backend/services/students-attendance-service
npm install
npm run dev
```

### Super-admin panel (Python / FastAPI) — port 4000

```bash
cd backend/services/super-admin
pip install -r requirements.txt
python -m uvicorn app.main:app --reload --port 4000
```

### Data portal (Python / FastAPI) — port 4002

```bash
cd backend/services/data-portal
pip install -r requirements.txt
python -m uvicorn app.main:app --reload --port 4002
```

### Reports service (Python / FastAPI) — port 4003

```bash
cd backend/services/reports
pip install -r requirements.txt
python -m uvicorn app.main:app --reload --port 4003
```

### Auditing service (Python / FastAPI) — port 4004

```bash
cd backend/services/auditing
pip install -r requirements.txt
python -m uvicorn app.main:app --reload --port 4004
```

---

## Database migrations

All migration scripts are Node.js files. Run them with:

```bash
node <path-to-migration-file>
```

### Super-admin migrations

| File | What it does | How to run |
|------|-------------|------------|
| `services/super-admin/migrations/001_super_admin_tables.sql` | Creates `support_tickets`, `support_messages`, `platform_audit_log`, `institution_notes` | Run as SQL in psql / any SQL client |
| `services/super-admin/migrations/002_patch_support.js` | Adds missing columns to existing tables (run if 001 was never applied) | `node backend/services/super-admin/migrations/002_patch_support.js` |

> The JS migrations read `DATABASE_URL` from `backend/server/.env`.

### Server migrations — patch scripts

| File | What it does | How to run |
|------|-------------|------------|
| `server/migrations/003_home_arrivals.js` | Creates `child_home_arrivals` table for parent home-arrival QR scan | `node backend/server/migrations/003_home_arrivals.js` |

```bash
# Initial schema
node backend/server/scripts/database/setup-db.js

# Seed default data
node backend/server/scripts/database/seed.js

# Drop all tables (destructive!)
node backend/server/scripts/database/drop-tables.js
```

SQL migration files live in `backend/server/scripts/database/migrations/`. Run them directly in psql:

```bash
psql $DATABASE_URL -f backend/server/scripts/database/migrations/<filename>.sql
```

---

## Environment variables

Copy `backend/server/.env.example` to `backend/server/.env`:

```env
DATABASE_URL=postgresql://user:pass@localhost:5432/clockee
JWT_SECRET=...
REFRESH_TOKEN_SECRET=...
TERMII_API_KEY=...
TERMII_BASE_URI=https://v3.api.termii.com
TERMII_SENDER_ID=Clockee
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
IMPERSONATION_SECRET=change-this-in-production
ADMIN_PANEL_URL=http://localhost:5173
```

Python services share the same `.env` — they load it from `backend/server/.env` via `python-dotenv`.

---

## Super-admin

URL: `http://localhost:4000`

Default login (seeded via `seed.js`):

| Field | Value |
|-------|-------|
| Email | `superadmin@clockee.com` |
| Password | `SuperAdmin@123` |

---

## Deploying to Vercel

Each service is deployed as a **separate Vercel project**. All services have a `vercel.json` at their root.

### Service deployment table

| Service | Root directory | Framework | Vercel URL (example) |
|---------|---------------|-----------|----------------------|
| Main server | `backend/server` | Node.js | `clockee-backend.vercel.app` |
| Password reset | `backend/services/password-reset` | Node.js | `clockee-password-reset.vercel.app` |
| Students attendance | `backend/services/students-attendance-service` | Node.js | `clockee-attendance.vercel.app` |
| Super-admin | `backend/services/super-admin` | Python | `clockee-super-admin.vercel.app` |
| Data portal | `backend/services/data-portal` | Python | `clockee-data-portal.vercel.app` |
| Reports | `backend/services/reports` | Python | `clockee-reports.vercel.app` |
| Auditing | `backend/services/auditing` | Python | `clockee-auditing.vercel.app` |

### Steps to deploy a service

```bash
# 1. Install Vercel CLI (one-time)
npm i -g vercel

# 2. Navigate to the service directory
cd backend/services/reports   # or whichever service

# 3. Deploy
vercel --prod
```

When prompted by the Vercel CLI, set the **Root Directory** to `.` (the service folder you're in).

### Environment variables to set on Vercel

For every service, add these in the Vercel project dashboard → Settings → Environment Variables:

| Variable | All services | Notes |
|----------|-------------|-------|
| `DATABASE_URL` | ✅ | Full PostgreSQL connection string incl. SSL params |
| `JWT_SECRET` | ✅ | Must be identical across all services |
| `DB_SSLMODE` | Python services | Set to `require` for hosted DBs |
| `ALLOWED_ORIGINS` | ✅ | Comma-separated frontend URLs |
| `SUPER_ADMIN_API_KEY` | super-admin, reports, auditing | Shared secret for cross-service calls |
| `CLOUDINARY_*` | main server | Cloud storage |
| `TERMII_*` | main server | SMS provider |
| `SESSION_SECRET` | data-portal | Random string |

> **Tip:** Use Vercel's "Link environment variables" to share `DATABASE_URL` and `JWT_SECRET` across all projects from a single source.

### DATABASE_URL format for SSL

Most hosted PostgreSQL providers require SSL. Use:

```
postgresql://user:password@host:5432/dbname?sslmode=require
```

Or set `DATABASE_URL` without SSL params and add `DB_SSLMODE=require` separately (Python services read this).

### Post-deploy migrations

Run these once against the production database before going live:

```bash
# From backend/server/ (reads DATABASE_URL from .env)
node migrations/003_home_arrivals.js
node migrations/004_two_factor_and_otp.js

# From backend/services/super-admin/ (needs npm install first)
node migrations/002_patch_support.js

# Students attendance table
node backend/services/students-attendance-service/scripts/migrate.js
```
