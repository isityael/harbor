#!/bin/sh
# Shared by both release images and the PR build check. Keep one pin list.
set -eu

# renovate: datasource=go depName=github.com/gorilla/handlers
GORILLA_HANDLERS_VERSION=v1.5.2
# renovate: datasource=go depName=github.com/miekg/dns
DNS_VERSION=v1.1.73
# renovate: datasource=go depName=github.com/prometheus/client_golang
PROMETHEUS_VERSION=v1.25.0
# renovate: datasource=go depName=golang.org/x/crypto
X_CRYPTO_VERSION=v0.58.0
# renovate: datasource=go depName=golang.org/x/net
X_NET_VERSION=v0.61.0
# renovate: datasource=go depName=golang.org/x/oauth2
OAUTH2_VERSION=v0.37.0
# renovate: datasource=go depName=golang.org/x/sys
X_SYS_VERSION=v0.49.0
# renovate: datasource=go depName=golang.org/x/text
X_TEXT_VERSION=v0.43.0
# renovate: datasource=go depName=google.golang.org/api
GOOGLE_API_VERSION=v0.308.0
# renovate: datasource=go depName=google.golang.org/grpc
GRPC_VERSION=v1.84.0
# renovate: datasource=go depName=gopkg.in/square/go-jose.v1
JOSE_VERSION=v1.1.2

go get \
    github.com/gorilla/handlers@"$GORILLA_HANDLERS_VERSION" \
    github.com/miekg/dns@"$DNS_VERSION" \
    github.com/prometheus/client_golang@"$PROMETHEUS_VERSION" \
    golang.org/x/crypto@"$X_CRYPTO_VERSION" \
    golang.org/x/net@"$X_NET_VERSION" \
    golang.org/x/oauth2/google@"$OAUTH2_VERSION" \
    golang.org/x/sys@"$X_SYS_VERSION" \
    golang.org/x/text@"$X_TEXT_VERSION" \
    google.golang.org/api@"$GOOGLE_API_VERSION" \
    google.golang.org/grpc@"$GRPC_VERSION" \
    gopkg.in/square/go-jose.v1@"$JOSE_VERSION"
go mod tidy
go mod vendor
python3 "$(dirname "$0")/patch-distribution-compat.py"
