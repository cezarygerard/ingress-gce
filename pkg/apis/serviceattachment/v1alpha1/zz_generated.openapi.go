// +build !ignore_autogenerated

/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachment":       schema_pkg_apis_serviceattachment_v1alpha1_ServiceAttachment(ref),
		"k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachmentSpec":   schema_pkg_apis_serviceattachment_v1alpha1_ServiceAttachmentSpec(ref),
		"k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachmentStatus": schema_pkg_apis_serviceattachment_v1alpha1_ServiceAttachmentStatus(ref),
	}
}

func schema_pkg_apis_serviceattachment_v1alpha1_ServiceAttachment(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceAttachment represents a Service Attachment associated with a service/ingress/gateway class",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachmentSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachmentStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta", "k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachmentSpec", "k8s.io/ingress-gce/pkg/apis/serviceattachment/v1alpha1.ServiceAttachmentStatus"},
	}
}

func schema_pkg_apis_serviceattachment_v1alpha1_ServiceAttachmentSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceAttachmentSpec is the spec for a ServiceAttachment resource",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"connectionPreference": {
						SchemaProps: spec.SchemaProps{
							Description: "ConnectionPreference determines how consumers are accepted. Only allowed value is `acceptAutomatic`.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"natSubnets": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-type": "atomic",
							},
						},
						SchemaProps: spec.SchemaProps{
							Description: "NATSubnets contains the list of subnet names for PSC",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
					"resourceRef": {
						SchemaProps: spec.SchemaProps{
							Description: "ResourceRef is the reference to the K8s resource that created the forwarding rule Only Services can be used as a reference",
							Ref:         ref("k8s.io/api/core/v1.TypedLocalObjectReference"),
						},
					},
				},
				Required: []string{"connectionPreference", "natSubnets", "resourceRef"},
			},
		},
		Dependencies: []string{
			"k8s.io/api/core/v1.TypedLocalObjectReference"},
	}
}

func schema_pkg_apis_serviceattachment_v1alpha1_ServiceAttachmentStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ServiceAttachmentStatus is the status for a ServiceAttachment resource",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"serviceAttachmentURL": {
						SchemaProps: spec.SchemaProps{
							Description: "ServiceAttachmentURL is the URL for the GCE Service Attachment resource",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"forwardingRuleURL": {
						SchemaProps: spec.SchemaProps{
							Description: "ForwardingRuleURL is the URL to the GCE Forwarding Rule resource the Service Attachment points to",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
				Required: []string{"serviceAttachmentURL", "forwardingRuleURL"},
			},
		},
	}
}
