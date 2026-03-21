# Phase 1 — Foundation

## Goal
Bootstrap the `super-admin/` project: directory structure, Express app, session auth, login/logout UI, database migration, and seed script.

---

## Scope

### Files to Create

```
super-admin/
├── package.json
├── .env.example
├── .env                          ← copy from .env.example, fill in values
├── app.js                        ← Express app setup
├── server.js                     ← entry point
├── migrations/
│   └── 001_super_admin_tables.sql
├── scripts/
│   └── seed-super-admin.js
├── src/
│   ├── config/
│   │   ├── database.js
│   │   └── constants.js
│   ├── middleware/
│   │   └── requireSuperAdmin.js
│   ├── routes/
│   │   ├── index.js              ← EJS page GET routes
│   │   └── auth.js               ← POST /login, GET /logout
│   ├── public/
│   │   ├── css/
│   │   │   └── style.css
│   │   └── js/
│   │       └── graphql-client.js ← thin fetch wrapper (stubbed for now)
│   └── views/
│       ├── layouts/
│       │   ├── main.ejs          ← authenticated shell
│       │   └── auth.ejs          ← clean auth shell
│       ├── partials/
│       │   ├── sidebar.ejs
│       │   ├── topbar.ejs
│       │   └── flash.ejs
│       └── pages/
│           ├── login.ejs
│           └── dashboard.ejs     ← placeholder (fleshed out in Phase 3)
```

---

## Implementation Steps

### Step 1 — package.json
Create `super-admin/package.json` with all dependencies:

```json
{
  "name": "clockee-super-admin",
  "version": "1.0.0",
  "type": "commonjs",
  "main": "server.js",
  "scripts": {
    "start": "node server.js",
    "dev": "nodemon server.js"
  },
  "dependencies": {
    "@apollo/server": "^4.11.0",
    "@graphql-tools/merge": "^9.0.0",
    "bcryptjs": "^2.4.3",
    "connect-pg-simple": "^9.0.1",
    "dotenv": "^16.0.0",
    "ejs": "^3.1.10",
    "express": "^4.18.0",
    "express-session": "^1.17.3",
    "graphql": "^16.8.0",
    "graphql-ws": "^5.16.0",
    "jsonwebtoken": "^9.0.0",
    "morgan": "^1.10.0",
    "pg": "^8.11.0",
    "socket.io": "^4.7.0",
    "ws": "^8.17.0"
  },
  "devDependencies": {
    "nodemon": "^3.0.0"
  }
}
```

### Step 2 — .env.example

```env
PORT=4000
NODE_ENV=development

DATABASE_URL=postgresql://user:pass@host:5432/clockee

SESSION_SECRET=your-session-secret

JWT_SECRET=same-as-main-server-jwt-secret
IMPERSONATION_SECRET=separate-secret-for-impersonation-tokens

SUPER_ADMIN_EMAIL=superadmin@clockee.app
SUPER_ADMIN_PASSWORD=strongpassword

ADMIN_PANEL_URL=https://admin.clockee.app
```

### Step 3 — Database migration
File: `migrations/001_super_admin_tables.sql`

```sql
-- Support tickets
CREATE TABLE IF NOT EXISTS support_tickets (
  id             SERIAL PRIMARY KEY,
  institution_id INTEGER REFERENCES institutions(id) ON DELETE CASCADE,
  opened_by      INTEGER REFERENCES users(id) ON DELETE SET NULL,
  subject        VARCHAR(255) NOT NULL,
  status         VARCHAR(50)  DEFAULT 'open',
  priority       VARCHAR(50)  DEFAULT 'normal',
  created_at     TIMESTAMP    DEFAULT NOW(),
  updated_at     TIMESTAMP    DEFAULT NOW(),
  resolved_at    TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_support_tickets_institution_id ON support_tickets(institution_id);
CREATE INDEX IF NOT EXISTS idx_support_tickets_status ON support_tickets(status);

-- Support messages
CREATE TABLE IF NOT EXISTS support_messages (
  id          SERIAL PRIMARY KEY,
  ticket_id   INTEGER NOT NULL REFERENCES support_tickets(id) ON DELETE CASCADE,
  sender_id   INTEGER REFERENCES users(id) ON DELETE SET NULL,
  sender_role VARCHAR(50) NOT NULL,
  message     TEXT NOT NULL,
  attachments JSONB DEFAULT '[]',
  read_at     TIMESTAMP,
  created_at  TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_support_messages_ticket_id ON support_messages(ticket_id);

-- Platform audit log
CREATE TABLE IF NOT EXISTS platform_audit_log (
  id           SERIAL PRIMARY KEY,
  performed_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
  action       VARCHAR(100) NOT NULL,
  target_type  VARCHAR(50),
  target_id    INTEGER,
  details      JSONB DEFAULT '{}',
  ip_address   VARCHAR(50),
  created_at   TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_performed_by ON platform_audit_log(performed_by);
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_action ON platform_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_platform_audit_log_created_at ON platform_audit_log(created_at);

-- Institution notes
CREATE TABLE IF NOT EXISTS institution_notes (
  id             SERIAL PRIMARY KEY,
  institution_id INTEGER NOT NULL REFERENCES institutions(id) ON DELETE CASCADE,
  note           TEXT NOT NULL,
  created_by     INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at     TIMESTAMP DEFAULT NOW(),
  updated_at     TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_institution_notes_institution_id ON institution_notes(institution_id);
```

Run with:
```
psql $DATABASE_URL -f migrations/001_super_admin_tables.sql
```

### Step 4 — src/config/database.js

Copy from `server/config/database.js` (same pg pool pattern, reads `DATABASE_URL` from `.env`).

