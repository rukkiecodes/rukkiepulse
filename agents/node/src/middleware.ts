import { context, trace, SpanStatusCode } from '@opentelemetry/api'
import type { Framework } from './detect'
import type { RukkieConfig } from './setup'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type AnyApp = any

export function injectMiddleware(app: AnyApp, framework: Framework, config: RukkieConfig): void {
  if (framework === 'express') {
    app.use(expressMiddleware(config.serviceName))
  } else if (framework === 'fastify') {
    app.addHook('onRequest', fastifyOnRequest(config.serviceName))
    app.addHook('onResponse', fastifyOnResponse())
    app.addHook('onError', fastifyOnError())
  }
  // 'unknown': OTel HTTP auto-instrumentation handles it
}

function expressMiddleware(serviceName: string) {
  return function rukkieMiddleware(
    req: AnyApp,
    res: AnyApp,
    next: AnyApp
  ): void {
    const tracer = trace.getTracer(serviceName)
    const span = tracer.startSpan(`${req.method} ${req.path ?? req.url}`)
    const ctx = trace.setSpan(context.active(), span)

    const start = Date.now()

    context.with(ctx, () => {
      res.on('finish', () => {
        const duration = Date.now() - start
        span.setAttribute('http.method', req.method)
        span.setAttribute('http.route', req.path ?? req.url)
        span.setAttribute('http.status_code', res.statusCode)
        span.setAttribute('http.duration_ms', duration)

        if (res.statusCode >= 500) {
          span.setStatus({ code: SpanStatusCode.ERROR, message: `HTTP ${res.statusCode}` })
        }
        span.end()
      })

      next()
    })
  }
}

function fastifyOnRequest(serviceName: string) {
  return async (request: AnyApp, _reply: AnyApp): Promise<void> => {
    const tracer = trace.getTracer(serviceName)
    const span = tracer.startSpan(`${request.method} ${request.url}`)
    request._rukkieSpan = span
    request._rukkieStart = Date.now()
  }
}

function fastifyOnResponse() {
  return async (request: AnyApp, reply: AnyApp): Promise<void> => {
    const span = request._rukkieSpan
    if (!span) return
    const duration = Date.now() - (request._rukkieStart ?? 0)
    span.setAttribute('http.status_code', reply.statusCode)
    span.setAttribute('http.duration_ms', duration)
    if (reply.statusCode >= 500) {
      span.setStatus({ code: SpanStatusCode.ERROR, message: `HTTP ${reply.statusCode}` })
    }
    span.end()
  }
}

function fastifyOnError() {
  return async (error: AnyApp, request: AnyApp, _reply: AnyApp): Promise<void> => {
    const span = request._rukkieSpan
    if (!span) return
    span.recordException(error)
    span.setStatus({ code: SpanStatusCode.ERROR, message: error.message })
  }
}
