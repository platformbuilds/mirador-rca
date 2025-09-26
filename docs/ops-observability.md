# Operations Observability Handbook

This guide captures the SLOs, alert policies, and dashboards that ship with mirador-rca. Pair it with the Helm chart (`charts/mirador-rca`) when rolling out to new clusters.

## 1. Service-Level Objectives

| SLO | Target | Measurement | Notes |
| --- | ------ | ----------- | ----- |
| Investigation p95 latency | ≤ 4 s over 15 min windows | `mirador_rca_investigation_seconds` histogram | Matches action-plan exit criteria; investigate extractor thresholds if breached |
| Investigation success rate | ≥ 99.5% over 1 h windows | `mirador_rca_investigations_total{outcome}` | Treat `outcome="error"` spikes as availability incidents |
| gRPC availability | ≥ 99.5% | `grpc_server_handled_total` + `grpc_server_handled_total{grpc_code!="OK"}` | Delivered by `go-grpc-prometheus` interceptors |

## 2. Metrics Surface

mirador-rca starts an HTTP metrics listener on `server.metricsAddress` (default `:2112`). Scrape `/metrics` via Prometheus or an OpenTelemetry collector using the `prometheusreceiver`.

Key series:

- `mirador_rca_investigations_total{outcome}` – counter partitioned by `success` and `error` outcomes.
- `mirador_rca_investigation_seconds` – histogram backing the p95 latency SLO.
- `grpc_server_handled_total` / `grpc_server_handled_seconds_bucket` – emitted by `go-grpc-prometheus` for gRPC level telemetry.
- `process_*` and Go runtime stats – provided by the Prometheus client for capacity trending.

Set `.Values.metrics.enabled=false` (or blank `server.metricsAddress`) to disable the listener when an internal service mesh handles scraping. Update `.Values.metrics.annotations`/`.labels` to add `prometheus.io/*` hints or ServiceMonitor selectors.

## 3. Alert Catalogue

The Helm chart renders a `PrometheusRule` when `alerts.enabled=true`:

1. **MiradorRCAHighLatency** – triggers if the p95 latency exceeds 4 s for 5 minutes.
2. **MiradorRCANoTraffic** – warns when no investigations are handled for 10 minutes (helps catch stuck queues or routing issues).

Tune thresholds and severities via `values.yaml` (`alerts.rules`). For multi-tenant clusters, use namespace selectors in your Prometheus configuration or set custom labels through `alerts.labels`.

## 4. Dashboards

`dashboards.enabled=true` installs a Grafana dashboard ConfigMap. Panels included by default:

- p95 investigation latency time-series with threshold colouring.
- Investigations-per-minute stat sourced from `grpc_server_handled_total`.

Add additional panels (cache hit rate, Weaviate latency) by editing `values.yaml` (`dashboards.content`).

## 5. Runbook Snippets

- **High latency alert**: check recent Weaviate response times, mirador-core API latency, and Valkey availability. Enable debug logging temporarily (`logging.level=debug`) to dump extraction timings.
- **No traffic alert**: confirm mirador-core /rca gateway can reach the gRPC endpoint, inspect ingress/Service endpoints, and look for TLS handshake errors in logs.
- **Error-rate spike**: pivot on `mirador_rca_investigations_total{outcome="error"}`. Failures usually correspond to upstream mirador-core 5xx responses (metric/log/traces gaps) or Weaviate outages.

## 6. Configuration Reference

| Setting | Location | Purpose |
| ------- | -------- | ------- |
| `server.metricsAddress` | `configs/config.example.yaml` / `.Values.config.server.metricsAddress` | Bind address for the `/metrics` HTTP listener. Set to `""` to disable. |
| `.Values.metrics.*` | `charts/mirador-rca/values.yaml` | Controls port exposure, annotations, and labels for the metrics Service port. |
| `.Values.alerts.*` | Helm values | Enables/overrides Prometheus alerts. |
| `.Values.dashboards.*` | Helm values | Configures packaged Grafana dashboard. |

Keep this document updated as new metrics, alerts, or dashboards are introduced.
