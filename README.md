# Managed LLM service scaffolded from Backstage (MLOps Golden Path).

## Architecture

- **App**: Go microservice with an embedded SQLite store.
- **Routing**: delegates to the FinOps arbitrage router. It sends
  `X-Router-Priority` (from the SLA) so the router decides: local Ollama CPU
  (cost-optimized) or managed fallback (max-performance). The service never
  reaches the model directly.
- **Network**: zero-trust egress — DNS + Kong only.
- **Delivery**: GitOps via Argo CD; this repo was published by Backstage, the
  control plane registers it through a PR and Argo syncs it. No kubectl needed.

## Local development

```bash
go run . --config config.yaml
curl -X POST localhost:8080/chat -d '{"prompt":"hola"}'
```

## Deploy

A PR to `control-plane` (workflow `workflow-alta-pr.yml`) registers this
workload; Argo CD applies it. Endpoint: `https://llm-microservice.local`.
