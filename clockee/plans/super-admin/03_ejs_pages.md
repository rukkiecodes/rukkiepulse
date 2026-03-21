# Phase 3 — EJS Pages

## Goal
Build all the EJS page views and their corresponding GET routes: dashboard stats, institutions list, institution detail, create-admin form, and audit log.

**Prerequisite:** Phase 2 complete (GraphQL resolvers working, `platformStats` + `institutions` queries return data).

---

## Scope

### Files to Create / Replace

```
super-admin/src/views/
├── layouts/
│   ├── main.ejs          ← full sidebar + topbar layout (replace placeholder)
│   └── auth.ejs          ← clean auth shell (already done in Phase 1)
├── partials/
│   ├── sidebar.ejs       ← nav links (replace placeholder)
│   ├── topbar.ejs        ← user name + logout (replace placeholder)
│   └── flash.ejs         ← success/error banner (replace placeholder)
└── pages/
    ├── dashboard.ejs     ← replace placeholder with full stats page
    ├── institutions/
    │   ├── list.ejs      ← new
    │   ├── detail.ejs    ← new
    │   └── create-admin.ejs ← new
    └── audit/
        └── log.ejs       ← new

super-admin/src/routes/index.js   ← add all new GET routes
```

---

## Layout Design (Tailwind CDN)

All pages use Tailwind CDN in the `<head>`:
```html
<script src="https://cdn.tailwindcss.com"></script>
```

### main.ejs layout structure
```
┌──────────────────────────────────────────────────────┐
│ Sidebar (fixed, 240px)  │  Main content area          │
│  Logo: Clockee SA       │  <%- topbar %>              │
│  ─────────────          │  ─────────────────────────  │
│  Dashboard              │  <%- body %>                │
│  Institutions           │                             │
│  Support                │                             │
│  Audit Log              │                             │
│  ─────────────          │                             │
│  [Logout]               │                             │
└──────────────────────────────────────────────────────┘
```

Sidebar active link is highlighted by comparing `req.path` to the current route (pass `currentPath` local to all pages).

---

## Page: /dashboard

### Data
The route fetches `platformStats` and `recentActivity` via internal GraphQL call (or direct service call — pick one pattern and use it consistently).

**Recommended pattern:** Call services directly in the route handler (no HTTP overhead):

```js
// src/routes/index.js
const InstitutionService = require('../services/InstitutionService');
const AuditService = require('../services/AuditService');
// ...
router.get('/dashboard', requireSuperAdmin, async (req, res) => {
  const [stats, activity] = await Promise.all([
    DashboardService.getStats(),
    AuditService.getPage({ limit: 20, page: 1 }),
  ]);
  res.render('pages/dashboard', { user: req.session.user, stats, activity: activity.items, currentPath: '/dashboard' });
});
```

### Layout
```
┌─────────────────────────────────────────────────────────┐
│ Page title: "Dashboard"                                  │
├──────────┬──────────┬──────────┬────────────────────────┤
│  Institutions        │  Users   │  Attendance Today       │
│  XX total            │  XX      │  XX clock-ins           │
│  XX active           │          │                         │
│  XX suspended        │          │                         │
├──────────┴──────────┴──────────┴────────────────────────┤
│  Open Tickets  │  Expiring Subscriptions (7d)            │
├─────────────────────────────────────────────────────────┤
│  Recent Activity Feed (last 20 audit entries)            │
│  [timestamp] [action] [target] [performed by]            │
└─────────────────────────────────────────────────────────┘
```

Stat cards: white bg, rounded-lg, shadow-sm, p-6. Value in large bold text, label in gray below.

---

## Page: /institutions

### Data
```js
router.get('/institutions', requireSuperAdmin, async (req, res) => {
  const { status, search, page = 1 } = req.query;
  const result = await InstitutionService.getPage({ status, search, page: parseInt(page), limit: 20 });
  res.render('pages/institutions/list', {
    user: req.session.user,
    institutions: result.items,
    total: result.total,
    pages: result.pages,
    currentPage: result.page,
    filters: { status, search },
    currentPath: '/institutions',
  });
});
```

