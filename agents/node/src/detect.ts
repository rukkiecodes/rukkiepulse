export type Framework = 'express' | 'fastify' | 'unknown'

export function detectFramework(app: unknown): Framework {
  if (!app || typeof app !== 'object') return 'unknown'

  const a = app as Record<string, unknown>

  // Express: has _router or use() + get() + listen()
  if (typeof a['use'] === 'function' && typeof a['get'] === 'function' && '_router' in a) {
    return 'express'
  }

  // Fastify: has route() and ready()
  if (typeof a['route'] === 'function' && typeof a['ready'] === 'function') {
    return 'fastify'
  }

  // Fallback Express detection (app.use + app.listen without _router yet)
  if (typeof a['use'] === 'function' && typeof a['listen'] === 'function') {
    return 'express'
  }

  return 'unknown'
}
