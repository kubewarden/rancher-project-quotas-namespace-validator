package main

import (
	kubewarden "github.com/kubewarden/policy-sdk-go"
)

func validateSettings(payload []byte) ([]byte, error) {
	return kubewarden.AcceptSettings()
}
