//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023 The Kubernetes Authors.

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

package v1

import (
	common "k8s.io/kube-openapi/pkg/common"
	spec "k8s.io/kube-openapi/pkg/validation/spec"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BackendConfig":               schema_pkg_apis_backendconfig_v1_BackendConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BackendConfigSpec":           schema_pkg_apis_backendconfig_v1_BackendConfigSpec(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BypassCacheOnRequestHeader":  schema_pkg_apis_backendconfig_v1_BypassCacheOnRequestHeader(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CDNConfig":                   schema_pkg_apis_backendconfig_v1_CDNConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CacheKeyPolicy":              schema_pkg_apis_backendconfig_v1_CacheKeyPolicy(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.ConnectionDrainingConfig":    schema_pkg_apis_backendconfig_v1_ConnectionDrainingConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CustomRequestHeadersConfig":  schema_pkg_apis_backendconfig_v1_CustomRequestHeadersConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CustomResponseHeadersConfig": schema_pkg_apis_backendconfig_v1_CustomResponseHeadersConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.HealthCheckConfig":           schema_pkg_apis_backendconfig_v1_HealthCheckConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.IAPConfig":                   schema_pkg_apis_backendconfig_v1_IAPConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.LogConfig":                   schema_pkg_apis_backendconfig_v1_LogConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.NegativeCachingPolicy":       schema_pkg_apis_backendconfig_v1_NegativeCachingPolicy(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.OAuthClientCredentials":      schema_pkg_apis_backendconfig_v1_OAuthClientCredentials(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SecurityPolicyConfig":        schema_pkg_apis_backendconfig_v1_SecurityPolicyConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SessionAffinityConfig":       schema_pkg_apis_backendconfig_v1_SessionAffinityConfig(ref),
		"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SignedUrlKey":                schema_pkg_apis_backendconfig_v1_SignedUrlKey(ref),
	}
}

func schema_pkg_apis_backendconfig_v1_BackendConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"object"},
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
							Default: map[string]interface{}{},
							Ref:     ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BackendConfigSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Default: map[string]interface{}{},
							Ref:     ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BackendConfigStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BackendConfigSpec", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BackendConfigStatus"},
	}
}

func schema_pkg_apis_backendconfig_v1_BackendConfigSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BackendConfigSpec is the spec for a BackendConfig resource",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"iap": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.IAPConfig"),
						},
					},
					"cdn": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CDNConfig"),
						},
					},
					"securityPolicy": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SecurityPolicyConfig"),
						},
					},
					"timeoutSec": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					"connectionDraining": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.ConnectionDrainingConfig"),
						},
					},
					"sessionAffinity": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SessionAffinityConfig"),
						},
					},
					"customRequestHeaders": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CustomRequestHeadersConfig"),
						},
					},
					"customResponseHeaders": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CustomResponseHeadersConfig"),
						},
					},
					"healthCheck": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.HealthCheckConfig"),
						},
					},
					"logging": {
						SchemaProps: spec.SchemaProps{
							Description: "Logging specifies the configuration for access logs.",
							Ref:         ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.LogConfig"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CDNConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.ConnectionDrainingConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CustomRequestHeadersConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CustomResponseHeadersConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.HealthCheckConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.IAPConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.LogConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SecurityPolicyConfig", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SessionAffinityConfig"},
	}
}

