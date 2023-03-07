#!/bin/bash
set -e

echo "Create k3d cluster"
# kubernetes < 1.25 is currently required by rancher
k3d cluster create -i "rancher/k3s:v1.24.10-k3s1" kw-policy-e2e

echo "Add cert-manager helm chart repository"
helm repo add jetstack https://charts.jetstack.io

echo "Add rancher-latest helm chart repository"
helm repo add rancher-latest https://releases.rancher.com/server-charts/latest

echo "Update local Helm chart repository cache"
helm repo update

echo "Install cert-manager CRDs"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.7.1/cert-manager.crds.yaml

echo "Install the cert-manager Helm chart"
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.7.1

echo "Install Rancher"
kubectl create namespace cattle-system
helm install rancher rancher-latest/rancher \
  --namespace cattle-system \
  --set hostname=127.0.0.1.sslip.io \
  --set bootstrapPassword=admin

echo "Waiting for the Project CRD to be created by Rancher"

timeout 10m bash -c 'until kubectl api-resources| grep projects; do sleep 10; done'

echo "Create Project used by e2e tests"
kubectl create ns local
kubectl apply -f test_data/project.yaml

echo "Waiting for Rancher to be fully operational"
sleep 30s
