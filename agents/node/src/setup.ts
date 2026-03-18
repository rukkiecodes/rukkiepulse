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
export function initRukkie(config: RukkieConfig, app?: unknown): void {
  // 1. Boot OTel first — must happen before any HTTP handlers are registered
  setupOtel(config)

  if (!app) return

  const framework = detectFramework(app)

  // 2. Inject request tracing middleware
  injectMiddleware(app, framework, config)

  // 3. Expose /__rukkie/health
  registerHealthRoute(app, framework, config)
}
