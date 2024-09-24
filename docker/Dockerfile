#
# DEV - BASE
FROM golang:1.22 AS api-dev

RUN mkdir -p /app/docker/bin
WORKDIR /app

ENV DEBIAN_FRONTEND noninteractive
RUN set -eux; \
	apt -qq update && \
    apt install -qq -y --no-install-recommends \
        bash \
        inotify-tools && \
    apt -qq clean

RUN go install honnef.co/go/tools/cmd/staticcheck@latest

COPY docker/init/api-dev /init
ENTRYPOINT ["/init", "api"]


#
# BASE
FROM api-dev AS base

# copy and download go mod dependencies
COPY go/go.mod go/go.sum ./
RUN go mod download

# copy src
COPY go/ .

#
# BUILD
FROM base AS build

# test and lint before building
RUN go test -timeout 60s ./... && \
    staticcheck ./...

# build service
RUN GOARCH=amd64 CGO_ENABLED=1 GOOS=linux \
	go build -a -ldflags '-extldflags "-static"' -ldflags="-w -s" -o api nathejk.dk/cmd/api


#
# PROD
FROM alpine:3.14.1 AS prod

ARG CI_VERSION
ENV SENTRY_RELEASE=$CI_VERSION

RUN set -eux; \
	apk add --update --no-cache \
	    bash \
	    coreutils \
	    libc6-compat \
	    ca-certificates \
        tzdata \
        && \
	rm -rf /tmp/* /var/cache/apk/*

WORKDIR /app
COPY --from=build /app/api /diplom-api
COPY go/public /app/public
COPY go/assets /app/assets
COPY docker/bin/init /init

#HEALTHCHECK --interval=30s --timeout=15s --start-period=900s --retries=3 CMD test -f /tmp/healthy
EXPOSE 80
ENTRYPOINT ["/init"]
CMD ["/diplom-api"]
