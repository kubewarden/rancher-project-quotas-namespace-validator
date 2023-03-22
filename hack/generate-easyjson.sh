#!/bin/sh

set -e

TMP_DIR=tmpdir-easyjson

echo "easyjson code generator doesn't work inside of the 'main' package"
echo creating a temporary go module...
echo

rm -rf $TMP_DIR
mkdir $TMP_DIR
cd $TMP_DIR

go mod init tmp-easyjson

echo "replace github.com/go-openapi/strfmt => github.com/kubewarden/strfmt v0.1.2" >> go.mod

sed -e 's/package main/package tmp_easyjson/g' ../types.go > types.go

echo Install easyjson
go get github.com/mailru/easyjson && go install github.com/mailru/easyjson/...@latest
go get github.com/mailru/easyjson/gen
go get github.com/mailru/easyjson/jlexer
go get github.com/mailru/easyjson/jwriter

echo Install Kubernetes objects types
go get github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1

echo Generate easyjson helper files
easyjson -all types.go

sed -e 's/package tmp_easyjson/package main/g' types_easyjson.go > ../types_easyjson.go

cd -
rm -rf $TMP_DIR

echo "Don't forget to run \`go mod tidy\`"
