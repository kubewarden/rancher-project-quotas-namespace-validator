//go:build wasi
// +build wasi

// This provides a way to mock the
// note well: we have to use the tinygo wasi target, because the wasm one is
// meant to be used inside of the browser

package main

import (
	capabilities "github.com/kubewarden/policy-sdk-go/pkg/capabilities"
)

func getWapcHost() capabilities.Host {
	return capabilities.NewHost()
}
