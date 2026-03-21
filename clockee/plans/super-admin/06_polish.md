# Phase 6 — Polish

## Goal
Complete the remaining UX details: proper flash messages, pagination everywhere, institution suspension emails, full audit log coverage, and final verification that every mutation is logged.

**Prerequisite:** Phases 1–5 complete (all features working end-to-end).

---

## Scope

### Areas Covered

1. Flash messages — replace query-string approach with `connect-flash`
2. Pagination — verify all list pages have working prev/next
3. Suspension email — send warning email when institution is suspended
4. Audit log completeness — verify every mutation writes to `platform_audit_log`
5. Input validation — add server-side validation to all POST forms
6. Error pages — add 404 and 500 error views
7. Production hardening — session cookie `secure`, CSP headers, rate limiting on login

---

## Implementation Steps

### Step 1 — Proper flash messages with connect-flash

Add `connect-flash` to `package.json`:
```json
"connect-flash": "^0.1.1"
```

Wire in `app.js` (after session middleware):
```js
const flash = require('connect-flash');
app.use(flash());

// Make flash messages available to all EJS templates
app.use((req, res, next) => {
  res.locals.successFlash = req.flash('success');
  res.locals.errorFlash = req.flash('error');
  next();
});
```

Update `src/views/partials/flash.ejs`:
```ejs
<% if (successFlash && successFlash.length) { %>
  <div class="bg-green-50 border border-green-200 text-green-800 px-4 py-3 rounded-lg mb-4 flex items-center gap-2">
    <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <polyline points="20 6 9 17 4 12"></polyline>
    </svg>
    <%= successFlash[0] %>
  </div>
<% } %>
<% if (errorFlash && errorFlash.length) { %>
  <div class="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-lg mb-4 flex items-center gap-2">
    <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <circle cx="12" cy="12" r="10"></circle>
      <line x1="12" y1="8" x2="12" y2="12"></line>
      <line x1="12" y1="16" x2="12.01" y2="16"></line>
    </svg>
    <%= errorFlash[0] %>
  </div>
<% } %>
```

Update all POST routes to use `req.flash('success', '...')` and `req.flash('error', '...')` instead of `?success=` and `?error=` query params.

### Step 2 — Pagination component (partial)

Create `src/views/partials/pagination.ejs`:
```ejs
<% if (pages > 1) { %>
  <div class="flex items-center justify-between mt-6">
    <p class="text-sm text-gray-500">
      Showing page <%= currentPage %> of <%= pages %> (<%= total %> total)
    </p>
    <div class="flex gap-2">
      <% if (currentPage > 1) { %>
        <a href="?<%= queryString %>&page=<%= currentPage - 1 %>"
           class="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50">
          Previous
        </a>
      <% } %>
      <% if (currentPage < pages) { %>
        <a href="?<%= queryString %>&page=<%= currentPage + 1 %>"
           class="px-3 py-1 border border-gray-300 rounded text-sm hover:bg-gray-50">
          Next
        </a>
      <% } %>
    </div>
  </div>
<% } %>
```

Pass `queryString` from each route (current filter params serialized, without `page`):
```js
const qs = new URLSearchParams({ status: status || '', search: search || '' }).toString();
res.render('...', { queryString: qs, currentPage, pages, total });
```

Include in each list page:
```ejs
<%- include('../../partials/pagination', { pages, currentPage, total, queryString }) %>
```

Verify pagination is present and working on:
- [ ] `/institutions` (20 per page)
- [ ] `/support` (25 per page)
- [ ] `/audit` (50 per page)

### Step 3 — Suspension email

When `suspendInstitution` is called, send an email to the institution's primary admin.

Add to `InstitutionService.suspend()`:
```js
static async suspend(institutionId, reason) {
  await pool.query(
    `UPDATE institutions SET subscription_status = 'suspended' WHERE id = $1`,
    [institutionId]
  );
  // Also suspend the primary admin user
  await pool.query(
    `UPDATE users SET status = 'suspended'
     WHERE institution_id = $1 AND role = 'admin'`,
    [institutionId]
  );
  // Send suspension email
  await EmailService.sendSuspensionNotice(institutionId, reason);
}
```

`EmailService.sendSuspensionNotice` reuses the existing email infrastructure from `server/utils/emailService.js` (import it or call the same SMTP/SendGrid config). Email template:

**Subject:** Your Clockee account has been suspended

**Body:**
> Hi [Admin Name],
>
> Your institution ([Institution Name]) has been suspended on the Clockee platform.
>
> Reason: [reason]
>
> If you believe this is a mistake, please contact support at support@clockee.app.
>
> — The Clockee Team

