# syntax=docker/dockerfile:1

FROM golang:1.27.2-alpine AS build

WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Atlas releases are distributed as binaries, not versioned Go modules.
ADD --checksum=sha256:10d7913e3dce43ab99b8d71534a4cbadaf11a16dc293adf3b91d10e83a0ac70b \
    https://release.ariga.io/atlas/atlas-community-linux-amd64-v1.3.0 /go/bin/atlas

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go install github.com/go-task/task/v3/cmd/task@v3.54.0 \
    && chmod +x /go/bin/atlas

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
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
