import type { Framework } from './detect'
import type { RukkieConfig } from './setup'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyApp = any

export function registerHealthRoute(app: AnyApp, framework: Framework, config: RukkieConfig): void {
  if (framework === 'express') {
    app.get('/__rukkie/health', async (_req: AnyApp, res: AnyApp) => {
      const body = await buildHealthResponse(config)
      res.status(body.status === 'ok' ? 200 : 207).json(body)
    })
  } else if (framework === 'fastify') {
    app.get('/__rukkie/health', async (_request: AnyApp, reply: AnyApp) => {
      const body = await buildHealthResponse(config)
      reply.status(body.status === 'ok' ? 200 : 207).send(body)
    })
  }
}

async function buildHealthResponse(config: RukkieConfig): Promise<HealthResponse> {
  const deps: Record<string, string> = {}

  if (config.dependencies) {
    await Promise.all(
      Object.entries(config.dependencies).map(async ([name, check]) => {
        try {
          const ok = await check()
          deps[name] = ok ? 'connected' : 'error'
        } catch {
          deps[name] = 'error'
        }
      })
    )
  }

  const allOk = Object.values(deps).every((v) => v === 'connected')
  const status = Object.keys(deps).length === 0 ? 'ok' : allOk ? 'ok' : 'degraded'

  return {
    status,
    service: config.serviceName,
    dependencies: deps,
  }
}

interface HealthResponse {
  status: 'ok' | 'degraded'
  service: string
  dependencies: Record<string, string>
}