func schema_pkg_apis_backendconfig_v1_BypassCacheOnRequestHeader(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "BypassCacheOnRequestHeader contains configuration for how requests containing specific request headers bypass the cache, even if the content was previously cached.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"headerName": {
						SchemaProps: spec.SchemaProps{
							Description: "The header field name to match on when bypassing cache. Values are case-insensitive.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_CDNConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CDNConfig contains configuration for CDN-enabled backends.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"enabled": {
						SchemaProps: spec.SchemaProps{
							Default: false,
							Type:    []string{"boolean"},
							Format:  "",
						},
					},
					"bypassCacheOnRequestHeaders": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BypassCacheOnRequestHeader"),
									},
								},
							},
						},
					},
					"cachePolicy": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CacheKeyPolicy"),
						},
					},
					"cacheMode": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"clientTtl": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					"defaultTtl": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					"maxTtl": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					"negativeCaching": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"negativeCachingPolicy": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.NegativeCachingPolicy"),
									},
								},
							},
						},
					},
					"requestCoalescing": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"serveWhileStale": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					"signedUrlCacheMaxAgeSec": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
					"signedUrlKeys": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SignedUrlKey"),
									},
								},
							},
						},
					},
				},
				Required: []string{"enabled"},
			},
		},
		Dependencies: []string{
			"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.BypassCacheOnRequestHeader", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.CacheKeyPolicy", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.NegativeCachingPolicy", "k8s.io/ingress-gce/pkg/apis/backendconfig/v1.SignedUrlKey"},
	}
}

