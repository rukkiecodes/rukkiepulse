# Phase 4 — Support System

## Goal
Build the real-time support ticket system: GraphQL schema + resolvers, Socket.io server, EJS ticket list and thread pages.

**Prerequisite:** Phase 3 complete (EJS pages working, layout in place, routes structured).

---

## Scope

### Files to Create

```
super-admin/src/graphql/schema/
└── support.graphql

super-admin/src/graphql/resolvers/
└── supportResolvers.js

super-admin/src/services/
└── SupportService.js

super-admin/src/views/pages/support/
├── tickets.ejs
└── thread.ejs

super-admin/src/public/js/
└── support-socket.js
```

### Files to Modify

```
super-admin/src/graphql/schema/index.js   ← include support.graphql
super-admin/src/graphql/resolvers/index.js ← include supportResolvers
super-admin/src/routes/index.js           ← add /support and /support/:id GET routes
super-admin/server.js                     ← upgrade HTTP server to Socket.io
super-admin/app.js                        ← export app before Socket.io wraps it
```

---

## Implementation Steps

### Step 1 — support.graphql

```graphql
type SupportTicket {
  id: ID!
  institutionId: Int!
  institutionName: String
  openedBy: String
  subject: String!
  status: String!
  priority: String!
  createdAt: String!
  updatedAt: String!
  resolvedAt: String
  messageCount: Int
}

type SupportMessage {
  id: ID!
  ticketId: Int!
  senderId: Int
  senderRole: String!
  senderName: String
  message: String!
  attachments: String
  readAt: String
  createdAt: String!
}

type TicketPage {
  items: [SupportTicket!]!
  total: Int!
  page: Int!
  pages: Int!
}

input CreateTicketInput {
  institutionId: Int!
  subject: String!
  priority: String
  message: String!
}

extend type Query {
  supportTickets(status: String, institutionId: ID, page: Int, limit: Int): TicketPage
  supportTicket(id: ID!): SupportTicket
  supportMessages(ticketId: ID!): [SupportMessage]
}

extend type Mutation {
  createSupportTicket(data: CreateTicketInput!): SupportTicket
  sendSupportMessage(ticketId: ID!, message: String!): SupportMessage
  updateTicketStatus(ticketId: ID!, status: String!): SupportTicket
}

type Subscription {
  supportMessageReceived(ticketId: ID!): SupportMessage
  newSupportTicket: SupportTicket
}
```

### Step 2 — SupportService.js

```js
const pool = require('../config/database');

class SupportService {
  static async getTicketsPage({ status, institutionId, page = 1, limit = 20 }) {
    const conditions = [];
    const params = [];

    if (status) { params.push(status); conditions.push(`st.status = $${params.length}`); }
    if (institutionId) { params.push(institutionId); conditions.push(`st.institution_id = $${params.length}`); }

    const where = conditions.length ? `WHERE ${conditions.join(' AND ')}` : '';

    const countRes = await pool.query(`SELECT COUNT(*) FROM support_tickets st ${where}`, params);
    const total = parseInt(countRes.rows[0].count, 10);

    const offset = (page - 1) * limit;
    params.push(limit, offset);

    const rows = await pool.query(`
      SELECT st.*, i.name AS institution_name,
        u.name AS opened_by_name,
        (SELECT COUNT(*) FROM support_messages sm WHERE sm.ticket_id = st.id) AS message_count
      FROM support_tickets st
      LEFT JOIN institutions i ON i.id = st.institution_id
      LEFT JOIN users u ON u.id = st.opened_by
      ${where}
      ORDER BY st.updated_at DESC
      LIMIT $${params.length - 1} OFFSET $${params.length}
    `, params);

    return { items: rows.rows, total, page, pages: Math.ceil(total / limit) };
  }

  static async getTicketById(id) {
    const res = await pool.query(`
      SELECT st.*, i.name AS institution_name, u.name AS opened_by_name
      FROM support_tickets st
      LEFT JOIN institutions i ON i.id = st.institution_id
      LEFT JOIN users u ON u.id = st.opened_by
      WHERE st.id = $1
    `, [id]);
    return res.rows[0] || null;
  }

  static async getMessages(ticketId) {
    const res = await pool.query(`
      SELECT sm.*, u.name AS sender_name
      FROM support_messages sm
      LEFT JOIN users u ON u.id = sm.sender_id
      WHERE sm.ticket_id = $1
      ORDER BY sm.created_at ASC
    `, [ticketId]);
    return res.rows;
  }

  static async createTicket({ institutionId, openedBy, subject, priority = 'normal', message }) {
    const client = await pool.connect();
    try {
      await client.query('BEGIN');
      const ticketRes = await client.query(
        `INSERT INTO support_tickets (institution_id, opened_by, subject, priority)
         VALUES ($1, $2, $3, $4) RETURNING *`,
        [institutionId, openedBy, subject, priority]
      );
      const ticket = ticketRes.rows[0];
      await client.query(
        `INSERT INTO support_messages (ticket_id, sender_id, sender_role, message)
         VALUES ($1, $2, 'super_admin', $3)`,
        [ticket.id, openedBy, message]
      );
      await client.query('COMMIT');
      return ticket;
    } catch (err) {
      await client.query('ROLLBACK');
      throw err;
    } finally {
      client.release();
    }
  }

  static async addMessage({ ticketId, senderId, senderRole, message }) {
    const res = await pool.query(
      `INSERT INTO support_messages (ticket_id, sender_id, sender_role, message)
       VALUES ($1, $2, $3, $4) RETURNING *`,
      [ticketId, senderId, senderRole, message]
    );
    // Update ticket's updated_at
    await pool.query(
      `UPDATE support_tickets SET updated_at = NOW() WHERE id = $1`,
      [ticketId]
    );
    return res.rows[0];
  }

  static async updateStatus(ticketId, status) {
    const resolved = status === 'resolved' ? 'NOW()' : 'NULL';
    const res = await pool.query(
      `UPDATE support_tickets
       SET status = $1, updated_at = NOW(), resolved_at = ${resolved}
       WHERE id = $2 RETURNING *`,
      [status, ticketId]
    );
    return res.rows[0];
  }

  static async getByInstitution(institutionId) {
    const res = await pool.query(
      `SELECT * FROM support_tickets WHERE institution_id = $1 ORDER BY updated_at DESC`,
      [institutionId]
    );
    return res.rows;
  }
}

module.exports = SupportService;
```

