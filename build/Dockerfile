# syntax=docker/dockerfile:1
# Copyright 2026 YLD Limited
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.26-alpine AS builder

WORKDIR /app

ARG version
COPY . .

RUN go build -ldflags "-s -w -X 'main.version=${version}'" -o /bin/cinzel ./cinzel.go

FROM scratch
COPY --from=builder /bin/cinzel /bin/cinzel

ENTRYPOINT ["/bin/cinzel"]
