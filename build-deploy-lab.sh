#!/bin/bash

set -e

export IMAGE="registry.k8s.lab/custom-self-adapter-operator"
export TAG=v0.1.$(date +%s)

make generate
make manifests
kubebuilder edit --plugins=helm/v1-alpha

make docker-build IMG=$IMAGE:$TAG
make docker-push IMG=$IMAGE:$TAG

envsubst < dist/chart/values.yaml > dist/chart/values-lab.yaml

helm -n csa upgrade --install --create-namespace csa dist/chart -f dist/chart/values-lab.yaml