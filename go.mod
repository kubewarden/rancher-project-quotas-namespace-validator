module github.com/kubewarden/rancher-project-quotas-namespace-validator

go 1.19

replace github.com/go-openapi/strfmt => github.com/kubewarden/strfmt v0.1.2

require (
	github.com/kubewarden/k8s-objects v1.24.0-kw3
	github.com/kubewarden/policy-sdk-go v0.2.5-0.20230307171214-0d6ab1698c55
	github.com/mailru/easyjson v0.7.7
	github.com/wapc/wapc-guest-tinygo v0.3.3
	gopkg.in/inf.v0 v0.9.1
)

require (
	github.com/go-openapi/strfmt v0.0.0-00010101000000-000000000000 // indirect
	github.com/josharian/intern v1.0.0 // indirect
)
