# syntax=docker/dockerfile:1

FROM golang:1.27.2-alpine AS build

WORKDIR /src
# Bound package-level compile concurrency on small deployment builders.
# Override with --build-arg GO_BUILD_PARALLELISM=<n> when the builder has more capacity.
ARG GO_BUILD_PARALLELISM=2
ENV GOFLAGS=-p=${GO_BUILD_PARALLELISM}
RUN apk add --no-cache git

COPY go.mod go.sum ./
# Keep dependencies in a normal layer so the remote cache preserves them.
RUN go mod download

# Atlas releases are distributed as binaries, not versioned Go modules.
ADD --checksum=sha256:10d7913e3dce43ab99b8d71534a4cbadaf11a16dc293adf3b91d10e83a0ac70b \
    https://release.ariga.io/atlas/atlas-community-linux-amd64-v1.3.0 /go/bin/atlas

ADD --checksum=sha256:680859dbb4d881a9c72d4d9a8f510825450849af8567deacd7302c01124416fb \
    https://github.com/go-task/task/releases/download/v3.54.0/task_linux_amd64.tar.gz /tmp/task.tar.gz
RUN tar -xzf /tmp/task.tar.gz -C /go/bin task \
    && chmod +x /go/bin/atlas \
    && rm /tmp/task.tar.gz

COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    task build-all TARGET_OS=linux TARGET_ARCH=amd64

FROM alpine:3.24.2

RUN apk add --no-cache ca-certificates wget tzdata
WORKDIR /app

COPY --from=build /src/bin/server /app/server
COPY --from=build /src/bin/migrate /app/migrate
COPY --from=build /src/bin/worker /app/worker
COPY --from=build /src/config /app/config
COPY --from=build /src/migrations /app/migrations
COPY --from=build /src/public /app/public
COPY --from=build /go/bin/atlas /usr/local/bin/atlas

EXPOSE 8080
USER 65532:65532
ENTRYPOINT ["/app/server"]
