# syntax=docker/dockerfile:1
#
# Runtime image for maal_business. Deliberately the same shape as ledger's —
# binaries at /app/<name>, WORKDIR /app, USER 1001 — because L2's go-service
# chart encodes that layout: workloads[*].command is ["/app/worker"], config is
# mounted over /app/config, and podSecurityContext pins runAsUser 1001.
#
# Only cmd/worker is built. cmd/starter kicks off a workflow and exits and
# cmd/local is a dev harness; neither is a long-running process, so neither
# belongs in a Deployment (see L2/envs/dev-local/maal-business.yaml).

FROM registry.access.redhat.com/ubi10/go-toolset:1.26 AS build
USER 0
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# CGO_ENABLED=0: this service is pure Go — Temporal SDK, pgx, cleanenv — so a
# static binary costs nothing and drops the runtime's glibc coupling. ledger
# needs CGO for its TigerBeetle client; this does not.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker


FROM registry.access.redhat.com/ubi10/ubi-minimal:10.2 AS runtime
WORKDIR /app
COPY --from=build /out/worker       /app/worker
# Shipped so the image runs standalone. In-cluster the go-service chart mounts a
# ConfigMap over /app/config, which replaces these.
COPY --from=build /src/config/*.yml /app/config/

USER 1001

ENTRYPOINT ["/app/worker"]