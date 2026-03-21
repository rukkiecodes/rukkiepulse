# Phase 5 — Impersonation

## Goal
Allow the super admin to log into any institution's admin panel as that institution's primary admin, using a short-lived JWT passed via URL. No password needed. The existing admin panel (`admin/`) must verify and accept the token.

**Prerequisite:** Phase 3 complete (institution detail page exists with an Actions tab).

---

## Scope

### Files to Create

```
super-admin/src/services/
└── ImpersonationService.js

server/controllers/auth/
└── ImpersonationController.js
```

### Files to Modify

```
super-admin/src/graphql/schema/  ← add generateImpersonationToken to institution.graphql (or new impersonation.graphql)
super-admin/src/graphql/resolvers/ ← add to institutionResolvers.js or new impersonationResolvers.js
super-admin/src/views/pages/institutions/detail.ejs  ← wire "Control Dashboard" button
server/routes/auth.js             ← add GET /api/auth/impersonate
```

---

## Implementation Steps

### Step 1 — Shared secret

Both `super-admin/.env` and `server/.env` must have the same value:
```env
IMPERSONATION_SECRET=a-long-random-secret-shared-between-both-apps
```

### Step 2 — ImpersonationService.js (super-admin)

```js
const jwt = require('jsonwebtoken');
const pool = require('../config/database');

class ImpersonationService {
  /**
   * Generate a 15-minute JWT for logging into an institution's admin panel.
   * @param {number} institutionId
   * @param {number} superAdminId - the super admin performing the action
   * @returns {string} signed JWT
   */
  static async generateToken(institutionId, superAdminId) {
    // Find the primary admin for this institution
    const res = await pool.query(
      `SELECT id, email, name FROM users
       WHERE institution_id = $1 AND role = 'admin' AND status = 'approved'
       ORDER BY created_at ASC LIMIT 1`,
      [institutionId]
    );
    const admin = res.rows[0];
    if (!admin) throw new Error('No active admin found for this institution');

    const payload = {
      userId: admin.id,
      institutionId: parseInt(institutionId),
      impersonatedBy: superAdminId,
      type: 'impersonation',
    };

    return jwt.sign(payload, process.env.IMPERSONATION_SECRET, { expiresIn: '15m' });
  }
}

module.exports = ImpersonationService;
```

### Step 3 — GraphQL: generateImpersonationToken mutation

Add to `super-admin/src/graphql/schema/institution.graphql`:
```graphql
type ImpersonationPayload {
  token: String!
  adminPanelUrl: String!
}

extend type Mutation {
  generateImpersonationToken(institutionId: ID!): ImpersonationPayload
}
```

Add to `institutionResolvers.js`:
```js
const ImpersonationService = require('../../services/ImpersonationService');
const AuditService = require('../../services/AuditService');

// Inside Mutation:
generateImpersonationToken: async (_, { institutionId }, { user, assertSuperAdmin }) => {
  assertSuperAdmin();
  const token = await ImpersonationService.generateToken(institutionId, user.id);
  await AuditService.log({
    performedBy: user.id,
    action: 'GENERATE_IMPERSONATION_TOKEN',
    targetType: 'institution',
    targetId: parseInt(institutionId),
    details: { note: '15-minute impersonation token generated' },
  });
  const adminPanelUrl = `${process.env.ADMIN_PANEL_URL}/auth/impersonate?token=${token}`;
  return { token, adminPanelUrl };
},
```

### Step 4 — "Control Dashboard" button in detail.ejs

In the **Actions tab** of `src/views/pages/institutions/detail.ejs`:

```html
<div class="bg-white border border-gray-200 rounded-lg p-6">
  <h3 class="text-sm font-semibold text-gray-700 mb-1">Control Dashboard</h3>
  <p class="text-sm text-gray-500 mb-4">
    Log into this institution's admin panel as their primary admin.
    The session token expires in 15 minutes.
  </p>
  <button
    id="impersonateBtn"
    data-institution-id="<%= institution.id %>"
    class="bg-purple-600 hover:bg-purple-700 text-white px-4 py-2 rounded-lg text-sm font-medium transition-colors"
  >
    Control Dashboard
  </button>
  <p id="impersonateError" class="text-red-600 text-sm mt-2 hidden"></p>
</div>

<script src="/js/graphql-client.js"></script>
<script>
  document.getElementById('impersonateBtn').addEventListener('click', async function () {
    const institutionId = this.dataset.institutionId;
    const errorEl = document.getElementById('impersonateError');
    errorEl.classList.add('hidden');
    this.disabled = true;
    this.textContent = 'Generating token...';

    try {
      const data = await gql(`
        mutation GenerateToken($id: ID!) {
          generateImpersonationToken(institutionId: $id) {
            adminPanelUrl
          }
        }
      `, { id: institutionId });

      window.open(data.generateImpersonationToken.adminPanelUrl, '_blank');
    } catch (err) {
      errorEl.textContent = err.message || 'Failed to generate token.';
      errorEl.classList.remove('hidden');
    } finally {
      this.disabled = false;
      this.textContent = 'Control Dashboard';
    }
  });
</script>
```