func schema_pkg_apis_backendconfig_v1_CacheKeyPolicy(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CacheKeyPolicy contains configuration for how requests to a CDN-enabled backend are cached.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"includeHost": {
						SchemaProps: spec.SchemaProps{
							Description: "If true, requests to different hosts will be cached separately.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"includeProtocol": {
						SchemaProps: spec.SchemaProps{
							Description: "If true, http and https requests will be cached separately.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"includeQueryString": {
						SchemaProps: spec.SchemaProps{
							Description: "If true, query string parameters are included in the cache key according to QueryStringBlacklist and QueryStringWhitelist. If neither is set, the entire query string is included and if false the entire query string is excluded.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"queryStringBlacklist": {
						SchemaProps: spec.SchemaProps{
							Description: "Names of query strint parameters to exclude from cache keys. All other parameters are included. Either specify QueryStringBlacklist or QueryStringWhitelist, but not both.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
					"queryStringWhitelist": {
						SchemaProps: spec.SchemaProps{
							Description: "Names of query string parameters to include in cache keys. All other parameters are excluded. Either specify QueryStringBlacklist or QueryStringWhitelist, but not both.",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_ConnectionDrainingConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ConnectionDrainingConfig contains configuration for connection draining. For now the draining timeout. May manage more settings in the future.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"drainingTimeoutSec": {
						SchemaProps: spec.SchemaProps{
							Description: "Draining timeout in seconds.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_CustomRequestHeadersConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CustomRequestHeadersConfig contains configuration for custom request headers",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"headers": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_CustomResponseHeadersConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CustomResponseHeadersConfig contains configuration for custom response headers",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"headers": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Default: "",
										Type:    []string{"string"},
										Format:  "",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_HealthCheckConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "HealthCheckConfig contains configuration for the health check.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"checkIntervalSec": {
						SchemaProps: spec.SchemaProps{
							Description: "CheckIntervalSec is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
					"timeoutSec": {
						SchemaProps: spec.SchemaProps{
							Description: "TimeoutSec is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
					"healthyThreshold": {
						SchemaProps: spec.SchemaProps{
							Description: "HealthyThreshold is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
					"unhealthyThreshold": {
						SchemaProps: spec.SchemaProps{
							Description: "UnhealthyThreshold is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
					"type": {
						SchemaProps: spec.SchemaProps{
							Description: "Type is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"port": {
						SchemaProps: spec.SchemaProps{
							Description: "Port is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks. If Port is used, the controller updates portSpecification as well",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
					"requestPath": {
						SchemaProps: spec.SchemaProps{
							Description: "RequestPath is a health check parameter. See https://cloud.google.com/compute/docs/reference/rest/v1/healthChecks.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_IAPConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "IAPConfig contains configuration for IAP-enabled backends.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"enabled": {
						SchemaProps: spec.SchemaProps{
							Default: false,
							Type:    []string{"boolean"},
							Format:  "",
						},
					},
					"oauthclientCredentials": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/ingress-gce/pkg/apis/backendconfig/v1.OAuthClientCredentials"),
						},
					},
				},
				Required: []string{"enabled", "oauthclientCredentials"},
			},
		},
		Dependencies: []string{
			"k8s.io/ingress-gce/pkg/apis/backendconfig/v1.OAuthClientCredentials"},
	}
}

func schema_pkg_apis_backendconfig_v1_LogConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "LogConfig contains configuration for logging.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"enable": {
						SchemaProps: spec.SchemaProps{
							Description: "This field denotes whether to enable logging for the load balancer traffic served by this backend service.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"sampleRate": {
						SchemaProps: spec.SchemaProps{
							Description: "This field can only be specified if logging is enabled for this backend service. The value of the field must be in [0, 1]. This configures the sampling rate of requests to the load balancer where 1.0 means all logged requests are reported and 0.0 means no logged requests are reported. The default value is 1.0.",
							Type:        []string{"number"},
							Format:      "double",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_NegativeCachingPolicy(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "NegativeCachingPolicy contains configuration for how negative caching is applied.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"code": {
						SchemaProps: spec.SchemaProps{
							Description: "The HTTP status code to define a TTL against. Only HTTP status codes 300, 301, 308, 404, 405, 410, 421, 451 and 501 are can be specified as values, and you cannot specify a status code more than once.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
					"ttl": {
						SchemaProps: spec.SchemaProps{
							Description: "The TTL (in seconds) for which to cache responses with the corresponding status code. The maximum allowed value is 1800s (30 minutes), noting that infrequently accessed objects may be evicted from the cache before the defined TTL.",
							Type:        []string{"integer"},
							Format:      "int64",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_OAuthClientCredentials(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "OAuthClientCredentials contains credentials for a single IAP-enabled backend.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"secretName": {
						SchemaProps: spec.SchemaProps{
							Description: "The name of a k8s secret which stores the OAuth client id & secret.",
							Default:     "",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"clientID": {
						SchemaProps: spec.SchemaProps{
							Description: "Direct reference to OAuth client id.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"clientSecret": {
						SchemaProps: spec.SchemaProps{
							Description: "Direct reference to OAuth client secret.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
				Required: []string{"secretName"},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_SecurityPolicyConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SecurityPolicyConfig contains configuration for CloudArmor-enabled backends. If not specified, the controller will not reconcile the security policy configuration. In other words, users can make changes in GCE without the controller overwriting them.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"name": {
						SchemaProps: spec.SchemaProps{
							Description: "Name of the security policy that should be associated. If set to empty, the existing security policy on the backend will be removed.",
							Default:     "",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
				Required: []string{"name"},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_SessionAffinityConfig(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SessionAffinityConfig contains configuration for stickyness parameters.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"affinityType": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"affinityCookieTtlSec": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int64",
						},
					},
				},
			},
		},
	}
}

func schema_pkg_apis_backendconfig_v1_SignedUrlKey(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "SignedUrlKey represents a customer-supplied Signing Key used by Cloud CDN Signed URLs",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"keyName": {
						SchemaProps: spec.SchemaProps{
							Description: "KeyName: Name of the key. The name must be 1-63 characters long, and comply with RFC1035. Specifically, the name must be 1-63 characters long and match the regular expression `[a-z]([-a-z0-9]*[a-z0-9])?` which means the first character must be a lowercase letter, and all following characters must be a dash, lowercase letter, or digit, except the last character, which cannot be a dash.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"keyValue": {
						SchemaProps: spec.SchemaProps{
							Description: "KeyValue: 128-bit key value used for signing the URL. The key value must be a valid RFC 4648 Section 5 base64url encoded string.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"secretName": {
						SchemaProps: spec.SchemaProps{
							Description: "The name of a k8s secret which stores the 128-bit key value used for signing the URL. The key value must be a valid RFC 4648 Section 5 base64url encoded string",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
			},
		},
	}
}
