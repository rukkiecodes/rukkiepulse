# Clockee Microservices Plan

## Overview

All microservices are Node.js + GraphQL backends.
Download/Deletion services also include an EJS web UI (same pattern as super-admin panel).
Student Attendance service exposes GraphQL only (consumed by admin Vue panel).

---

## Project Layout (new directories to create)

```
clockee/
├── students-attendance-service/   ← Student attendance GraphQL microservice
├── data-portal/                   ← Download + Deletion microservice (EJS + GraphQL)
│     ├── src/
│     │   ├── views/               ← EJS templates (staff/parent/admin portals)
│     │   ├── routes/              ← Express routes (GraphQL endpoint + EJS pages)
│     │   ├── graphql/             ← schema.js + resolvers/
│     │   ├── services/            ← PDF generation, deletion queue
│     │   └── middleware/
│     └── package.json
└── legal/                         ← Static HTML (Privacy Policy, Data Handling)
```

---

## DB Tables Required (new migrations)

### 1. `student_attendance` — dedicated table (microservice owns it)
```sql
CREATE TABLE student_attendance (
  id              SERIAL PRIMARY KEY,
  student_id      INTEGER NOT NULL REFERENCES children(id),
  institution_id  INTEGER NOT NULL REFERENCES institutions(id),
  checked_in_by   INTEGER NOT NULL REFERENCES users(id),  -- staff/admin
  checked_out_by  INTEGER REFERENCES users(id),
  date            DATE NOT NULL DEFAULT CURRENT_DATE,
  check_in_time   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  check_out_time  TIMESTAMPTZ,
  status          VARCHAR(20) NOT NULL DEFAULT 'present',  -- present|absent|late
  notes           TEXT,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_sa_student_id       ON student_attendance(student_id);
CREATE INDEX idx_sa_institution_id   ON student_attendance(institution_id);
CREATE INDEX idx_sa_date             ON student_attendance(date);
```

### 2. `data_deletion_requests` — deletion queue (30-day soft delete)
```sql
CREATE TABLE data_deletion_requests (
  id              SERIAL PRIMARY KEY,
  user_id         INTEGER NOT NULL REFERENCES users(id),
  institution_id  INTEGER REFERENCES institutions(id),
  role            VARCHAR(20) NOT NULL,  -- staff|parent|admin
  requested_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  scheduled_delete_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '30 days'),
  status          VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending|cancelled|completed
  cancelled_at    TIMESTAMPTZ,
  completed_at    TIMESTAMPTZ,
  reason          TEXT
);
```

---

## Environment Variables

### students-attendance-service/.env
```
PORT=4001
DATABASE_URL=<same as main server>
JWT_SECRET=<same as main server>
MAIN_API_URL=http://localhost:5000
```

### data-portal/.env
```
PORT=4002
DATABASE_URL=<same as main server>
JWT_SECRET=<same as main server>
SESSION_SECRET=<random>
PUPPETEER_SKIP_CHROMIUM_DOWNLOAD=true  # if using preinstalled chrome
```

---

## BATCH 1 — Student Attendance Microservice (8 items)

**Goal:** Standalone Node.js + GraphQL service. Staff/admin check students in or out by
scanning a QR code or entering the student's passcode. Admin panel consumes via Apollo Client.

### Item 1 — Scaffold `students-attendance-service/`
```
students-attendance-service/
├── src/
│   ├── config/
│   │   └── database.js         ← pg Pool (same DB)
│   ├── graphql/
│   │   ├── schema.js           ← type definitions
│   │   └── resolvers/
│   │       ├── index.js
│   │       ├── query.js
│   │       └── mutation.js
│   ├── middleware/
│   │   └── auth.js             ← JWT verify (same secret)
│   └── app.js
├── package.json
└── .env.example
```

**package.json deps:**
- `express`, `express-graphql` or `apollo-server-express`
- `graphql`
- `pg`
- `jsonwebtoken`
- `dotenv`
- `cors`

### Item 2 — Database config + migration script
`src/config/database.js` — `pg.Pool` using `DATABASE_URL` from `.env`
`scripts/migrate.js` — creates `student_attendance` table if not exists

### Item 3 — Auth middleware
`src/middleware/auth.js`:
- Read `Authorization: Bearer <token>` header
- Verify with `JWT_SECRET`
- Attach `req.user = { id, role, institution_id, name }`
- Only allow `role IN ('admin', 'staff')`

