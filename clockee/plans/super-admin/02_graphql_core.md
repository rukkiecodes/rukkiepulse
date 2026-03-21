# Phase 2 — GraphQL Core

## Goal
Wire up Apollo Server 4 into the Express app and implement all GraphQL schemas and resolvers for auth, dashboard stats, institutions, users, and audit log.

**Prerequisite:** Phase 1 complete (Express app running, DB migrated, session auth working).

---

## Scope

### Files to Create

```
super-admin/src/graphql/
├── schema/
│   ├── index.js            ← mergeTypeDefs from all .graphql files
│   ├── auth.graphql
│   ├── dashboard.graphql
│   ├── institution.graphql
│   ├── user.graphql
│   └── audit.graphql
├── resolvers/
│   ├── index.js            ← mergeResolvers from all resolver files
│   ├── authResolvers.js
│   ├── dashboardResolvers.js
│   ├── institutionResolvers.js
│   ├── userResolvers.js
│   └── auditResolvers.js
└── context.js

super-admin/src/services/
├── AuthService.js
├── InstitutionService.js
└── AuditService.js

super-admin/src/middleware/
└── auditLogger.js
```

### Files to Modify

```
super-admin/app.js   ← add Apollo Server middleware at /graphql
```

---

## Implementation Steps

### Step 1 — GraphQL context (src/graphql/context.js)

```js
module.exports = async ({ req }) => {
  const user = req.session?.user || null;
  return {
    user,
    req,
    assertSuperAdmin() {
      if (!user || user.role !== 'super_admin') {
        throw new Error('UNAUTHORIZED');
      }
    },
  };
};
```

### Step 2 — Schema files

**auth.graphql**
```graphql
type SuperAdminProfile {
  id: ID!
  email: String!
  name: String!
  role: String!
}

type AuthPayload {
  success: Boolean!
  user: SuperAdminProfile
}

type Query {
  me: SuperAdminProfile
}

type Mutation {
  login(email: String!, password: String!): AuthPayload
}
```

**dashboard.graphql**
```graphql
type PlatformStats {
  totalInstitutions: Int!
  activeInstitutions: Int!
  suspendedInstitutions: Int!
  trialInstitutions: Int!
  totalUsers: Int!
  totalAttendanceToday: Int!
  openSupportTickets: Int!
  expiringSubscriptions: Int!
}

type ActivityEntry {
  id: ID!
  action: String!
  targetType: String
  targetId: Int
  details: String
  ipAddress: String
  performedBy: String
  createdAt: String!
}

extend type Query {
  platformStats: PlatformStats
  recentActivity(limit: Int): [ActivityEntry]
}
```

**institution.graphql**
```graphql
type Institution {
  id: ID!
  name: String!
  type: String
  city: String
  country: String
  subscriptionStatus: String
  subscriptionEndDate: String
  staffCount: Int
  createdAt: String!
}

type InstitutionPage {
  items: [Institution!]!
  total: Int!
  page: Int!
  pages: Int!
}

type InstitutionNote {
  id: ID!
  note: String!
  createdBy: String
  createdAt: String!
}

input InstitutionFilter {
  status: String
  subscriptionType: String
  country: String
  search: String
}

input SubscriptionInput {
  status: String
  endDate: String
  plan: String
}

extend type Query {
  institutions(filter: InstitutionFilter, page: Int, limit: Int): InstitutionPage
  institution(id: ID!): Institution
  institutionNotes(institutionId: ID!): [InstitutionNote]
}

extend type Mutation {
  suspendInstitution(id: ID!, reason: String!): Institution
  reinstateInstitution(id: ID!): Institution
  deleteInstitution(id: ID!, reason: String!): Boolean
  updateInstitutionSubscription(id: ID!, data: SubscriptionInput!): Institution
  addInstitutionNote(institutionId: ID!, note: String!): InstitutionNote
}
```

**user.graphql**
```graphql
type User {
  id: ID!
  name: String!
  email: String!
  role: String!
  status: String!
  createdAt: String!
}

input CreateAdminInput {
  name: String!
  email: String!
  password: String!
  phone: String
}

extend type Query {
  institutionUsers(institutionId: ID!, role: String): [User]
}

extend type Mutation {
  createInstitutionAdmin(institutionId: ID!, data: CreateAdminInput!): User
  suspendUser(userId: ID!, reason: String!): User
  reinstateUser(userId: ID!): User
}
```

**audit.graphql**
```graphql
type AuditEntry {
  id: ID!
  action: String!
  targetType: String
  targetId: Int
  details: String
  ipAddress: String
  performedBy: String
  createdAt: String!
}

type AuditPage {
  items: [AuditEntry!]!
  total: Int!
  page: Int!
  pages: Int!
}

input AuditFilter {
  action: String
  targetType: String
  fromDate: String
  toDate: String
}

extend type Query {
  platformAuditLog(filter: AuditFilter, page: Int, limit: Int): AuditPage
}
```

### Step 3 — Schema index (src/graphql/schema/index.js)

```js
const { mergeTypeDefs } = require('@graphql-tools/merge');
const { loadFilesSync } = require('@graphql-tools/load-files');
const path = require('path');

const typesArray = loadFilesSync(path.join(__dirname, '.'), { extensions: ['graphql'] });
module.exports = mergeTypeDefs(typesArray);
```

