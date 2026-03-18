import { NodeSDK } from '@opentelemetry/sdk-node'
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-grpc'
import { HttpInstrumentation } from '@opentelemetry/instrumentation-http'
import { ExpressInstrumentation } from '@opentelemetry/instrumentation-express'
import { FastifyInstrumentation } from '@opentelemetry/instrumentation-fastify'
import { Resource } from '@opentelemetry/resources'
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions'
import type { RukkieConfig } from './setup'

let sdk: NodeSDK | null = null

export function setupOtel(config: RukkieConfig): void {
  if (sdk) return // already initialized

  const collectorUrl = config.collectorUrl ?? 'http://localhost:4317'

  sdk = new NodeSDK({
    resource: new Resource({
      [ATTR_SERVICE_NAME]: config.serviceName,
    }),
    traceExporter: new OTLPTraceExporter({ url: collectorUrl }),
    instrumentations: [
      new HttpInstrumentation(),
      new ExpressInstrumentation(),
      new FastifyInstrumentation(),
    ],
  })

  sdk.start()

  process.on('SIGTERM', () => sdk?.shutdown())
  process.on('SIGINT', () => sdk?.shutdown())
}