### Item 4 — GraphQL Schema
```graphql
type Student {
  id: ID!
  name: String!
  studentId: String
  gradeLevel: String
  classSection: String
  photoUrl: String
  status: String
}

type AttendanceRecord {
  id: ID!
  student: Student!
  date: String!
  checkInTime: String
  checkOutTime: String
  status: String!          # present | absent | late
  checkedInBy: String
  checkedOutBy: String
  notes: String
}

type TodayStats {
  totalStudents: Int!
  checkedIn: Int!
  checkedOut: Int!
  absent: Int!
  stillInside: Int!        # checked in but not yet checked out
}

type RecentScan {
  student: Student!
  action: String!          # check_in | check_out
  time: String!
  recordedBy: String!
}

type Query {
  students(institutionId: ID!): [Student!]!
  studentByPasscode(passcode: String!): Student
  attendanceHistory(institutionId: ID!, date: String, studentId: ID, page: Int, limit: Int): AttendanceHistoryResult!
  todayStats(institutionId: ID!): TodayStats!
  recentScans(institutionId: ID!, limit: Int): [RecentScan!]!
  studentAttendanceRecord(studentId: ID!, date: String!): AttendanceRecord
}

type AttendanceHistoryResult {
  records: [AttendanceRecord!]!
  total: Int!
  page: Int!
  totalPages: Int!
}

type MutationResult {
  success: Boolean!
  message: String!
  record: AttendanceRecord
  student: Student
}

type Mutation {
  checkInStudent(passcode: String!, notes: String): MutationResult!
  checkOutStudent(passcode: String!, notes: String): MutationResult!
  markAbsent(studentId: ID!, date: String, notes: String): MutationResult!
  undoAttendance(recordId: ID!): MutationResult!
}
```

### Item 5 — Query Resolvers
`src/graphql/resolvers/query.js`:
- `students(institutionId)` — query `children` table WHERE institution_id = $1 AND deleted_at IS NULL
- `studentByPasscode(passcode)` — query `users` WHERE passcode = $1 (children linked via student_id)
  - NOTE: passcode is stored on the `children` table (field: `passcode` or generated invite code)
- `attendanceHistory(institutionId, date?, studentId?, page, limit)` — query `student_attendance` with joins
- `todayStats(institutionId)` — aggregate counts for today from `student_attendance`
- `recentScans(institutionId, limit)` — last N records from `student_attendance` ordered by check_in_time DESC
- `studentAttendanceRecord(studentId, date)` — single record lookup

### Item 6 — Mutation Resolvers
`src/graphql/resolvers/mutation.js`:

**checkInStudent(passcode, notes):**
1. Look up student from `children` by passcode field (or via `users.passcode` → `users.student_id` → `children.student_id`)
2. Check no existing check-in today (in `student_attendance`)
3. Insert record: `{ student_id, institution_id, checked_in_by: req.user.id, date, check_in_time: NOW(), status: 'present' }`
4. Return student info + record

**checkOutStudent(passcode, notes):**
1. Look up student from passcode
2. Find today's open record (check_in_time IS NOT NULL AND check_out_time IS NULL)
3. Update: `SET check_out_time = NOW(), checked_out_by = req.user.id`
4. Return updated record

**markAbsent(studentId, date, notes):**
1. Insert/upsert record with `status = 'absent'`, `check_in_time = NULL`
2. Admin/staff only

**undoAttendance(recordId):**
1. Delete the attendance record (same-day only for safety)

### Item 7 — `app.js` (Express + Apollo)
```js
const express = require('express')
const { ApolloServer } = require('apollo-server-express')
const { typeDefs } = require('./graphql/schema')
const resolvers = require('./graphql/resolvers')
const { verifyToken } = require('./middleware/auth')

const app = express()
app.use(cors())

const server = new ApolloServer({
  typeDefs,
  resolvers,
  context: ({ req }) => {
    const user = verifyToken(req)  // throws if invalid/missing
    return { user, pool }
  },
})

server.applyMiddleware({ app, path: '/graphql' })
app.listen(process.env.PORT || 4001)
```

### Item 8 — DB migration file
`scripts/migrate.js` — creates `student_attendance` table.
Also includes a seeding check: reads existing `children.passcode` column to confirm
the passcode field exists (if not, logs a warning to run the main server migration).

---

## BATCH 2 — Admin Panel GraphQL Integration (8 items)

**Goal:** Admin Vue panel consumes the student attendance GraphQL API.
Staff/admin can check students in/out and view reports from the admin panel.