> Add `@graphql-tools/load-files` to package.json, or load each file manually with `fs.readFileSync` and `gql` tag.

### Step 4 — Services

**src/services/AuditService.js**
```js
const pool = require('../config/database');

class AuditService {
  static async log({ performedBy, action, targetType, targetId, details, ipAddress }) {
    await pool.query(
      `INSERT INTO platform_audit_log (performed_by, action, target_type, target_id, details, ip_address)
       VALUES ($1, $2, $3, $4, $5, $6)`,
      [performedBy, action, targetType || null, targetId || null, JSON.stringify(details || {}), ipAddress || null]
    );
  }

  static async getPage({ action, targetType, fromDate, toDate, page = 1, limit = 50 }) {
    const offset = (page - 1) * limit;
    const conditions = [];
    const params = [];

    if (action) { params.push(action); conditions.push(`pal.action = $${params.length}`); }
    if (targetType) { params.push(targetType); conditions.push(`pal.target_type = $${params.length}`); }
    if (fromDate) { params.push(fromDate); conditions.push(`pal.created_at >= $${params.length}`); }
    if (toDate) { params.push(toDate); conditions.push(`pal.created_at <= $${params.length}`); }

    const where = conditions.length ? `WHERE ${conditions.join(' AND ')}` : '';

    const countRes = await pool.query(`SELECT COUNT(*) FROM platform_audit_log pal ${where}`, params);
    const total = parseInt(countRes.rows[0].count, 10);

    params.push(limit, offset);
    const rows = await pool.query(`
      SELECT pal.*, u.name AS performed_by_name
      FROM platform_audit_log pal
      LEFT JOIN users u ON u.id = pal.performed_by
      ${where}
      ORDER BY pal.created_at DESC
      LIMIT $${params.length - 1} OFFSET $${params.length}
    `, params);

    return { items: rows.rows, total, page, pages: Math.ceil(total / limit) };
  }
}

module.exports = AuditService;
```

**src/services/InstitutionService.js** — methods: `getPage`, `getById`, `suspend`, `reinstate`, `softDelete`, `updateSubscription`, `getNotes`, `addNote`.

**src/services/AuthService.js** — `login(email, password)` (bcrypt compare, return user row or null).

### Step 5 — Resolvers

**authResolvers.js**
```js
const AuthService = require('../../services/AuthService');

module.exports = {
  Query: {
    me: (_, __, { user }) => user,
  },
  Mutation: {
    login: async (_, { email, password }, { req }) => {
      const user = await AuthService.login(email, password);
      if (!user) return { success: false, user: null };
      req.session.user = { id: user.id, email: user.email, name: user.name, role: user.role };
      return { success: true, user };
    },
  },
};
```

**dashboardResolvers.js** — queries `institutions`, `users`, `attendance_records`, `support_tickets`, `platform_audit_log` for counts.

**institutionResolvers.js** — calls `InstitutionService` methods, logs each mutation via `AuditService.log`.

**userResolvers.js** — `createInstitutionAdmin` (hash password, insert into `users` with role `admin`), `suspendUser`, `reinstateUser`.

**auditResolvers.js** — calls `AuditService.getPage`.

### Step 6 — Resolver index (src/graphql/resolvers/index.js)

```js
const { mergeResolvers } = require('@graphql-tools/merge');
const authResolvers = require('./authResolvers');
const dashboardResolvers = require('./dashboardResolvers');
const institutionResolvers = require('./institutionResolvers');
const userResolvers = require('./userResolvers');
const auditResolvers = require('./auditResolvers');

module.exports = mergeResolvers([
  authResolvers, dashboardResolvers, institutionResolvers, userResolvers, auditResolvers,
]);
```

### Step 7 — Wire Apollo into app.js

```js
const { ApolloServer } = require('@apollo/server');
const { expressMiddleware } = require('@apollo/server/express4');
const typeDefs = require('./src/graphql/schema/index');
const resolvers = require('./src/graphql/resolvers/index');
const buildContext = require('./src/graphql/context');

// After session middleware, before page routes:
const apolloServer = new ApolloServer({ typeDefs, resolvers });
await apolloServer.start();
app.use('/graphql', expressMiddleware(apolloServer, { context: buildContext }));
```

> Note: `apolloServer.start()` is async. Wrap app setup in an async function or use top-level await.

### Step 8 — src/public/js/graphql-client.js

```js
async function gql(query, variables = {}) {
  const res = await fetch('/graphql', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ query, variables }),
  });
  const json = await res.json();
  if (json.errors) throw new Error(json.errors[0].message);
  return json.data;
}
```

---

## Done Criteria

- [ ] `GET /graphql` (sandbox) is accessible in dev mode
- [ ] `me` query returns `null` when not logged in, returns user object when logged in
- [ ] `platformStats` query returns correct counts from the DB
- [ ] `institutions` query returns paginated list
- [ ] `suspendInstitution` mutation sets `subscription_status = 'suspended'` and logs to `platform_audit_log`
- [ ] `reinstateInstitution` mutation sets `subscription_status = 'active'` and logs
- [ ] `createInstitutionAdmin` mutation inserts a new user with role `admin` for the institution
- [ ] `platformAuditLog` query returns paginated log entries
- [ ] All mutations call `context.assertSuperAdmin()` and throw `UNAUTHORIZED` if not super admin
