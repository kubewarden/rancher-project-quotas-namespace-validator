package main

import (
	"fmt"
	"strings"

	"github.com/kubewarden/rancher-project-quotas-namespace-validator/resource"
)

// QuantityParseError is a custom error raised when a string cannot be
// parsed to be be a resource.Quantity
type QuantityParseError struct {
	Message string
	Err     error
}

func (e *QuantityParseError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

// NamespaceRequestExceedsAvailabilityError a custom error raised when
// a namespace requests more resources than available
type NamespaceRequestExceedsAvailabilityError struct {
	requested string
	available string
}

func (e *NamespaceRequestExceedsAvailabilityError) Error() string {
	return fmt.Sprintf("Namespace requested limit exceeds the availability of the project resource: requested %s, available %s", e.requested, e.available)
}

// Compares the amount of resources requested by a namespace against the
// availability of a project.
//
// Returns an error when one of these situation occurs:
//   - The given strings cannot be converted to a Kubernetes Quantity
//   - The project is already out of resources
//   - The namespace has requested too much of a resource compared to the availability
//     of the project
func checkLimitVsAvailableQuota(nsLimit, prjLimit, prjUsed string) error {
	if nsLimit == "" {
		nsLimit = "0"
	}
	nsLimitQuantity, err := resource.ParseQuantity(nsLimit)
	if err != nil {
		return &QuantityParseError{
			Message: "Cannot convert namespace limit to quantity",
			Err:     err,
		}
	}

	if prjLimit == "" {
		prjLimit = "0"
	}
	prjLimitQuantity, err := resource.ParseQuantity(prjLimit)
	if err != nil {
		return &QuantityParseError{
			Message: "Cannot convert project limit to quantity",
			Err:     err,
		}
	}

	if prjUsed == "" {
		prjUsed = "0"
	}
	prjUsedQuantity, err := resource.ParseQuantity(prjUsed)
	if err != nil {
		return &QuantityParseError{
			Message: "Cannot convert project used quota to quantity",
			Err:     err,
		}
	}

	prjAvailableQuantity := prjLimitQuantity.DeepCopy()
	prjAvailableQuantity.Sub(prjUsedQuantity)

	if nsLimitQuantity.Cmp(prjAvailableQuantity) > 0 {
		return &NamespaceRequestExceedsAvailabilityError{
			requested: nsLimitQuantity.String(),
			available: prjAvailableQuantity.String(),
		}
	}

	return nil
}

func validateQuotas(project *Project, nsLimits *ResourceQuotaLimit) error {
	if project.Spec.ResourceQuota == nil {
		return nil
	}

	if nsLimits == nil {
		nsLimits = &ResourceQuotaLimit{}
	}

	errors := []error{}

	// Pods
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.Pods, project.Spec.ResourceQuota.Limit.Pods, project.Spec.ResourceQuota.UsedLimit.Pods); err != nil {
		errors = append(errors, fmt.Errorf("Pods limit: %w", err))
	}

	// Services
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.Services, project.Spec.ResourceQuota.Limit.Services, project.Spec.ResourceQuota.UsedLimit.Services); err != nil {
		errors = append(errors, fmt.Errorf("Services limit: %w", err))
	}

	// ReplicationControllers
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.ReplicationControllers, project.Spec.ResourceQuota.Limit.ReplicationControllers, project.Spec.ResourceQuota.UsedLimit.ReplicationControllers); err != nil {
		errors = append(errors, fmt.Errorf("ReplicationControllers limit: %w", err))
	}

	// Secrets
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.Secrets, project.Spec.ResourceQuota.Limit.Secrets, project.Spec.ResourceQuota.UsedLimit.Secrets); err != nil {
		errors = append(errors, fmt.Errorf("Secrets limit: %w", err))
	}

	// ConfigMaps
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.ConfigMaps, project.Spec.ResourceQuota.Limit.ConfigMaps, project.Spec.ResourceQuota.UsedLimit.ConfigMaps); err != nil {
		errors = append(errors, fmt.Errorf("ConfigMaps limit: %w", err))
	}

	// PersistentVolumeClaims
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.PersistentVolumeClaims, project.Spec.ResourceQuota.Limit.PersistentVolumeClaims, project.Spec.ResourceQuota.UsedLimit.PersistentVolumeClaims); err != nil {
		errors = append(errors, fmt.Errorf("PersistentVolumeClaims limit: %w", err))
	}

	// ServicesNodePorts
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.ServicesNodePorts, project.Spec.ResourceQuota.Limit.ServicesNodePorts, project.Spec.ResourceQuota.UsedLimit.ServicesNodePorts); err != nil {
		errors = append(errors, fmt.Errorf("ServicesNodePorts limit: %w", err))
	}

	// ServicesLoadBalancers
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.ServicesLoadBalancers, project.Spec.ResourceQuota.Limit.ServicesLoadBalancers, project.Spec.ResourceQuota.UsedLimit.ServicesLoadBalancers); err != nil {
		errors = append(errors, fmt.Errorf("ServicesLoadBalancers limit: %w", err))
	}

	// RequestsCPU
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.RequestsCPU, project.Spec.ResourceQuota.Limit.RequestsCPU, project.Spec.ResourceQuota.UsedLimit.RequestsCPU); err != nil {
		errors = append(errors, fmt.Errorf("RequestsCPU limit: %w", err))
	}

	// RequestsMemory
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.RequestsMemory, project.Spec.ResourceQuota.Limit.RequestsMemory, project.Spec.ResourceQuota.UsedLimit.RequestsMemory); err != nil {
		errors = append(errors, fmt.Errorf("RequestsMemory limit: %w", err))
	}

	// RequestsStorage
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.RequestsStorage, project.Spec.ResourceQuota.Limit.RequestsStorage, project.Spec.ResourceQuota.UsedLimit.RequestsStorage); err != nil {
		errors = append(errors, fmt.Errorf("RequestsStorage limit: %w", err))
	}

	// LimitsCPU
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.LimitsCPU, project.Spec.ResourceQuota.Limit.LimitsCPU, project.Spec.ResourceQuota.UsedLimit.LimitsCPU); err != nil {
		errors = append(errors, fmt.Errorf("LimitsCPU limit: %w", err))
	}

	// LimitsMemory
	// nolint: staticcheck // keep k8s resources capitalized
	if err := checkLimitVsAvailableQuota(nsLimits.LimitsMemory, project.Spec.ResourceQuota.Limit.LimitsMemory, project.Spec.ResourceQuota.UsedLimit.LimitsMemory); err != nil {
		errors = append(errors, fmt.Errorf("LimitsMemory limit: %w", err))
	}

	if len(errors) == 0 {
		return nil
	}

	errorMsgs := []string{}
	for _, err := range errors {
		errorMsgs = append(errorMsgs, err.Error())
	}

	return fmt.Errorf("%s", strings.Join(errorMsgs, ", "))
}
