package main

import (
	"encoding/json"
	"fmt"
	"strings"

	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	meta_v1 "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	"github.com/kubewarden/policy-sdk-go/pkg/capabilities"
	"github.com/kubewarden/policy-sdk-go/pkg/capabilities/kubernetes"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
)

const (
	// RancherProjectIDAnnotation is the annotation used by Rancher Manager inside of
	// Namespace object that defines which Project the Namespace belongs to.
	RancherProjectIDAnnotation = "field.cattle.io/projectId"

	// RancherResourceQuotaAnnotation is the annotation used by Rancher
	// Manager inside of a Namespace object.
	// The value is a JSON object holding the `ResourceQuotaLimit` of the
	// Namespace
	RancherResourceQuotaAnnotation = "field.cattle.io/resourceQuota"

	// RancherProjectAPIVersion is the Kubernetes Group + Version used by the Project resources
	RancherProjectAPIVersion = "management.cattle.io/v3"

	// RancherProjectKind is the Kubernetes Kind used by the Project resources
	RancherProjectKind = "Project"
)

var host = capabilities.NewHost()

func validate(payload []byte) ([]byte, error) {
	// Create a ValidationRequest instance from the incoming payload
	validationRequest := kubewarden_protocol.ValidationRequest{}
	err := json.Unmarshal(payload, &validationRequest)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(400))
	}

	// Access the **raw** JSON that describes the object
	namespaceJSON := validationRequest.Request.Object

	// Try to create a Namespace instance using the RAW JSON we got from the
	// ValidationRequest.
	namespace := &corev1.Namespace{}
	if err := json.Unmarshal([]byte(namespaceJSON), namespace); err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(
				fmt.Sprintf("Cannot decode Namespace object: %s", err.Error())),
			kubewarden.Code(400))
	}

	nsMetadata := &meta_v1.ObjectMeta{}
	if namespace.Metadata != nil {
		nsMetadata = namespace.Metadata
	}

	projectIDAnnotation, found := nsMetadata.Annotations[RancherProjectIDAnnotation]
	if !found {
		return kubewarden.AcceptRequest()
	}

	projectNamespace, projectID, err := parseProjectIDAnnotation(projectIDAnnotation)
	if err != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(err.Error()),
			kubewarden.Code(400))
	}

	nsResourceQuota := NamespaceResourceQuota{}
	nsResourceQuotaRaw, found := nsMetadata.Annotations[RancherResourceQuotaAnnotation]
	if found {
		if err := json.Unmarshal([]byte(nsResourceQuotaRaw), &nsResourceQuota); err != nil {
			return kubewarden.RejectRequest(
				kubewarden.Message(
					fmt.Sprintf("Cannot decode NamespaceResourceQuota object: %s", err.Error())),
				kubewarden.Code(400))
		}
	}

	project, lookupError := findProject(projectID, projectNamespace)
	if lookupError != nil {
		return kubewarden.RejectRequest(
			lookupError.Message,
			lookupError.StatusCode)
	}

	validationErr := validateQuotas(&project, &nsResourceQuota.Limit)
	if validationErr != nil {
		return kubewarden.RejectRequest(
			kubewarden.Message(validationErr.Error()),
			kubewarden.NoCode)
	}

	return kubewarden.AcceptRequest()
}

// LookupError is a custom error that provides extra information
type LookupError struct {
	StatusCode kubewarden.Code
	Message    kubewarden.Message
}

func (l *LookupError) Error() string {
	return fmt.Sprintf("status %d: err %v", l.StatusCode, l.Message)
}

func findProject(projectID, projectNamespace string) (Project, *LookupError) {
	project := Project{}

	findPrjReq := kubernetes.GetResourceRequest{
		APIVersion:   RancherProjectAPIVersion,
		Kind:         RancherProjectKind,
		Name:         projectID,
		Namespace:    &projectNamespace,
		DisableCache: true,
	}

	projectRaw, err := kubernetes.GetResource(&host, findPrjReq)
	if err != nil {
		return project, &LookupError{
			Message:    kubewarden.Message(fmt.Sprintf("Error retrieving the Project: %v", err)),
			StatusCode: kubewarden.Code(500),
		}
	}

	if len(projectRaw) == 0 {
		return project, &LookupError{
			Message:    kubewarden.Message("Project not found"),
			StatusCode: kubewarden.Code(404),
		}
	}

	if err := json.Unmarshal(projectRaw, &project); err != nil {
		return project, &LookupError{
			Message:    kubewarden.Message(fmt.Sprintf("Cannot decode Project object: %s", err.Error())),
			StatusCode: kubewarden.Code(500),
		}
	}

	return project, nil
}

func parseProjectIDAnnotation(annotation string) (projectNamespace string, projectID string, err error) {
	chunks := strings.Split(annotation, ":")
	if len(chunks) != 2 {
		err = fmt.Errorf("cannot parse projectID annotation: wrong format")
		return
	}

	if len(chunks[0]) == 0 {
		err = fmt.Errorf("Project Namespace is empty")
		return
	}

	if len(chunks[1]) == 0 {
		err = fmt.Errorf("Project ID is empty")
		return
	}

	return chunks[0], chunks[1], nil
}
