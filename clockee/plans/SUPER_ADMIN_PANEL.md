# Clockee Super Admin Panel — Implementation Plan

## Overview

A standalone web application (`super-admin/`) for the single Clockee platform operator.
Built with **Node.js + Express**, **Apollo Server (GraphQL)**, and **EJS** templating.
Connects to the **same PostgreSQL database** as the existing backend.

---

## What the Super Admin Can Do

| Capability | Description |
|---|---|
| Institution management | View all institutions, suspend, reinstate, delete (soft), edit subscription |
| Admin creation | Create the first admin account for any institution |
| Activity monitoring | See all activity across every institution (attendance, pickups, staff, etc.) |
| Support system | Direct messaging with institution admins via ticket threads |
| Dashboard control | Generate a one-time impersonation token to log into any institution's admin panel |
| Audit log | Full log of every super admin action |
| Notes | Internal notes per institution |
| Platform stats | Cross-institution dashboard: total users, active institutions, today's attendance, etc. |

---

## Tech Stack

| Layer | Choice | Reason |
|---|---|---|
| Runtime | Node.js (CommonJS, same as server) | Consistent with existing backend |
| API | Apollo Server 4 + GraphQL | As specified |
| Web | Express.js | Familiar, session support, EJS integration |
| Frontend | EJS templates + vanilla JS | As specified — no frontend framework |
| Auth | express-session + bcrypt + JWT (for impersonation tokens) | Session for web UI, JWT for impersonation |
| Database | pg (same pool config as server/) | Same shared PostgreSQL DB |
| Real-time | Socket.io | Support ticket notifications |
| CSS | Tailwind CDN (via CDN in EJS head) | Rapid UI, no build step needed |

---

## Project Structure

```
super-admin/
├── src/
│   ├── config/
│   │   ├── database.js           # pg pool (copies server/config/database.js)
│   │   └── constants.js          # roles, statuses, etc.
│   │
│   ├── graphql/
│   │   ├── schema/
│   │   │   ├── index.js          # merges all .graphql files with mergeTypeDefs
│   │   │   ├── auth.graphql
│   │   │   ├── dashboard.graphql
│   │   │   ├── institution.graphql
│   │   │   ├── user.graphql
│   │   │   ├── support.graphql
│   │   │   └── audit.graphql
│   │   ├── resolvers/
│   │   │   ├── index.js          # merges all resolvers with mergeResolvers
│   │   │   ├── authResolvers.js
│   │   │   ├── dashboardResolvers.js
│   │   │   ├── institutionResolvers.js
│   │   │   ├── userResolvers.js
│   │   │   ├── supportResolvers.js
│   │   │   └── auditResolvers.js
│   │   └── context.js            # builds GraphQL context from session/token
│   │
│   ├── middleware/
│   │   ├── requireSuperAdmin.js  # 401 if session.user.role !== 'super_admin'
│   │   └── auditLogger.js        # writes to platform_audit_log on mutations
│   │
│   ├── services/
│   │   ├── AuthService.js
│   │   ├── InstitutionService.js
│   │   ├── SupportService.js
│   │   ├── AuditService.js
│   │   └── ImpersonationService.js
│   │
│   ├── views/
│   │   ├── layouts/
│   │   │   ├── main.ejs          # authenticated shell (sidebar + topbar)
│   │   │   └── auth.ejs          # clean auth shell
│   │   ├── partials/
│   │   │   ├── sidebar.ejs
│   │   │   ├── topbar.ejs
│   │   │   └── flash.ejs         # success/error flash messages
│   │   └── pages/
│   │       ├── login.ejs
│   │       ├── dashboard.ejs
│   │       ├── institutions/
│   │       │   ├── list.ejs      # table: all institutions with filters
│   │       │   ├── detail.ejs    # institution deep-dive + actions
│   │       │   └── create-admin.ejs
│   │       ├── support/
│   │       │   ├── tickets.ejs   # ticket list with status badges
│   │       │   └── thread.ejs    # real-time chat thread
│   │       └── audit/
│   │           └── log.ejs       # paginated audit log with filters
│   │
│   ├── public/
│   │   ├── css/
│   │   │   └── style.css         # minimal custom styles on top of Tailwind CDN
│   │   └── js/
│   │       ├── graphql-client.js # thin fetch wrapper for client-side GQL calls
│   │       └── support-socket.js # Socket.io client for support notifications
│   │
│   └── routes/
│       ├── index.js              # EJS page routes (GET handlers)
│       └── auth.js               # POST /login, GET /logout
│
├── scripts/
│   └── seed-super-admin.js       # one-time: create the super_admin user record
│
├── migrations/
│   └── 001_super_admin_tables.sql
├── app.js                        # Express app setup + Apollo middleware
├── server.js                     # entry point, HTTP + Socket.io
├── package.json
├── .env.example
└── .env
```

