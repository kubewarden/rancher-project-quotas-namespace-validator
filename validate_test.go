package main

import (
	"encoding/json"
	"fmt"
	"testing"

	corev1 "github.com/kubewarden/k8s-objects/api/core/v1"
	metav1 "github.com/kubewarden/k8s-objects/apimachinery/pkg/apis/meta/v1"
	kubewarden "github.com/kubewarden/policy-sdk-go"
	"github.com/kubewarden/policy-sdk-go/pkg/capabilities/kubernetes"
	kubewarden_protocol "github.com/kubewarden/policy-sdk-go/protocol"
	kubewarden_testing "github.com/kubewarden/policy-sdk-go/testing"
	easyjson "github.com/mailru/easyjson"
)

func TestValidation(t *testing.T) {
	cases := []struct {
		desc                   string
		namespaceResourceQuota NamespaceResourceQuota
		projectResourceQuota   ProjectResourceQuota
		isValid                bool
	}{
		{
			"valid",
			NamespaceResourceQuota{
				Limit: ResourceQuotaLimit{
					Pods: "10",
				},
			},
			ProjectResourceQuota{
				Limit: ResourceQuotaLimit{
					Pods: "20",
				},
				UsedLimit: ResourceQuotaLimit{
					Pods: "2",
				},
			},
			true,
		},
		{
			"not valid",
			NamespaceResourceQuota{
				Limit: ResourceQuotaLimit{
					Pods: "10",
				},
			},
			ProjectResourceQuota{
				Limit: ResourceQuotaLimit{
					Pods: "20",
				},
				UsedLimit: ResourceQuotaLimit{
					Pods: "18",
				},
			},
			false,
		},
	}

	for _, tc := range cases {
		settings := Settings{}

		projectID := "proj-id"
		projectNs := "proj-ns"
		annotations := make(map[string]string)
		annotations[RancherProjectIDAnnotation] = fmt.Sprintf("%s:%s", projectNs, projectID)

		namespaceResourceQuotaJSON, err := json.Marshal(tc.namespaceResourceQuota)
		if err != nil {
			t.Errorf("cannot marshal namespaceResourceQuota to JSON: %v", err)
		}
		annotations[RancherResourceQuotaAnnotation] = string(namespaceResourceQuotaJSON)

		namespace := corev1.Namespace{
			Metadata: &metav1.ObjectMeta{
				Name:        "test-ns",
				Annotations: annotations,
			},
		}

		wapcPayloadValidatorFn := func(payload []byte) error {
			req := kubernetes.GetResourceRequest{}
			if err := json.Unmarshal(payload, &req); err != nil {
				return fmt.Errorf("cannot unmarshal payload into GetResourceRequest: %w", err)
			}

			if projectID != req.Name {
				return fmt.Errorf("Looking for the wrong Project name: %s instead of %s", req.Name, projectID)
			}

			if projectNs != req.Namespace {
				return fmt.Errorf("Looking for the wrong Project namespace: %s instead of %s", req.Namespace, projectNs)
			}

			if req.APIVersion != RancherProjectAPIVersion {
				return fmt.Errorf("Wrong APIVersion for project: %s", req.APIVersion)
			}

			if req.Kind != RancherProjectKind {
				return fmt.Errorf("Wrong Kind for project: %s", req.Kind)
			}

			return nil
		}

		project := Project{
			Metadata: &metav1.ObjectMeta{
				Name:      projectID,
				Namespace: projectNs,
			},
			Spec: &ProjectSpec{
				DisplayName:   "a project",
				Description:   "something used by the tests",
				ResourceQuota: &tc.projectResourceQuota,
			},
		}
		projectJSON, err := json.Marshal(&project)
		if err != nil {
			t.Errorf("cannot create mock client with a Project as payload: %v", err)
		}

		mockWapcClient = &MockWapcClient{
			Err:                 nil,
			PayloadResponse:     projectJSON,
			Operation:           "get_resource",
			PayloadValidationFn: &wapcPayloadValidatorFn,
		}

		payload, err := kubewarden_testing.BuildValidationRequest(&namespace, &settings)
		if err != nil {
			t.Errorf("Unexpected error: %+v", err)
		}

		responsePayload, err := validate(payload)
		if err != nil {
			t.Errorf("Unexpected error: %+v", err)
		}

		var response kubewarden_protocol.ValidationResponse
		if err := easyjson.Unmarshal(responsePayload, &response); err != nil {
			t.Errorf("Unexpected error: %+v", err)
		}

		if !response.Accepted && tc.isValid {
			message := "no message set"
			if response.Message != nil {
				message = *response.Message
			}
			t.Errorf("%s - unexpected rejection: %v", tc.desc, message)
		}

		if response.Accepted && !tc.isValid {
			t.Errorf("%s - should have been rejected", tc.desc)
		}
	}
}

