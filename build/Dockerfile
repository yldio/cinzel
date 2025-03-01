# syntax=docker/dockerfile:1
FROM golang:1.24.0-alpine AS builder

WORKDIR /app

ARG version
COPY . .

RUN go build -ldflags "-s -w -X 'main.version=${version}'" -o /bin/acto ./acto.go

FROM scratch
COPY --from=builder /bin/acto /bin/acto

ENTRYPOINT ["/bin/acto"]