### Step 4 — Audit log completeness audit

Go through every POST route and GraphQL mutation and verify `AuditService.log()` is called.

| Action | Log entry required | Logged in |
|---|---|---|
| Login | No (session event, not auditable action) | — |
| Suspend institution | Yes | route handler |
| Reinstate institution | Yes | route handler |
| Delete institution | Yes | route handler |
| Update subscription | Yes | route handler |
| Add institution note | No (low-stakes) | — |
| Create institution admin | Yes | route handler |
| Suspend user | Yes | route handler |
| Reinstate user | Yes | route handler |
| Create support ticket | No | — |
| Send support message | No | — |
| Update ticket status | Yes | resolver |
| Generate impersonation token | Yes | resolver |

Add any missing `AuditService.log()` calls.

### Step 5 — Server-side input validation

Add validation to critical POST routes (don't trust the browser):

**Suspend institution** — `reason` must be a non-empty string, max 500 chars.
**Delete institution** — `reason` must be a non-empty string.
**Create admin** — `name`, `email`, `password` required; email must be valid format; password min 8 chars.
**Send support message** — `message` non-empty, max 5000 chars.

Simple inline validation pattern:
```js
router.post('/institutions/:id/suspend', requireSuperAdmin, async (req, res) => {
  const { reason } = req.body;
  if (!reason || reason.trim().length < 5) {
    req.flash('error', 'A suspension reason of at least 5 characters is required.');
    return res.redirect(`/institutions/${req.params.id}?tab=actions`);
  }
  // ...proceed
});
```

### Step 6 — Error pages

Create `src/views/pages/error.ejs`:
```ejs
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title><%= code %> — Clockee Super Admin</title>
  <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50 flex items-center justify-center min-h-screen">
  <div class="text-center">
    <p class="text-6xl font-bold text-gray-200"><%= code %></p>
    <h1 class="text-xl font-semibold text-gray-800 mt-2"><%= message %></h1>
    <a href="/dashboard" class="mt-4 inline-block text-blue-600 hover:underline text-sm">← Back to Dashboard</a>
  </div>
</body>
</html>
```

Add to `app.js` after routes:
```js
// 404
app.use((req, res) => {
  res.status(404).render('pages/error', { code: 404, message: 'Page not found.' });
});

// 500
app.use((err, req, res, next) => {
  console.error('[super-admin] unhandled error', err);
  res.status(500).render('pages/error', { code: 500, message: 'Something went wrong.' });
});
```

### Step 7 — Login rate limiting

```js
const rateLimit = require('express-rate-limit');

const loginLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 10,
  message: 'Too many login attempts. Please try again in 15 minutes.',
});

// Apply only to POST /auth/login
app.use('/auth/login', loginLimiter);
```

Add `express-rate-limit` to `package.json`.

### Step 8 — Production hardening checklist

Before deploying to production:

- [ ] `NODE_ENV=production` in `.env`
- [ ] Session cookie `secure: true` (already conditional on `NODE_ENV === 'production'`)
- [ ] `SESSION_SECRET` is a long random string (not a placeholder)
- [ ] `IMPERSONATION_SECRET` is different from `JWT_SECRET` and `SESSION_SECRET`
- [ ] `SUPER_ADMIN_PASSWORD` is strong (min 16 chars, mixed case, numbers, symbols)
- [ ] Tailwind CDN → acceptable for internal admin tool; note it's not ideal for production performance but acceptable
- [ ] Apollo Server `introspection: false` in production (prevents schema exposure):
  ```js
  new ApolloServer({ typeDefs, resolvers, introspection: process.env.NODE_ENV !== 'production' })
  ```
- [ ] `morgan` logging set to `'combined'` format in production
- [ ] Socket.io CORS origin restricted to `ADMIN_PANEL_URL` in production

---

## Done Criteria

- [ ] Flash messages (success and error) display correctly after all POST actions — no query-param leakage in URL
- [ ] Pagination works on institutions, support, and audit pages — page count is correct, filters preserved across pages
- [ ] Suspending an institution sends an email to the institution's primary admin
- [ ] Every auditable action (see table in Step 4) creates a `platform_audit_log` entry
- [ ] Form validation prevents empty/invalid submissions with a clear error message
- [ ] Accessing a non-existent `/institutions/99999` renders a 404 page
- [ ] An unhandled server error renders the 500 error page instead of crashing
- [ ] Login rate limiting rejects after 10 attempts within 15 minutes
- [ ] Apollo Server introspection is disabled in production
