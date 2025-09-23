# mirador-rca

> Work-in-progress Root Cause Analysis service for the Mirador stack.

## Prerequisites
- Go 1.23+
- `protoc` with Go & gRPC plugins (`protoc-gen-go`, `protoc-gen-go-grpc`).
- External Weaviate cluster reachable from the service.
- mirador-core API access for metrics/logs/traces aggregation.
- **Mandatory:** Deploy the OpenTelemetry Collector [servicegraphconnector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/servicegraphconnector) and ensure its emitted service graph metrics are available. mirador-rca relies on this topology data to correlate anomalies across services; if the endpoint is missing or empty, investigations fail.
- Configure mirador-core to expose a service-graph endpoint (default `/api/v1/rca/service-graph`) that proxies the connector metrics so mirador-rca can fetch the dependency topology prior to each investigation.
- mirador-rca performs no synthetic fallbacks—metrics, logs, traces, and service graph data **must** be returned by mirador-core for investigations to succeed.
- See `docs/phase3-eval.md` for precision@1 evaluation details and `docs/indicent-analysis.md` for a deep dive of the pipeline.

## Quickstart
```
make phase2-verify   # run unit suites
make smoke-weaviate  # verify external Weaviate readiness
```

Run the service:
```
go run ./cmd/rca-engine --config configs/config.yaml
```

Configuration fields are documented in `configs/config.example.yaml`.
