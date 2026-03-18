from __future__ import annotations

from opentelemetry import trace
from opentelemetry.sdk.resources import Resource, SERVICE_NAME
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter

_initialized = False


def setup_otel(config: "RukkieConfig") -> None:  # type: ignore[name-defined]
    global _initialized
    if _initialized:
        return

    resource = Resource(attributes={SERVICE_NAME: config.service_name})
    provider = TracerProvider(resource=resource)

    exporter = OTLPSpanExporter(endpoint=config.collector_url, insecure=True)
    provider.add_span_processor(BatchSpanProcessor(exporter))

    trace.set_tracer_provider(provider)

    _instrument_libraries()
    _initialized = True


def _instrument_libraries() -> None:
    """Auto-instrument common HTTP libraries if available."""
    try:
        from opentelemetry.instrumentation.requests import RequestsInstrumentor
        RequestsInstrumentor().instrument()
    except ImportError:
        pass

    try:
        from opentelemetry.instrumentation.httpx import HTTPXClientInstrumentor
        HTTPXClientInstrumentor().instrument()
    except ImportError:
        pass
