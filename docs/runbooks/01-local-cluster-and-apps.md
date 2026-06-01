# Runbook 01 — Local Cluster & Apps

> Goal: deploy an app to the local cluster, reach it, and read accurate status.
> Verifies Phase 1.

## Prereqs

- Runbook 00 completed: cluster up, tools installed, `/etc/hosts` mapped.

## Steps

```bash
# 1. Confirm the cluster
bin/labctl status

# 2. Build and deploy go-api
make build  APP_NAME=go-api          # or: bin/labctl app build go-api
make deploy APP_NAME=go-api          # or: bin/labctl app deploy go-api

# 3. Reach it
curl http://go-api.k3d.local/health  # -> 200 / healthy
curl http://go-api.k3d.local/ready   # -> 200 when ready
curl http://go-api.k3d.local/metrics | grep http_requests_total

# 4. Watch pods / logs
kubectl get pods -n go-api -w
kubectl logs -n go-api -l app=go-api -f

# 5. Deploy the second app (Redis-backed)
bin/labctl service up redis          # echo-server depends on Redis
make deploy APP_NAME=echo-server
curl http://echo-server.k3d.local/health
```

## Expected

- `/health` returns `{"status":"ok"}` (plus a `redis` key when Redis is configured).
- `/ready` always returns 200 — it is not gated on Redis.
- `bin/labctl status` shows the correct **Kubernetes version** parsed from `kubectl version -o json`.
- echo-server `/ready` passes even when Redis is temporarily down (Task 007).
- Posting a body >1 MiB to `/echo` or `/cache` returns HTTP 413 (Task 016).
- `labctl scenario up <name>` fails fast with a clear error if the runtime or a
  prerequisite is missing — before any component install starts (Task 002).
- API calls with path params like `../etc` or `app;rm` return HTTP 400 (Task 008).

### Verify scenario preflight

```bash
# Should fail fast — k3d cluster does not support aks-only scenarios
bin/labctl scenario up gitops-cicd   # check the runtimes field in scenario.yaml

# Should fail if go-api is not deployed
bin/labctl scenario up observability-sre
```

### Verify API input validation

```bash
curl -s -X POST http://localhost:3939/api/apps/..%2F..%2Fetc%2Fpasswd/deploy | jq .
# -> {"error":"invalid app name ..."}  HTTP 400
```

### Verify body size limit

```bash
dd if=/dev/urandom bs=2M count=1 | curl -s -X POST -d @- http://echo-server.k3d.local/echo
# -> {"error":"request body too large"}  HTTP 413
```

## Alternative access (no /etc/hosts)

```bash
kubectl port-forward -n go-api svc/go-api 8080:8080
curl http://localhost:8080/health
```

## Helm value profiles to try

```bash
make deploy APP_NAME=go-api HELM_VALUES=values-dev.yaml         # 1 replica
make deploy APP_NAME=go-api HELM_VALUES=values-prod-like.yaml   # 3 replicas + HPA
make deploy APP_NAME=go-api HELM_VALUES=values-test.yaml        # readiness-failure demo
```

## Cleanup

```bash
make destroy-app APP_NAME=echo-server
make destroy-app APP_NAME=go-api
bin/labctl service down redis
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `ImagePullBackOff` on k3d | Image not imported: `k3d image import go-api:latest -c <cluster>`. |
| Wrong k8s version in status | Task 015 (version parsing). |
| Pod count looks wrong | Task 020 (JSON pod parsing). |
| echo-server never Ready | Redis missing (`labctl service up redis`) or Task 007. |
| Ingress 404 | Check `/etc/hosts`, then `kubectl get ingress -n go-api`. |
