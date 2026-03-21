/**
 * GET /v1/services
 *
 * Returns all registered services with their connection status.
 * Called by the RukkiePulse CLI (rukkie scan / rukkie watch).
 *
 * Request headers:
 *   x-rukkie-cli: <CLI_SECRET>
 *
 * Response:
 *   { services: [{ id, name, language, description, activeKeys, lastUsedAt }] }
 */

const express = require('express');
const supabase = require('../db');

const router = express.Router();

const CLI_SECRET = process.env.RUKKIE_CLI_SECRET || 'rukkie-cli-v1-xqmjdjjwprnqogokoejz';

router.get('/', async (req, res) => {
  const cliHeader = req.headers['x-rukkie-cli'] ?? '';
  if (cliHeader !== CLI_SECRET) {
    return res.status(401).json({ error: 'Unauthorized' });
  }

  const { data: services, error } = await supabase
    .from('services')
    .select('id, name, language, description, created_at, api_keys(id, label, key_prefix, last_used_at, revoked_at)')
    .order('created_at', { ascending: true });

  if (error) {
    return res.status(500).json({ error: error.message });
  }

  const result = (services ?? []).map((svc) => {
    const keys = Array.isArray(svc.api_keys) ? svc.api_keys : [];
    const activeKeys = keys.filter((k) => !k.revoked_at);

    let lastUsedAt = null;
    for (const k of activeKeys) {
      if (!lastUsedAt || (k.last_used_at && k.last_used_at > lastUsedAt)) {
        lastUsedAt = k.last_used_at;
      }
    }

    return {
      id: svc.id,
      name: svc.name,
      language: svc.language ?? 'other',
      description: svc.description ?? '',
      activeKeys: activeKeys.length,
      lastUsedAt,
    };
  });

  return res.json({ services: result });
});

module.exports = router;