### Step 3 — supportResolvers.js

```js
const SupportService = require('../../services/SupportService');

module.exports = {
  Query: {
    supportTickets: async (_, { status, institutionId, page, limit }, { assertSuperAdmin }) => {
      assertSuperAdmin();
      return SupportService.getTicketsPage({ status, institutionId, page, limit });
    },
    supportTicket: async (_, { id }, { assertSuperAdmin }) => {
      assertSuperAdmin();
      return SupportService.getTicketById(id);
    },
    supportMessages: async (_, { ticketId }, { assertSuperAdmin }) => {
      assertSuperAdmin();
      return SupportService.getMessages(ticketId);
    },
  },
  Mutation: {
    createSupportTicket: async (_, { data }, { user, assertSuperAdmin }) => {
      assertSuperAdmin();
      return SupportService.createTicket({ ...data, openedBy: user.id });
    },
    sendSupportMessage: async (_, { ticketId, message }, { user, assertSuperAdmin, io }) => {
      assertSuperAdmin();
      const msg = await SupportService.addMessage({ ticketId, senderId: user.id, senderRole: 'super_admin', message });
      // Emit via Socket.io
      if (io) io.to(`ticket:${ticketId}`).emit('new_message', msg);
      return msg;
    },
    updateTicketStatus: async (_, { ticketId, status }, { assertSuperAdmin }) => {
      assertSuperAdmin();
      return SupportService.updateStatus(ticketId, status);
    },
  },
};
```

### Step 4 — Upgrade server.js for Socket.io

```js
require('dotenv').config();
const http = require('http');
const { Server: SocketServer } = require('socket.io');
const app = require('./app');

const PORT = process.env.PORT || 4000;
const httpServer = http.createServer(app);

const io = new SocketServer(httpServer, {
  cors: { origin: '*' }, // restrict in production
});

// Attach io to app so routes/context can access it
app.set('io', io);

io.on('connection', (socket) => {
  console.log('[socket] connected', socket.id);

  socket.on('join_ticket', (ticketId) => {
    socket.join(`ticket:${ticketId}`);
    console.log(`[socket] ${socket.id} joined ticket:${ticketId}`);
  });

  socket.on('send_message', async (data) => {
    // data: { ticketId, message, senderId, senderRole }
    const { SupportService } = require('./src/services/SupportService');
    const msg = await SupportService.addMessage(data);
    io.to(`ticket:${data.ticketId}`).emit('new_message', msg);
  });

  socket.on('disconnect', () => {
    console.log('[socket] disconnected', socket.id);
  });
});

// Pass io into GraphQL context via app
httpServer.listen(PORT, () => {
  console.log(`[super-admin] Running on port ${PORT}`);
});
```