---

## Database — New Tables

These tables are added to the **existing shared database**. They are backwards-compatible
and do not modify any existing table.

```sql
-- ==================== SUPPORT TICKETS ====================
CREATE TABLE support_tickets (
  id           SERIAL PRIMARY KEY,
  institution_id INTEGER REFERENCES institutions(id) ON DELETE CASCADE,
  opened_by    INTEGER REFERENCES users(id) ON DELETE SET NULL,
  subject      VARCHAR(255) NOT NULL,
  status       VARCHAR(50)  DEFAULT 'open',   -- open | in_progress | resolved | closed
  priority     VARCHAR(50)  DEFAULT 'normal', -- low | normal | high | urgent
  created_at   TIMESTAMP DEFAULT NOW(),
  updated_at   TIMESTAMP DEFAULT NOW(),
  resolved_at  TIMESTAMP
);

CREATE INDEX idx_support_tickets_institution_id ON support_tickets(institution_id);
CREATE INDEX idx_support_tickets_status ON support_tickets(status);

-- ==================== SUPPORT MESSAGES ====================
CREATE TABLE support_messages (
  id        SERIAL PRIMARY KEY,
  ticket_id INTEGER NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
  sender_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
  sender_role VARCHAR(50) NOT NULL,  -- 'super_admin' | 'admin'
  message   TEXT NOT NULL,
  attachments JSONB DEFAULT '[]',
  read_at   TIMESTAMP,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_support_messages_ticket_id ON support_messages(ticket_id);

-- ==================== PLATFORM AUDIT LOG ====================
CREATE TABLE platform_audit_log (
  id           SERIAL PRIMARY KEY,
  performed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
  action       VARCHAR(100) NOT NULL,  -- e.g. 'SUSPEND_INSTITUTION', 'CREATE_ADMIN'
  target_type  VARCHAR(50),            -- 'institution' | 'user' | 'support_ticket'
  target_id    INTEGER,
  details      JSONB DEFAULT '{}',
  ip_address   VARCHAR(50),
  created_at   TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_platform_audit_log_performed_by ON platform_audit_log(performed_by);
CREATE INDEX idx_platform_audit_log_action ON platform_audit_log(action);
CREATE INDEX idx_platform_audit_log_created_at ON platform_audit_log(created_at);

-- ==================== INSTITUTION NOTES ====================
CREATE TABLE institution_notes (
  id             SERIAL PRIMARY KEY,
  institution_id INTEGER NOT NULL REFERENCES institutions(id) ON DELETE CASCADE,
  note           TEXT NOT NULL,
  created_by     INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at     TIMESTAMP DEFAULT NOW(),
  updated_at     TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_institution_notes_institution_id ON institution_notes(institution_id);
```

No column changes to existing tables. Institution suspension is handled via the existing
`institutions.subscription_status` field (new value: `'suspended'`) and
`users.status = 'suspended'` on the admin user.

---

## GraphQL Schema Summary

### Queries

```graphql
type Query {
  # Auth
  me: SuperAdminProfile

  # Platform dashboard
  platformStats: PlatformStats

  # Institutions
  institutions(filter: InstitutionFilter, page: Int, limit: Int): InstitutionPage
  institution(id: ID!): Institution
  institutionActivity(institutionId: ID!, page: Int, limit: Int): ActivityPage
  institutionUsers(institutionId: ID!, role: String): [User]
  institutionNotes(institutionId: ID!): [InstitutionNote]

  # Support
  supportTickets(status: String, institutionId: ID, page: Int): TicketPage
  supportTicket(id: ID!): SupportTicket

  # Audit
  platformAuditLog(filter: AuditFilter, page: Int, limit: Int): AuditPage
}
```

### Mutations

