# syntax=docker/dockerfile:1

FROM golang:1.26.6-alpine AS build

WORKDIR /src
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go install github.com/go-task/task/v3/cmd/task@v3.53.1 \
    && go install ariga.io/atlas/cmd/atlas@v1.3.0 \
    && task build-linux-all

FROM alpine:3.21

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
