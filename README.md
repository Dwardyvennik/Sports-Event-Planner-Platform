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

Prometheus runs on `localhost:9090`, Grafana on `localhost:3000`, NATS monitoring on `localhost:8222`, and the API gateway on `localhost:8080`.

## API Gateway Routes

Public:

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`

Protected with `Authorization: Bearer <token>`:

- `GET /v1/auth/me`
- `POST /v1/events`
- `GET /v1/events/:id`
- `GET /v1/events`
- `PUT /v1/events/:id`
- `DELETE /v1/events/:id`
- `POST /v1/events/:id/join`
- `POST /v1/events/:id/leave`
- `POST /v1/notifications`
- `POST /v1/events/:id/reminders`
- `GET /v1/users/:id/notifications`

Operational:

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`

The gateway uses Gin for REST routing and typed gRPC clients for service-to-service calls. The current stubs are intentionally minimal and use the shared JSON gRPC codec in `pkg/grpcx` so the skeleton compiles and runs without requiring local protobuf codegen plugins.
