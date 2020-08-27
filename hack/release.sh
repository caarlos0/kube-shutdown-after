#!/bin/sh
set -e

goreleaser --rm-dist

./hack/helm-release.sh