Update `src/graphql/context.js` to pull `io` from `req.app.get('io')`:
```js
module.exports = async ({ req }) => ({
  user: req.session?.user || null,
  req,
  io: req.app.get('io'),
  assertSuperAdmin() {
    if (!this.user || this.user.role !== 'super_admin') throw new Error('UNAUTHORIZED');
  },
});
```

### Step 5 — EJS: /support (tickets list)

Route:
```js
router.get('/support', requireSuperAdmin, async (req, res) => {
  const { status, page = 1 } = req.query;
  const result = await SupportService.getTicketsPage({ status, page: parseInt(page), limit: 25 });
  res.render('pages/support/tickets', {
    user: req.session.user, tickets: result.items,
    total: result.total, pages: result.pages, currentPage: result.page,
    filters: { status }, currentPath: '/support',
  });
});
```

Layout:
- Filter tabs: All | Open | In Progress | Resolved | Closed (as colored nav tabs)
- Table: Ticket # | Institution | Subject | Priority | Status | Opened | Messages | Actions
- Status badges: color-coded
- Priority badges: `urgent`=red, `high`=orange, `normal`=blue, `low`=gray
- "View Thread" button per row

### Step 6 — EJS: /support/:id (thread)

Route:
```js
router.get('/support/:id', requireSuperAdmin, async (req, res) => {
  const [ticket, messages] = await Promise.all([
    SupportService.getTicketById(req.params.id),
    SupportService.getMessages(req.params.id),
  ]);
  if (!ticket) return res.status(404).render('pages/404');
  res.render('pages/support/thread', {
    user: req.session.user, ticket, messages, currentPath: '/support',
  });
});
```

Layout:
```
┌──────────────────────────────────────────────────────┐
│ Ticket header: Subject | Institution | Status control │
├──────────────────────────────────────────────────────┤
│ Message bubbles (scrollable):                         │
│   Admin messages: left-aligned (gray bg)              │
│   Super admin messages: right-aligned (blue bg)       │
├──────────────────────────────────────────────────────┤
│ Message input + Send button                           │
└──────────────────────────────────────────────────────┘
```

Status control: dropdown select + "Update Status" button (POST form).

### Step 7 — src/public/js/support-socket.js

```js
// Loaded only on /support/:id thread page
const socket = io();

// Get ticketId from data attribute on body
const ticketId = document.body.dataset.ticketId;
socket.emit('join_ticket', ticketId);

socket.on('new_message', (msg) => {
  appendMessage(msg);
  scrollToBottom();
});

function appendMessage(msg) {
  const container = document.getElementById('messages');
  const isMe = msg.sender_role === 'super_admin';
  const div = document.createElement('div');
  div.className = `flex ${isMe ? 'justify-end' : 'justify-start'} mb-3`;
  div.innerHTML = `
    <div class="max-w-xs lg:max-w-md px-4 py-2 rounded-lg ${isMe ? 'bg-blue-600 text-white' : 'bg-gray-100 text-gray-900'}">
      <p class="text-sm">${escapeHtml(msg.message)}</p>
      <p class="text-xs mt-1 opacity-70">${new Date(msg.created_at).toLocaleTimeString()}</p>
    </div>
  `;
  container.appendChild(div);
}

// Send message via Socket.io (real-time) instead of form POST
document.getElementById('sendForm').addEventListener('submit', (e) => {
  e.preventDefault();
  const input = document.getElementById('messageInput');
  const message = input.value.trim();
  if (!message) return;
  socket.emit('send_message', { ticketId: parseInt(ticketId), message, senderRole: 'super_admin' });
  input.value = '';
});

function escapeHtml(text) {
  return text.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
}

function scrollToBottom() {
  const container = document.getElementById('messages');
  container.scrollTop = container.scrollHeight;
}

scrollToBottom();
```

---

## Done Criteria

- [ ] `/support` shows paginated ticket list with status filters
- [ ] `/support/:id` shows message thread with correct bubble alignment
- [ ] Super admin can send a message via the input box — it appears immediately without page reload
- [ ] If two browser tabs have the same ticket open, a message sent in one appears in the other in real time
- [ ] Status can be updated via the dropdown on the thread page
- [ ] New ticket can be created from institution detail page → appears in `/support` list
- [ ] `socketio` `join_ticket` event correctly joins the socket room
- [ ] Message count in ticket list updates after new messages are sent
