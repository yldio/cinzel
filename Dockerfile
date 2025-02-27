# syntax=docker/dockerfile:1
FROM golang:1.23.5-alpine AS builder

WORKDIR /app

ARG version
COPY go.mod go.sum ./
COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./service ./service

RUN go build -ldflags "-s -w -X 'main.version=${version}'" -o /bin/acto ./cmd/acto/main.go

FROM scratch
COPY --from=builder /bin/acto /bin/acto

ENTRYPOINT ["/bin/acto"]