### Step 5 — ImpersonationController.js (existing server/)

File: `server/controllers/auth/ImpersonationController.js`

```js
const jwt = require('jsonwebtoken');
const pool = require('../../config/database');

class ImpersonationController {
  static async verify(req, res) {
    const { token } = req.query;
    if (!token) {
      return res.status(400).json({ error: 'Missing impersonation token.' });
    }

    let payload;
    try {
      payload = jwt.verify(token, process.env.IMPERSONATION_SECRET);
    } catch (err) {
      return res.status(401).json({ error: 'Invalid or expired impersonation token.' });
    }

    if (payload.type !== 'impersonation') {
      return res.status(401).json({ error: 'Invalid token type.' });
    }

    // Look up the user
    const result = await pool.query(
      'SELECT id, email, name, role, institution_id FROM users WHERE id = $1 AND status = $2',
      [payload.userId, 'approved']
    );
    const user = result.rows[0];
    if (!user) {
      return res.status(404).json({ error: 'User not found or not approved.' });
    }

    // This endpoint is called by the Vue admin panel via a redirect.
    // It returns a JWT that the admin panel uses to authenticate.
    // The admin panel's existing login flow issues a JWT — replicate that here.
    const authToken = jwt.sign(
      { userId: user.id, institutionId: user.institution_id, role: user.role, impersonated: true },
      process.env.JWT_SECRET,
      { expiresIn: '8h' }
    );

    // Return the token so the Vue app can store it and redirect to dashboard.
    // The Vue app's /auth/impersonate route handles this redirect.
    return res.json({ token: authToken, user: { id: user.id, name: user.name, email: user.email, role: user.role } });
  }
}

module.exports = ImpersonationController;
```

### Step 6 — Add route to server/routes/auth.js

```js
const ImpersonationController = require('../controllers/auth/ImpersonationController');

// Existing routes above...

// GET /api/auth/impersonate?token=xxx
// Called by the Vue admin panel when the super admin redirects there
router.get('/impersonate', ImpersonationController.verify);
```

### Step 7 — Vue admin panel: /auth/impersonate route

The Vue admin panel needs a new route that:
1. Reads `?token=xxx` from the URL
2. Calls `GET /api/auth/impersonate?token=xxx`
3. Gets back a JWT + user object
4. Stores them the same way as normal login (in the auth store)
5. Redirects to `/dashboard`

This is a small addition to the Vue router and the auth store. Handled in the admin panel separately — outside the super-admin project.

**File:** `admin/src/pages/auth/impersonate.vue` (new page)
**Router entry:** `{ path: '/auth/impersonate', component: () => import('./pages/auth/impersonate.vue'), meta: { requiresAuth: false } }`

---

## Security Notes

- The `IMPERSONATION_SECRET` is a separate secret from `JWT_SECRET`. It is **only** used for impersonation tokens.
- The impersonation JWT expires in **15 minutes** — after that the link is dead.
- The `ImpersonationController` verifies `payload.type === 'impersonation'` to prevent using normal user JWTs.
- Every token generation is logged to `platform_audit_log` with `GENERATE_IMPERSONATION_TOKEN`.
- The resulting admin session token is marked with `impersonated: true` in its payload for auditing.
- The admin panel can show a banner "You are logged in as [institution] via super admin" if `impersonated === true`.

---

## Done Criteria

- [ ] "Control Dashboard" button on institution detail page generates a token without errors
- [ ] Clicking the button opens a new tab pointing to `ADMIN_PANEL_URL/auth/impersonate?token=xxx`
- [ ] The Vue admin panel's `/auth/impersonate` page calls the backend, gets a JWT, stores it, and redirects to dashboard
- [ ] The resulting admin session has the correct institution's data visible
- [ ] An expired token (>15 min) returns an error from the backend
- [ ] A reused or tampered token is rejected
- [ ] The action is logged in `platform_audit_log` with action `GENERATE_IMPERSONATION_TOKEN`
- [ ] `IMPERSONATION_SECRET` is set in both `super-admin/.env` and `server/.env`