### Layout
- Filter bar at top: search input + status dropdown + "Apply" button (form GET)
- Table columns: Name | Type | City | Subscription Status | Staff Count | Created | Actions
- Actions: "View" (→ `/institutions/:id`) | "Suspend" (POST form) | "Delete" (POST form with confirmation)
- Pagination: prev/next links with `?page=N`
- Status badges: color-coded pills (active=green, suspended=red, trial=yellow, expired=gray)

---

## Page: /institutions/:id

### Data
```js
router.get('/institutions/:id', requireSuperAdmin, async (req, res) => {
  const [institution, notes, users, tickets] = await Promise.all([
    InstitutionService.getById(req.params.id),
    InstitutionService.getNotes(req.params.id),
    UserService.getByInstitution(req.params.id),
    SupportService.getByInstitution(req.params.id),
  ]);
  if (!institution) return res.status(404).render('pages/404');
  res.render('pages/institutions/detail', {
    user: req.session.user,
    institution, notes, users, tickets,
    tab: req.query.tab || 'overview',
    currentPath: '/institutions',
  });
});
```

### Layout — Tabbed Interface
Tabs: Overview | Users | Support | Notes | Actions

**Overview tab:**
- Institution name, type, city, country
- Subscription status + end date
- GPS coordinates (if set)
- Staff count

**Users tab:**
- Table: Name | Email | Role | Status | Actions (Suspend / Reinstate)
- POST forms for suspend/reinstate

**Support tab:**
- List of tickets opened by this institution
- Link to each ticket thread

**Notes tab:**
- List of existing notes (newest first)
- Textarea form to add a new note

**Actions tab:**
- "Create Admin" button → `/institutions/:id/create-admin`
- "Suspend Institution" button (with reason textarea, POST form)
- "Reinstate Institution" button (POST form)
- "Control Dashboard" button (impersonation — wired in Phase 5)

---

## Page: /institutions/:id/create-admin

### Data
```js
router.get('/institutions/:id/create-admin', requireSuperAdmin, async (req, res) => {
  const institution = await InstitutionService.getById(req.params.id);
  res.render('pages/institutions/create-admin', { user: req.session.user, institution, error: null, currentPath: '/institutions' });
});

router.post('/institutions/:id/create-admin', requireSuperAdmin, async (req, res) => {
  const { name, email, password, phone } = req.body;
  try {
    await UserService.createAdmin({ institutionId: req.params.id, name, email, password, phone });
    await AuditService.log({ performedBy: req.session.user.id, action: 'CREATE_ADMIN',
      targetType: 'institution', targetId: req.params.id,
      details: { email }, ipAddress: req.ip });
    res.redirect(`/institutions/${req.params.id}?tab=users&success=Admin+created`);
  } catch (err) {
    const institution = await InstitutionService.getById(req.params.id);
    res.render('pages/institutions/create-admin', { user: req.session.user, institution, error: err.message, currentPath: '/institutions' });
  }
});
```

### Layout
- Form fields: Name, Email, Password, Phone (optional)
- Submit → POST to same URL
- Cancel → back to institution detail

---

## Page: /audit

### Data
```js
router.get('/audit', requireSuperAdmin, async (req, res) => {
  const { action, targetType, fromDate, toDate, page = 1 } = req.query;
  const result = await AuditService.getPage({ action, targetType, fromDate, toDate, page: parseInt(page), limit: 50 });
  res.render('pages/audit/log', {
    user: req.session.user,
    entries: result.items,
    total: result.total,
    pages: result.pages,
    currentPage: result.page,
    filters: { action, targetType, fromDate, toDate },
    currentPath: '/audit',
  });
});
```