### Item 9 — Apollo Client setup in admin Vue
Install `@apollo/client` + `graphql` in `admin/`.
Create `admin/src/plugins/apollo.js`:
```js
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client/core'
const link = createHttpLink({ uri: import.meta.env.VITE_ATTENDANCE_GRAPHQL_URL })
export const apolloClient = new ApolloClient({ link, cache: new InMemoryCache() })
```
Add `VITE_ATTENDANCE_GRAPHQL_URL=http://localhost:4001/graphql` to admin `.env`.

### Item 10 — Student Attendance Store
`admin/src/stores/admin/studentAttendance.js` (Pinia):
- `checkIn(passcode)` → mutation `checkInStudent`
- `checkOut(passcode)` → mutation `checkOutStudent`
- `fetchTodayStats(institutionId)` → query `todayStats`
- `fetchRecentScans(institutionId)` → query `recentScans`
- `fetchHistory(filters)` → query `attendanceHistory`

### Item 11 — StudentAttendanceStation.vue (check-in/out terminal)
`admin/src/pages/admin/StudentAttendanceStation.vue`:
- Input field for passcode or QR scan (use camera via `getUserMedia` or just text input)
- Toggle: Check In / Check Out mode
- After scan: show student photo + name for 3s with success animation
- Live "Today's Stats" sidebar: total, checked-in, checked-out, still inside
- Recent scans list (last 5)
- Colors: green for check-in, orange for check-out

### Item 12 — StudentAttendanceHistory.vue
`admin/src/pages/admin/StudentAttendanceHistory.vue`:
- Date picker filter + student name search
- Paginated table: Name | Grade | Class | Check In | Check Out | Status | Recorded By
- Status chips: present (green), absent (red), late (orange)
- Export button (CSV download from table data)

### Item 13 — Router + Appbar nav items
In `admin/src/router/index.js` add:
```js
{ path: 'students/attendance',  name: 'StudentAttendanceStation', component: () => import('@/pages/admin/StudentAttendanceStation.vue') },
{ path: 'students/history',     name: 'StudentAttendanceHistory',  component: () => import('@/pages/admin/StudentAttendanceHistory.vue') },
```
In `Appbar.vue` items array add (school type only):
```js
{ text: 'Attendance Station', icon: 'mdi-qrcode-scan',     to: '/app/students/attendance' },
{ text: 'Student History',    icon: 'mdi-book-clock-outline', to: '/app/students/history' },
```

### Item 14 — Scaffold `data-portal/`
```
data-portal/
├── src/
│   ├── config/
│   │   └── database.js
│   ├── graphql/
│   │   ├── schema.js
│   │   └── resolvers/
│   │       ├── download.js
│   │       └── deletion.js
│   ├── services/
│   │   ├── PDFService.js        ← puppeteer-based PDF generation
│   │   └── DeletionService.js   ← deletion queue management
│   ├── views/
│   │   ├── layouts/
│   │   │   └── main.ejs
│   │   ├── pages/
│   │   │   ├── staff/
│   │   │   │   ├── download.ejs
│   │   │   │   └── delete.ejs
│   │   │   ├── parent/
│   │   │   │   ├── download.ejs
│   │   │   │   └── delete.ejs
│   │   │   └── admin/
│   │   │       ├── download.ejs
│   │   │       └── delete.ejs
│   │   └── partials/
│   │       ├── header.ejs
│   │       └── success.ejs
│   ├── routes/
│   │   ├── index.js
│   │   └── graphql.js
│   ├── middleware/
│   │   └── auth.js              ← same JWT verify
│   └── app.js
├── scripts/
│   └── migrate.js               ← create data_deletion_requests table
│   └── cron.js                  ← 30-day hard delete cron
├── package.json
└── .env.example
```

**package.json deps:**
- `express`, `express-graphql`, `graphql`
- `pg`, `jsonwebtoken`, `dotenv`
- `puppeteer` or `puppeteer-core` (PDF generation)
- `ejs`, `express-session`
- `node-cron` (for the 30-day deletion cron)

### Item 15 — `data-portal` GraphQL Schema
```graphql
# Download
type DownloadRequest {
  url: String!          # signed URL or download path
  expiresAt: String!
  filename: String!
}

type Mutation {
  requestDataDownload(role: String!): DownloadRequest!
  requestAccountDeletion(reason: String): DeletionRequest!
  cancelDeletionRequest: Boolean!
}

type DeletionRequest {
  id: ID!
  scheduledDeleteAt: String!
  status: String!
  message: String!
}

type Query {
  myDeletionStatus: DeletionRequest
  myDataSummary: DataSummary!
}

type DataSummary {
  accountInfo: Boolean!
  attendanceRecords: Int
  pickupCodes: Int
  children: Int
  messages: Int
  supportTickets: Int
}
```

