/**
 * POST /v1/connect
 *
 * Called by any backend service on startup (and periodically) to register
 * itself and update its heartbeat. Auto-creates the service + API key record
 * on first call — no dashboard setup required.
 *
 * Request headers:
 *   Authorization: Bearer <api_key>
 *
 * Request body (JSON):
 *   { serviceName: string, language?: string }
 *
 * Response:
 *   { ok: true, service: string, registered?: true }
 */

const express = require('express');
const crypto = require('crypto');
const supabase = require('../db');

const router = express.Router();

router.post('/', async (req, res) => {
  try {
    // Extract API key from Authorization header
    const authHeader = req.headers['authorization'] ?? '';
    const apiKey = authHeader.replace(/^Bearer\s+/i, '').trim();

    if (!apiKey) {
      return res.status(401).json({ error: 'Missing API key — send Authorization: Bearer <key>' });
    }

    const { serviceName, language = 'other' } = req.body ?? {};

    if (!serviceName) {
      return res.status(400).json({ error: 'Missing serviceName in request body' });
    }

    // Hash the API key with SHA-256 (same algorithm used by the dashboard)
    const keyHash = crypto.createHash('sha256').update(apiKey).digest('hex');
    const keyPrefix = apiKey.slice(0, 14); // rk_live_xxxxxx

    // Check if the key already exists
    const { data: existingKey } = await supabase
      .from('api_keys')
      .select('id, service_id, revoked_at, services(name, language)')
      .eq('key_hash', keyHash)
      .single();

    if (existingKey) {
      if (existingKey.revoked_at) {
        return res.status(403).json({ error: 'API key has been revoked' });
      }

      // Update heartbeat timestamp
      await supabase
        .from('api_keys')
        .update({ last_used_at: new Date().toISOString() })
        .eq('id', existingKey.id);

      const svc = Array.isArray(existingKey.services)
        ? existingKey.services[0]
        : existingKey.services;

      return res.json({ ok: true, service: svc?.name ?? serviceName });
    }

    // ── Auto-register: first time this key is seen ────────────────────────────

    // Find or create a system owner to attach the service to
    const ownerId = await resolveOwnerId();

    // Create the service record
    const { data: newService, error: svcErr } = await supabase
      .from('services')
      .insert({
        name: serviceName,
        language,
        owner_id: ownerId,
        description: `Auto-registered via RukkiePulse Connect API`,
      })
      .select('id, name')
      .single();

    if (svcErr) {
      console.error('Failed to create service:', svcErr);
      return res.status(500).json({ error: 'Failed to register service' });
    }

    // Store the hashed API key
    const { error: keyErr } = await supabase
      .from('api_keys')
      .insert({
        service_id: newService.id,
        label: 'auto-registered',
        key_prefix: keyPrefix,
        key_hash: keyHash,
        last_used_at: new Date().toISOString(),
      });

    if (keyErr) {
      console.error('Failed to store API key:', keyErr);
      return res.status(500).json({ error: 'Failed to store API key' });
    }

    return res.status(201).json({ ok: true, service: serviceName, registered: true });
  } catch (err) {
    console.error('Connect error:', err);
    return res.status(500).json({ error: 'Internal server error' });
  }
});

// ── Helpers ──────────────────────────────────────────────────────────────────

let _cachedOwnerId = null;

async function resolveOwnerId() {
  if (_cachedOwnerId) return _cachedOwnerId;

  // Use the configured owner ID from env if provided
  if (process.env.RUKKIE_OWNER_ID) {
    _cachedOwnerId = process.env.RUKKIE_OWNER_ID;
    return _cachedOwnerId;
  }

  // Fall back: look up the first admin user by email
  const { data } = await supabase.auth.admin.listUsers();
  const owner = data?.users?.find(u => u.email === process.env.RUKKIE_OWNER_EMAIL)
    ?? data?.users?.[0];

  if (!owner) throw new Error('No owner user found — set RUKKIE_OWNER_ID env var');
  _cachedOwnerId = owner.id;
  return _cachedOwnerId;
}

module.exports = router;
