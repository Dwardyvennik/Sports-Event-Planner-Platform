FROM golang:1.25-alpine AS builder

ARG SERVICE=notification-service
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o /out/service ./services/${SERVICE}/cmd/${SERVICE}

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/service /service
USER nonroot:nonroot
ENTRYPOINT ["/service"]