### Step 5 — src/config/constants.js

```js
module.exports = {
  ROLES: { SUPER_ADMIN: 'super_admin', ADMIN: 'admin', STAFF: 'staff' },
  TICKET_STATUSES: ['open', 'in_progress', 'resolved', 'closed'],
  TICKET_PRIORITIES: ['low', 'normal', 'high', 'urgent'],
  INSTITUTION_STATUSES: ['active', 'suspended', 'trial', 'expired'],
};
```

### Step 6 — src/middleware/requireSuperAdmin.js

```js
module.exports = function requireSuperAdmin(req, res, next) {
  if (req.session && req.session.user && req.session.user.role === 'super_admin') {
    return next();
  }
  res.redirect('/login');
};
```

### Step 7 — src/routes/auth.js (POST /login, GET /logout)

```js
const express = require('express');
const bcrypt = require('bcryptjs');
const pool = require('../config/database');
const router = express.Router();

router.post('/login', async (req, res) => {
  const { email, password } = req.body;
  try {
    const result = await pool.query(
      "SELECT id, email, name, password, role FROM users WHERE email = $1 AND role = 'super_admin'",
      [email]
    );
    const user = result.rows[0];
    if (!user || !(await bcrypt.compare(password, user.password))) {
      return res.render('pages/login', { error: 'Invalid email or password.' });
    }
    req.session.user = { id: user.id, email: user.email, name: user.name, role: user.role };
    res.redirect('/dashboard');
  } catch (err) {
    console.error('[auth] login error', err);
    res.render('pages/login', { error: 'Login failed. Please try again.' });
  }
});

router.get('/logout', (req, res) => {
  req.session.destroy(() => res.redirect('/login'));
});

module.exports = router;
```

### Step 8 — src/routes/index.js (GET page routes)

```js
const express = require('express');
const requireSuperAdmin = require('../middleware/requireSuperAdmin');
const router = express.Router();

router.get('/login', (req, res) => {
  if (req.session && req.session.user) return res.redirect('/dashboard');
  res.render('pages/login', { error: null });
});

router.get('/dashboard', requireSuperAdmin, (req, res) => {
  res.render('pages/dashboard', { user: req.session.user });
});

module.exports = router;
```

### Step 9 — app.js

```js
require('dotenv').config();
const express = require('express');
const session = require('express-session');
const pgSession = require('connect-pg-simple')(session);
const morgan = require('morgan');
const path = require('path');
const pool = require('./src/config/database');
const pageRoutes = require('./src/routes/index');
const authRoutes = require('./src/routes/auth');

const app = express();

app.set('view engine', 'ejs');
app.set('views', path.join(__dirname, 'src/views'));

app.use(morgan('dev'));
app.use(express.static(path.join(__dirname, 'src/public')));
app.use(express.urlencoded({ extended: true }));
app.use(express.json());

app.use(session({
  store: new pgSession({ pool, tableName: 'session' }),
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false,
  cookie: { maxAge: 24 * 60 * 60 * 1000, httpOnly: true, secure: process.env.NODE_ENV === 'production' },
}));

app.use('/', pageRoutes);
app.use('/auth', authRoutes);

module.exports = app;
```

### Step 10 — server.js

```js
const app = require('./app');
const PORT = process.env.PORT || 4000;
app.listen(PORT, () => console.log(`[super-admin] Running on port ${PORT}`));
```

### Step 11 — EJS views

**`src/views/layouts/auth.ejs`** — minimal centered card layout (Tailwind CDN).

**`src/views/layouts/main.ejs`** — sidebar + topbar layout (Tailwind CDN). Sidebar links: Dashboard, Institutions, Support, Audit Log.

**`src/views/partials/sidebar.ejs`** — nav links with active-state highlighting.

**`src/views/partials/topbar.ejs`** — user name, logout button.

**`src/views/partials/flash.ejs`** — renders `locals.success` and `locals.error` flash banners.

**`src/views/pages/login.ejs`** — email + password form. Posts to `/auth/login`. Shows `error` if present.

**`src/views/pages/dashboard.ejs`** — placeholder card: "Dashboard coming in Phase 3".

### Step 12 — Seed script

`scripts/seed-super-admin.js`:
```js
require('dotenv').config();
const bcrypt = require('bcryptjs');
const pool = require('../src/config/database');

async function seed() {
  const hash = await bcrypt.hash(process.env.SUPER_ADMIN_PASSWORD, 12);
  const result = await pool.query(`
    INSERT INTO users (email, password, name, role, status, email_verified_at)
    VALUES ($1, $2, 'Super Admin', 'super_admin', 'approved', NOW())
    ON CONFLICT (email) DO UPDATE SET password = $2
    RETURNING id, email
  `, [process.env.SUPER_ADMIN_EMAIL, hash]);
  console.log('Super admin seeded:', result.rows[0]);
  process.exit(0);
}
seed().catch(err => { console.error(err); process.exit(1); });
```

Run: `node scripts/seed-super-admin.js`

---

## Install & Run

```bash
cd super-admin
npm install
cp .env.example .env
# fill in .env
psql $DATABASE_URL -f migrations/001_super_admin_tables.sql
node scripts/seed-super-admin.js
npm run dev
```

---

## Done Criteria

- [ ] `npm run dev` starts without errors
- [ ] `GET /login` renders the login page
- [ ] Login with seeded credentials → redirects to `/dashboard`
- [ ] `/dashboard` shows placeholder with user name in topbar
- [ ] Unauthenticated request to `/dashboard` → redirects to `/login`
- [ ] `GET /auth/logout` → destroys session, redirects to `/login`
- [ ] All 4 new DB tables exist in the database