### Item 16 — `data-portal` DB migration + cron scaffold
`scripts/migrate.js` — creates `data_deletion_requests` table.
`scripts/cron.js` — node-cron job: every day at midnight, find records WHERE
`status = 'pending' AND scheduled_delete_at <= NOW()` and run hard delete.

---

## BATCH 3 — Data Download + Deletion (8 items)

### Item 17 — PDFService.js (puppeteer)
`data-portal/src/services/PDFService.js`:
Generates PDF by rendering an internal EJS HTML page and converting with puppeteer.
Methods:
- `generateStaffPDF(userId)` — fetches user profile + attendance history + shifts
- `generateParentPDF(userId)` — fetches parent profile + children + pickup codes
- `generateAdminPDF(userId)` — fetches admin profile + institution info + user list
Each method returns a `Buffer` (PDF bytes) for streaming as download.

### Item 18 — Staff Download: route + EJS page
`GET /staff/download` — EJS page showing:
- User info summary (name, email, role, joined date)
- "Download My Data (PDF)" button
- What's included: profile, attendance history, shift assignments, support tickets

`POST /staff/download/generate` — calls `PDFService.generateStaffPDF(userId)`, streams PDF.

**Staff PDF includes:**
- Full profile (name, email, phone, role, department, employee_id, status, joined)
- All attendance records (date, check-in, check-out, hours, status, shift)
- All shift assignments
- All support tickets + responses
- All direct messages (metadata only: partner name, count)

### Item 19 — Parent Download: route + EJS page
`GET /parent/download` — EJS page:
- User info summary
- "Download My Data (PDF)" button
- What's included: profile, children, pickup codes

`POST /parent/download/generate` — streams PDF.

**Parent PDF includes:**
- Full profile (name, email, phone, status, joined)
- All children (name, student_id, grade, class, photo, status)
- All pickup codes generated (code, child name, pickup person, status, created/expires)
- All support tickets + responses
- All direct messages (metadata)

### Item 20 — Admin Download: route + EJS page
`GET /admin/download` — EJS page:
- "Download My Data (PDF)" button
- What's included: profile, institution info, activity

`POST /admin/download/generate` — streams PDF.

**Admin PDF includes:**
- Full profile
- Institution details (name, type, subscription, settings)
- All staff in institution (name, email, role, status)
- All support tickets handled
- All direct messages (metadata)
- System activity log entries (last 100)

### Item 21 — DeletionService.js
`data-portal/src/services/DeletionService.js`:
```js
async requestDeletion(userId, role) {
  // 1. Create data_deletion_requests record (scheduled_delete_at = NOW() + 30d)
  // 2. Deactivate account: UPDATE users SET status = 'deactivating' WHERE id = $1
  // 3. Expire all sessions: DELETE FROM user_sessions WHERE user_id = $1
  // 4. Log activity
  return deletionRequest
}

async cancelDeletion(userId) {
  // 1. Find pending request for userId
  // 2. UPDATE status = 'cancelled', cancelled_at = NOW()
  // 3. Reactivate account: UPDATE users SET status = 'approved' WHERE id = $1
}

async executeDeletion(deletionRequestId) {
  // Hard delete — called by cron job
  // 1. DELETE FROM attendance WHERE user_id = $1
  // 2. DELETE FROM support_responses WHERE created_by = $1
  // 3. DELETE FROM direct_messages WHERE sender_id = $1 OR recipient_id = $1
  // 4. DELETE FROM announcements reads
  // 5. For parent: unassign children (DELETE FROM children parent_id)
  // 6. DELETE FROM users WHERE id = $1
  // 7. UPDATE data_deletion_requests SET status = 'completed', completed_at = NOW()
}
```

### Item 22 — Staff Deletion: route + EJS page
`GET /staff/delete` — EJS page:
- Warning: "This will permanently delete your account after 30 days"
- Lists what will be deleted
- Form with reason textarea + confirmation checkbox
- If already requested: shows countdown to deletion + "Cancel Request" button

`POST /staff/delete/request` — calls `DeletionService.requestDeletion(userId, 'staff')`
`POST /staff/delete/cancel` — calls `DeletionService.cancelDeletion(userId)`

