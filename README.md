# metrics

## Run

```
docker-compose up -d
```

```bash
watch -n 0.01 "curl localhost:5555/200"
```

```bash
watch -n 1    "curl localhost:5555/500"
```

 - Visit [localhost:9090](localhost:9090) for Prometheus.
 - Visit [localhost:3000](localhost:3000) for Grafana.
