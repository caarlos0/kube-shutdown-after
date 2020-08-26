#!/bin/sh
set -e

goreleaser --rm-dist

helm lint charts/kube-shutdown-after
mkdir -p pkg

gcloud config set account carlos@carlosbecker.com
gcloud config set project carlos-kube

gsutil rsync gs://carlos-charts pkg
helm package charts/kube-shutdown-after --destination pkg
helm repo index pkg/ --url "https://carlos-charts.storage.googleapis.com"
gsutil rsync pkg gs://carlos-charts
gsutil setmeta -h "Cache-Control:private,max-age=0,no-transform" gs://carlos-charts/index.yaml

helm repo update