**Immediate effects on request:**
1. Account status → `'deactivating'`
2. All active JWT sessions invalidated (delete from session store if any, or use a `token_invalidated_at` timestamp)
3. User logged out of all devices
4. Email confirmation sent

### Item 23 — Parent Deletion: route + EJS page
Same pattern as staff deletion.
`GET /parent/delete`, `POST /parent/delete/request`, `POST /parent/delete/cancel`

**Extra steps for parent hard delete:**
- Unassign children (set `parent_id = NULL` on `children` table)
- Cancel pending pickup codes

### Item 24 — Admin Deletion: route + EJS page
Same pattern.
`GET /admin/delete`, `POST /admin/delete/request`, `POST /admin/delete/cancel`

**Extra validation for admin:**
- Cannot delete if only admin in institution (warn: assign another admin first)
- Hard delete also needs to handle institution orphaning

---

## BATCH 4 — Cron + Legal (4 items)

### Item 25 — 30-day cron job (`scripts/cron.js`)
`node-cron` schedule: `'0 2 * * *'` (2am daily)
```js
cron.schedule('0 2 * * *', async () => {
  const dueRequests = await pool.query(
    `SELECT id, user_id, role FROM data_deletion_requests
     WHERE status = 'pending' AND scheduled_delete_at <= NOW()`
  )
  for (const req of dueRequests.rows) {
    await DeletionService.executeDeletion(req.id)
  }
})
```
Run this cron in `app.js` on startup.

### Item 26 — Email notifications
- On deletion request: send "Your account will be deleted on [date]. Click here to cancel."
- On deletion completion: send "Your account and data have been permanently deleted."
- Use `EmailService` from main server (or replicate SMTP config in data-portal)

### Item 27 — `legal/` static HTML project
```
legal/
├── index.html           ← redirects to privacy-policy.html
├── privacy-policy.html  ← full privacy policy
├── data-handling.html   ← how we handle, store, and delete user data
├── terms.html           ← terms of service
└── css/
    └── style.css
```

**Privacy Policy covers:**
- What data is collected (profile, attendance, location, messages, pickup codes)
- How it's used (attendance tracking, school management, communication)
- Data retention (active accounts: indefinitely; deleted accounts: 30-day queue → purged)
- Third parties (SMTP provider, Cloudinary for photos, hosting)
- User rights: request download, request deletion, correction
- Contact: privacy@clockee.app

**Data Handling page covers:**
- Data storage (PostgreSQL, encrypted at rest)
- JWT authentication (stateless, 15min access tokens)
- Photo storage (Cloudinary CDN)
- How to download your data (link to data-portal)
- How to delete your data (link to data-portal)

### Item 28 — Final integration + env docs
Create `.env.example` files for both microservices.
Update memory with new microservice ports and paths.

---

## Build Order

| Batch | Items | Focus |
|-------|-------|-------|
| 1 | 1–8   | Student Attendance Microservice (Node.js + GraphQL) |
| 2 | 9–16  | Admin Panel GraphQL integration + Data Portal scaffold |
| 3 | 17–24 | Download PDF generation + Deletion service + EJS pages |
| 4 | 25–28 | Cron job + Email + Legal HTML |

---

## Key Technical Decisions

1. **GraphQL server:** `apollo-server-express` (v3) — stable, widely used
2. **PDF generation:** `puppeteer` — renders HTML to PDF, full CSS support; fallback: `pdfkit` if puppeteer too heavy
3. **Auth:** Both microservices share `JWT_SECRET` with main server — no OAuth needed
4. **Data portal access:** Users access `data-portal` via token in query param OR by re-entering credentials (simple session-based login page per role)
5. **Passcode lookup:** Students have a `passcode` field on `children` table (already used in `StudentManagementController.inviteStudent`). Staff scan this to check in.
6. **No pub/sub:** Student attendance mutations broadcast via polling (recentScans query every 3s from admin panel). GraphQL subscriptions are out of scope for now.

---

## Summary Count

- **New microservices:** 2 (`students-attendance-service`, `data-portal`)
- **New static project:** 1 (`legal/`)
- **New DB tables:** 2 (`student_attendance`, `data_deletion_requests`)
- **New Vue pages:** 2 (`StudentAttendanceStation.vue`, `StudentAttendanceHistory.vue`)
- **New EJS pages:** 6 (staff/parent/admin × download/delete)
- **Total plan items:** 28, built in batches of 8
