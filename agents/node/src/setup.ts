import { setupOtel } from './otel'
import { injectMiddleware } from './middleware'
import { registerHealthRoute } from './health'
import { detectFramework } from './detect'

export type DependencyCheck = () => Promise<boolean>

export interface RukkieConfig {
  serviceName: string
  apiKey: string
  collectorUrl?: string  // default: http://localhost:4317
  dependencies?: Record<string, DependencyCheck>
}

/**
 * Initialize the RukkiePulse agent.
 *
 * Call this BEFORE any other app setup so OTel is registered first.
 *
 * @example
 * // Express
 * import express from 'express'
 * import { initRukkie } from 'rukkie-agent'
 *
 * const app = express()
 * initRukkie({ serviceName: 'auth-service', apiKey: 'rk_live_xxx' }, app)
 */
const HEARTBEAT_URL =
  'https://xqmjdjjwprnqogokoejz.supabase.co/functions/v1/heartbeat'

function pingHeartbeat(apiKey: string): void {
  // Fire-and-forget — never block startup
  fetch(HEARTBEAT_URL, {
    method: 'POST',
    headers: { Authorization: `Bearer ${apiKey}` },
  }).catch(() => {/* silent — observability must not crash the service */})
}

export function initRukkie(config: RukkieConfig, app?: unknown): void {
  // 1. Boot OTel first — must happen before any HTTP handlers are registered
  setupOtel(config)

  // 2. Ping RukkiePulse dashboard so the service shows as connected
  pingHeartbeat(config.apiKey)

  if (!app) return

  const framework = detectFramework(app)

  // 3. Inject request tracing middleware
  injectMiddleware(app, framework, config)

  // 4. Expose /__rukkie/health
  registerHealthRoute(app, framework, config)
}