### Layout
- Filter bar: action type select + target type select + date range inputs
- Table columns: Timestamp | Action | Target Type | Target ID | Performed By | IP | Details
- Details column: JSON preview (truncated, expand on click)
- Pagination
- Action badges: color-coded by type (SUSPEND=red, CREATE=green, REINSTATE=blue, etc.)

---

## POST Routes for Institution Actions

Add to `src/routes/index.js`:

```js
router.post('/institutions/:id/suspend', requireSuperAdmin, async (req, res) => {
  const { reason } = req.body;
  await InstitutionService.suspend(req.params.id, reason);
  await AuditService.log({ performedBy: req.session.user.id, action: 'SUSPEND_INSTITUTION',
    targetType: 'institution', targetId: parseInt(req.params.id),
    details: { reason }, ipAddress: req.ip });
  res.redirect(`/institutions/${req.params.id}?success=Institution+suspended`);
});

router.post('/institutions/:id/reinstate', requireSuperAdmin, async (req, res) => {
  await InstitutionService.reinstate(req.params.id);
  await AuditService.log({ performedBy: req.session.user.id, action: 'REINSTATE_INSTITUTION',
    targetType: 'institution', targetId: parseInt(req.params.id),
    details: {}, ipAddress: req.ip });
  res.redirect(`/institutions/${req.params.id}?success=Institution+reinstated`);
});

router.post('/institutions/:id/delete', requireSuperAdmin, async (req, res) => {
  const { reason } = req.body;
  await InstitutionService.softDelete(req.params.id, reason);
  await AuditService.log({ performedBy: req.session.user.id, action: 'DELETE_INSTITUTION',
    targetType: 'institution', targetId: parseInt(req.params.id),
    details: { reason }, ipAddress: req.ip });
  res.redirect('/institutions?success=Institution+deleted');
});

router.post('/institutions/:id/notes', requireSuperAdmin, async (req, res) => {
  const { note } = req.body;
  await InstitutionService.addNote(req.params.id, note, req.session.user.id);
  res.redirect(`/institutions/${req.params.id}?tab=notes`);
});

router.post('/users/:id/suspend', requireSuperAdmin, async (req, res) => {
  const { reason, institutionId } = req.body;
  await UserService.suspend(req.params.id, reason);
  await AuditService.log({ performedBy: req.session.user.id, action: 'SUSPEND_USER',
    targetType: 'user', targetId: parseInt(req.params.id),
    details: { reason }, ipAddress: req.ip });
  res.redirect(`/institutions/${institutionId}?tab=users`);
});

router.post('/users/:id/reinstate', requireSuperAdmin, async (req, res) => {
  const { institutionId } = req.body;
  await UserService.reinstate(req.params.id);
  await AuditService.log({ performedBy: req.session.user.id, action: 'REINSTATE_USER',
    targetType: 'user', targetId: parseInt(req.params.id),
    details: {}, ipAddress: req.ip });
  res.redirect(`/institutions/${institutionId}?tab=users`);
});
```

---

## Flash Messages

Pass success/error via query string on redirects (`?success=...` or `?error=...`) and render in `flash.ejs`:
```ejs
<% if (locals.query?.success) { %>
  <div class="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded mb-4">
    <%= decodeURIComponent(query.success) %>
  </div>
<% } %>
```

Or use `connect-flash` package for proper flash storage across redirects (simpler if added as a dependency).

---

## Done Criteria

- [ ] `/dashboard` shows real platform stats (not placeholder)
- [ ] `/institutions` shows paginated table with filters working
- [ ] `/institutions/:id` shows tabbed detail page with all tabs rendering
- [ ] `/institutions/:id/create-admin` form creates an admin user successfully
- [ ] Suspend / Reinstate / Delete institution POST routes work and redirect with success message
- [ ] Suspend / Reinstate user POST routes work
- [ ] All institution actions appear in `/audit` log
- [ ] `/audit` filter by action type and date range works
- [ ] Sidebar shows active link highlight based on current page
- [ ] Flash messages appear on redirect
