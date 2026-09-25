# Managed LLM service scaffolded from Backstage (MLOps Golden Path).

## Architecture

- **App**: Go HTTP service built into `docker.io/guslopezc/llm-microservice`.
- **Routing**: forwards OpenAI-compatible requests to the FinOps arbitrage
  router with `X-Router-Priority`, so the router selects local Ollama or its
  managed fallback. The service never reaches a model directly and never picks
  the model: the router owns both the backend and the model id.
- **Configuration**: the Deployment environment carries the router endpoint and
  priority; `config.yaml` records the same SLA plus the local model name.
- **Network**: zero-trust egress to DNS and the router service.
- **Delivery**: the repository workflow builds and pushes an immutable image
  tagged with the commit SHA, then updates the Deployment manifest. Argo CD
  syncs that commit.

## Local development

```bash
go test ./...
ROUTER_ENDPOINT=http://finops-arbitrage-router.finops-arbitrage-router.svc.cluster.local:8080 go run .
curl http://localhost:8080/healthz
curl -X POST http://localhost:8080/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -H 'X-Router-Priority: cost-optimized' \
  -d '{"messages":[{"role":"user","content":"hola"}]}'
```

## Image publication

The repository expects `DOCKER_USERNAME` and `DOCKER_PASSWORD` GitHub Actions
secrets. Each push to `main` publishes
`docker.io/guslopezc/llm-microservice:<commit-sha>` plus the rolling `:ci` tag,
then updates `deployment.yaml` so Argo CD rolls the exact image.

## Deploy

Backstage published this repository and dispatched the control-plane alta
workflow. After the control-plane PR is merged, Argo CD applies the workload.
Endpoint: `https://llm-microservice.local`.
