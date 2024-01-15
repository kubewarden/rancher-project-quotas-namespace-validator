module github.com/kubewarden/rancher-project-quotas-namespace-validator

go 1.20

replace github.com/go-openapi/strfmt => github.com/kubewarden/strfmt v0.1.3

require (
	github.com/kubewarden/k8s-objects v1.27.0-kw4
	github.com/kubewarden/policy-sdk-go v0.5.2
	github.com/wapc/wapc-guest-tinygo v0.3.3
	gopkg.in/inf.v0 v0.9.1
)

require github.com/go-openapi/strfmt v0.21.5 // indirect
