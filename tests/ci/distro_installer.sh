#!/bin/bash
set -x

set -e

sudo make package_online GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=dhi.io/golang:1.27.2-alpine3.24-dev@sha256:9a9160e3c89a37572e35271c65a2b38a2d1671c8ab0d6abee19ce3d6cecb9be9 COMPILETAG=compile_golangimage TRIVYFLAG=true EXPORTERFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
sudo make package_offline GOBUILDTAGS="include_oss include_gcs" VERSIONTAG=dev-gitaction PKGVERSIONTAG=dev-gitaction UIVERSIONTAG=dev-gitaction GOBUILDIMAGE=dhi.io/golang:1.27.2-alpine3.24-dev@sha256:9a9160e3c89a37572e35271c65a2b38a2d1671c8ab0d6abee19ce3d6cecb9be9 COMPILETAG=compile_golangimage TRIVYFLAG=true EXPORTERFLAG=true HTTPPROXY= PULL_BASE_FROM_DOCKERHUB=false