```graphql
type Mutation {
  # Auth
  login(email: String!, password: String!): AuthPayload

  # Institutions
  suspendInstitution(id: ID!, reason: String!): Institution
  reinstateInstitution(id: ID!): Institution
  deleteInstitution(id: ID!, reason: String!): Boolean
  updateInstitutionSubscription(id: ID!, data: SubscriptionInput!): Institution
  addInstitutionNote(institutionId: ID!, note: String!): InstitutionNote

  # Users
  createInstitutionAdmin(institutionId: ID!, data: CreateAdminInput!): User
  suspendUser(userId: ID!, reason: String!): User
  reinstateUser(userId: ID!): User

  # Support
  createSupportTicket(data: CreateTicketInput!): SupportTicket
  sendSupportMessage(ticketId: ID!, message: String!): SupportMessage
  updateTicketStatus(ticketId: ID!, status: String!): SupportTicket

  # Impersonation
  generateImpersonationToken(institutionId: ID!): ImpersonationPayload
}
```

### Subscriptions

```graphql
type Subscription {
  supportMessageReceived(ticketId: ID!): SupportMessage
  newSupportTicket: SupportTicket
}
```

---

## Authentication Flow

### Super Admin Login (Web UI)
1. `POST /auth/login` → validates credentials against `users` table (`role = 'super_admin'`)
2. Creates `express-session` → stores `{ id, email, role, name }`
3. Redirects to `/dashboard`

### Session Protection
- All EJS routes guarded by `requireSuperAdmin` middleware
- GraphQL context reads `req.session.user` and attaches to context
- GraphQL resolvers call `context.assertSuperAdmin()` at the top

### Impersonation Flow
1. Super admin clicks **"Control Dashboard"** on any institution's detail page
2. Frontend calls GraphQL mutation `generateImpersonationToken(institutionId)`
3. Server:
   - Looks up the institution's primary admin user
   - Signs a short-lived JWT: `{ userId, institutionId, impersonatedBy: superAdminId, exp: +15min }`
   - Logs the action to `platform_audit_log`
4. Frontend receives the token and opens:
   `https://admin.clockee.app/auth/impersonate?token=<jwt>`
5. The **existing admin panel** has a new route `GET /auth/impersonate` that:
   - Verifies the JWT (signed with a shared `IMPERSONATION_SECRET` env var)
   - Creates a normal admin session for that user
   - Redirects to the admin dashboard

This requires one small addition to the existing `server/routes/auth.js`:
```js
// GET /api/auth/impersonate?token=xxx
router.get('/impersonate', ImpersonationController.verify);
```

---

## EJS Pages — Layout & Content

### `/login`
- Clean centred card
- Email + password fields
- Error flash on failure

### `/dashboard`
Platform-wide stats grid:
- Total institutions (active / suspended / trial)
- Total users across all institutions
- Total attendance clock-ins today
- Open support tickets
- Institutions with expiring subscriptions (next 7 days)
- Recent activity feed (last 20 audit log entries)

### `/institutions`
Filterable table:
- Columns: Name, Type, City, Subscription status, Staff count, Created date, Actions
- Filters: status, subscription type, country
- Row actions: View, Suspend, Delete

### `/institutions/:id`
Institution detail page with tabs:
- **Overview** — details, subscription, GPS, QR code preview
- **Users** — staff/admin list with suspend/reinstate per user
- **Activity** — attendance events, pickup events, recent logins
- **Support** — tickets opened by this institution
- **Notes** — internal super admin notes
- **Actions** — Create Admin, Suspend Institution, Control Dashboard (impersonation)

### `/institutions/:id/create-admin`
Form: name, email, password, phone → calls `createInstitutionAdmin` mutation

### `/support`
Ticket list with status filter (open / in_progress / resolved / closed)

### `/support/:id`
Real-time thread view:
- Messages displayed as chat bubbles (super admin right, admin left)
- Input box sends via Socket.io emit → stored in DB → broadcast to both parties
- Status controls at top

### `/audit`
Paginated table:
- Columns: Timestamp, Action, Target, Details, IP
- Filters: action type, date range, target institution

---

## Seeding the Super Admin

The super admin record is **not registered** — it is seeded via script:

```js
// scripts/seed-super-admin.js
const bcrypt = require('bcryptjs');
const pool = require('../src/config/database');

async function seed() {
  const hash = await bcrypt.hash(process.env.SUPER_ADMIN_PASSWORD, 12);
  await pool.query(`
    INSERT INTO users (email, password, name, role, status, email_verified_at)
    VALUES ($1, $2, 'Super Admin', 'super_admin', 'approved', NOW())
    ON CONFLICT (email) DO NOTHING
  `, [process.env.SUPER_ADMIN_EMAIL, hash]);
  console.log('Super admin seeded');
}
seed();
```

