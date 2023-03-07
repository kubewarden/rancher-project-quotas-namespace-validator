#!/bin/bash
set -e

echo "Destroy k3d cluster"
k3d cluster delete kw-policy-e2e