func TestFindProject(t *testing.T) {
	aListOfStrings := []string{"not", "a", "project"}
	aListOfStringsJSON, err := json.Marshal(&aListOfStrings)
	if err != nil {
		t.Errorf("cannot create mock client with a list of strings as payload: %v", err)
	}
	mockClientWithAWrongPayload := MockWapcClient{
		PayloadResponse: aListOfStringsJSON,
		Err:             nil,
		Operation:       "get_resource",
	}

	aProject := Project{
		Metadata: &metav1.ObjectMeta{
			Name:      "a-project",
			Namespace: "a-namespace",
		},
		Spec: &ProjectSpec{
			DisplayName: "a project",
			Description: "something used by the tests",
		},
	}
	aProjectJSON, err := json.Marshal(&aProject)
	if err != nil {
		t.Errorf("cannot create mock client with a Project as payload: %v", err)
	}
	mockClientWithGoodPayload := MockWapcClient{
		PayloadResponse: aProjectJSON,
		Err:             nil,
		Operation:       "get_resource",
	}

	cases := []struct {
		desc           string
		mockWapcClient MockWapcClient
		expectError    *LookupError
	}{
		{
			"No project found",
			MockWapcClient{
				Err:             nil,
				PayloadResponse: []byte{},
				Operation:       "get_resource",
			},
			&LookupError{
				StatusCode: kubewarden.Code(404),
				Message:    kubewarden.Message("not relevant"),
			},
		},
		{
			"waPC host error",
			MockWapcClient{
				Err:             fmt.Errorf("something went wrong with waPC host"),
				PayloadResponse: []byte{},
				Operation:       "get_resource",
			},
			&LookupError{
				StatusCode: kubewarden.Code(500),
				Message:    kubewarden.Message("not relevant"),
			},
		},
		{
			"cannot unmarshal project",
			mockClientWithAWrongPayload,
			&LookupError{
				StatusCode: kubewarden.Code(500),
				Message:    kubewarden.Message("not relevant"),
			},
		},
		{
			"project found",
			mockClientWithGoodPayload,
			nil,
		},
	}

	for _, tc := range cases {
		mockWapcClient = &tc.mockWapcClient

		_, lookupErr := findProject("prj1", "local")

		if lookupErr == nil && tc.expectError != nil {
			t.Errorf("%s - didn't get an error as expected", tc.desc)
		}

		if lookupErr != nil && tc.expectError == nil {
			t.Errorf("%s - was not expected to fail with error: %v", tc.desc, lookupErr)
		}

		if lookupErr != nil && tc.expectError != nil {
			if lookupErr.StatusCode != tc.expectError.StatusCode {
				t.Errorf("%s - got the wrong status code. Expecting %d, got %d instead", tc.desc, tc.expectError.StatusCode, lookupErr.StatusCode)
			}
		}
	}
}

func TestParseProjectAnnotation(t *testing.T) {
	cases := []struct {
		desc        string
		annotation  string
		projectID   string
		projectNs   string
		expectError bool
	}{
		{
			"empty string",
			"",
			"",
			"",
			true,
		},
		{
			"projectNs empty",
			":ID",
			"",
			"ID",
			true,
		},
		{
			"projectID empty",
			"NS:",
			"NS",
			"",
			true,
		},
		{
			"too many chunks",
			"NS:ID:not expected",
			"NS",
			"ID",
			true,
		},
		{
			"all good",
			"NS:ID",
			"NS",
			"ID",
			false,
		},
	}

	for _, tc := range cases {
		projectID, projectNs, err := parseProjectIDAnnotation(tc.annotation)
		if tc.expectError && err == nil {
			t.Errorf("%s - was supposed to fail", tc.desc)
		}

		if !tc.expectError && err != nil {
			t.Errorf("%s - was not supposed to fail. Got this err: %v", tc.desc, err)
		}

		if !tc.expectError && err == nil {
			if projectID != tc.projectID {
				t.Errorf("%s - wrong projectID. Got %s instead of %s", tc.desc, projectID, tc.projectID)
			}

			if projectNs != tc.projectNs {
				t.Errorf("%s - wrong projectNs. Got %s instead of %s", tc.desc, projectNs, tc.projectNs)
			}
		}
	}
}
