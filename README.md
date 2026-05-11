# Sports Event Planner Platform

Clean Architecture microservices starter for a university sports event planner platform.

## Services

- `auth-service`
- `event-service`
- `notification-service`
- `api-gateway`

## Local Workflow

```bash
make tidy
make test
make docker-up
make migrate-up
```

Prometheus runs on `localhost:9090`, Grafana on `localhost:3000`, RabbitMQ management on `localhost:15672`, and the API gateway on `localhost:8080`.
