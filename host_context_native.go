//go:build !wasi

// This is native, as in non-wasi code, that is meant to be used only by unit tests

package main

import (
	"fmt"

	capabilities "github.com/kubewarden/policy-sdk-go/pkg/capabilities"
)

// MockWapcClient is implements the `host.WapcClient` interface.
// It's purpose is to be used by the unit tests of policies that leverage
// host capabilities
type MockWapcClient struct {
	Err                 error
	PayloadResponse     []byte
	Operation           string
	PayloadValidationFn *func([]byte) error
}

// HostCall implements the `host.WapcClient` interface
func (m *MockWapcClient) HostCall(binding, namespace, operation string, payload []byte) (response []byte, err error) {
	if binding != "kubewarden" {
		return []byte{}, fmt.Errorf("wrong binding: %s", binding)
	}
	if namespace != "kubernetes" {
		return []byte{}, fmt.Errorf("wrong namespace: %s", namespace)
	}
	if m.Operation != operation {
		return []byte{}, fmt.Errorf("wrong operation: got %s instead of %s", operation, m.Operation)
	}
	if m.PayloadValidationFn != nil {
		validationFn := *m.PayloadValidationFn
		if err := validationFn(payload); err != nil {
			return []byte{}, fmt.Errorf("wapc payload validation failed: %v", err)
		}
	}

	return m.PayloadResponse, m.Err
}

// This is the mock WapcClient use inside of the tests
var mockWapcClient *MockWapcClient

// This is the non-wasi implementation of the helper function
// that returns a waPC Host to be used.
// This is required to inject our mock client
func getWapcHost() capabilities.Host {
	return capabilities.Host{
		Client: mockWapcClient,
	}
}
