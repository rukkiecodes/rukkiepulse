require('dotenv').config();
const express = require('express');
const cors = require('cors');

const connectRoute = require('./routes/connect');
const servicesRoute = require('./routes/services');

const app = express();

app.use(cors());
app.use(express.json());

// ── Routes ────────────────────────────────────────────────────────────────────

// Health check — publicly accessible
app.get('/health', (req, res) => {
  res.json({ ok: true, service: 'rukkiepulse-api', timestamp: new Date().toISOString() });
});

// Services POST here to register + heartbeat
app.use('/v1/connect', connectRoute);

// CLI GETs this to see all services and their status
app.use('/v1/services', servicesRoute);

// ── 404 ───────────────────────────────────────────────────────────────────────
app.use((req, res) => {
  res.status(404).json({ error: 'Not found' });
});

// ── Error handler ─────────────────────────────────────────────────────────────
app.use((err, req, res, next) => {
  console.error(err);
  res.status(500).json({ error: 'Internal server error' });
});

// ── Start (local dev only — Vercel uses the exported app) ─────────────────────
if (require.main === module) {
  const PORT = process.env.PORT || 4010;
  app.listen(PORT, () => {
    console.log(`\n🚀 RukkiePulse API`);
    console.log(`✅ Running on http://localhost:${PORT}`);
    console.log(`📡 POST /v1/connect   — service heartbeat + auto-registration`);
    console.log(`📋 GET  /v1/services  — CLI status query`);
    console.log(`💚 GET  /health       — health check\n`);
  });
}

module.exports = app;
