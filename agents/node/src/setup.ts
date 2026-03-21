import { setupOtel } from './otel'
import { injectMiddleware } from './middleware'
import { registerHealthRoute } from './health'
import { detectFramework } from './detect'

export type DependencyCheck = () => Promise<boolean>

export interface RukkieConfig {
  serviceName: string
  apiKey: string
  collectorUrl?: string  // default: http://localhost:4317
  connectUrl?: string    // default: https://rukkieapi.vercel.app/v1/connect
  dependencies?: Record<string, DependencyCheck>
}

const DEFAULT_CONNECT_URL = 'https://rukkieapi.vercel.app/v1/connect'

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
function pingConnect(config: RukkieConfig): void {
  const url = config.connectUrl ?? DEFAULT_CONNECT_URL
  // Fire-and-forget — never block startup
  fetch(url, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${config.apiKey}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ serviceName: config.serviceName, language: 'node' }),
  }).catch(() => {/* silent — observability must not crash the service */})
}

export function initRukkie(config: RukkieConfig, app?: unknown): void {
  // 1. Boot OTel first — must happen before any HTTP handlers are registered
  setupOtel(config)

  // 2. Ping RukkiePulse Connect API — auto-registers + updates heartbeat
  pingConnect(config)

  if (!app) return

  const framework = detectFramework(app)

  // 3. Inject request tracing middleware
  injectMiddleware(app, framework, config)

  // 4. Expose /__rukkie/health
  registerHealthRoute(app, framework, config)
}