Run once: `node scripts/seed-super-admin.js`

---

## Environment Variables (.env.example)

```env
# Server
PORT=4000
NODE_ENV=development

# Database (same as existing server .env)
DATABASE_URL=postgresql://user:pass@host:5432/clockee

# Session
SESSION_SECRET=your-session-secret

# JWT (shared with main server for impersonation)
JWT_SECRET=same-as-main-server-jwt-secret
IMPERSONATION_SECRET=separate-secret-for-impersonation-tokens

# Super admin seed
SUPER_ADMIN_EMAIL=superadmin@clockee.app
SUPER_ADMIN_PASSWORD=strongpassword

# Admin panel URL (for impersonation redirect)
ADMIN_PANEL_URL=https://admin.clockee.app
```

---

## Package Dependencies

```json
{
  "name": "clockee-super-admin",
  "version": "1.0.0",
  "type": "commonjs",
  "dependencies": {
    "@apollo/server": "^4.x",
    "@graphql-tools/merge": "^9.x",
    "express": "^4.x",
    "express-session": "^1.x",
    "connect-pg-simple": "^9.x",
    "graphql": "^16.x",
    "pg": "^8.x",
    "bcryptjs": "^2.x",
    "jsonwebtoken": "^9.x",
    "ejs": "^3.x",
    "socket.io": "^4.x",
    "graphql-ws": "^5.x",
    "ws": "^8.x",
    "dotenv": "^16.x",
    "morgan": "^1.x"
  },
  "devDependencies": {
    "nodemon": "^3.x"
  }
}
```

Sessions stored in PostgreSQL via `connect-pg-simple` (no Redis dependency).

---

## Implementation Order

### Phase 1 — Foundation
1. Create `super-admin/` project directory and `package.json`
2. Copy/adapt `database.js` config
3. Run `migrations/001_super_admin_tables.sql`
4. Set up Express app with EJS + session + Morgan
5. Create login/logout routes and `login.ejs`
6. Seed super admin via script

### Phase 2 — GraphQL Core
7. Set up Apollo Server integrated with Express (`/graphql` endpoint)
8. Build GraphQL context (reads session, exposes `assertSuperAdmin()`)
9. Implement `auth.graphql` + `authResolvers.js` (login mutation)
10. Implement `dashboard.graphql` + `dashboardResolvers.js` (platformStats query)
11. Implement `institution.graphql` + `institutionResolvers.js` (all institution queries/mutations)
12. Implement `user.graphql` + `userResolvers.js` (createInstitutionAdmin, suspend/reinstate)
13. Implement `audit.graphql` + `auditResolvers.js` (log query + auditLogger middleware)

### Phase 3 — EJS Pages
14. Build `main.ejs` layout (sidebar + topbar using Tailwind CDN)
15. Build `dashboard.ejs` (stats cards, recent activity feed)
16. Build `institutions/list.ejs` + `institutions/detail.ejs`
17. Build `institutions/create-admin.ejs`
18. Build `audit/log.ejs`

### Phase 4 — Support System
19. Implement `support.graphql` + `supportResolvers.js`
20. Set up Socket.io server + `support-socket.js` client
21. Build `support/tickets.ejs` + `support/thread.ejs`

### Phase 5 — Impersonation
22. Add `ImpersonationService.js` + `generateImpersonationToken` mutation
23. Add `GET /api/auth/impersonate` route to existing **server/** backend
24. Wire "Control Dashboard" button in `institutions/detail.ejs`

### Phase 6 — Polish
25. Add flash messages (success/error) to all mutations
26. Add pagination to institutions list, audit log, support tickets
27. Add institution suspension warning email (reuse existing EmailService)
28. Final audit log coverage — ensure every mutation writes to `platform_audit_log`

---

## What Is NOT Included

- No billing/payment processing (subscription dates are set manually)
- No super admin registration UI (seeded via script only)
- No mobile app for super admin
- No changes to the main backend business logic
- No changes to admin panel, staff app, or parent app (except the one impersonation route)

---

## Files to Add to Existing Backend (server/)

Only **one new route** is needed in the existing server:

```
server/controllers/auth/ImpersonationController.js
server/routes/auth.js  ← add: GET /api/auth/impersonate
```

Everything else lives entirely within the `super-admin/` project.
