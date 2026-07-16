#!/bin/bash
set -x

set -e

sudo make package_online GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=dhi.io/golang:1.26.5-alpine3.24-dev@sha256:711ea0b8f09f549c50f2f550dc26859d3e6441ca11d5640caecf69c29a862f0c COMPILETAG=compile_golangimage TRIVYFLAG=true EXPORTERFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
sudo make package_offline GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=dhi.io/golang:1.26.5-alpine3.24-dev@sha256:711ea0b8f09f549c50f2f550dc26859d3e6441ca11d5640caecf69c29a862f0c COMPILETAG=compile_golangimage TRIVYFLAG=true EXPORTERFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
