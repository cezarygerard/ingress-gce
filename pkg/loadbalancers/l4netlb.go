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

package loadbalancers

import (
	"fmt"
	"time"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud"
	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/cloud-provider/service/helpers"
	"k8s.io/ingress-gce/pkg/annotations"
	"k8s.io/ingress-gce/pkg/backends"
	"k8s.io/ingress-gce/pkg/composite"
	"k8s.io/ingress-gce/pkg/firewalls"
	"k8s.io/ingress-gce/pkg/healthchecks"
	"k8s.io/ingress-gce/pkg/metrics"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/ingress-gce/pkg/utils/namer"
	"k8s.io/klog"
	"k8s.io/legacy-cloud-providers/gce"
)

// L4NetLB handles the resource creation/deletion/update for a given L4 External LoadBalancer service.
type L4NetLB struct {
	cloud       *gce.Cloud
	backendPool *backends.Backends
	scope       meta.KeyType
	namer       namer.L4ResourcesNamer
	// recorder is used to generate k8s Events.
	recorder       record.EventRecorder
	Service        *corev1.Service
	ServicePort    utils.ServicePort
	NamespacedName types.NamespacedName
	l4HealthChecks L4HealthChecks
}

// L4NetLBSyncResult contains information about the outcome of an L4 NetLB sync. It stores the list of resource name annotations,
// sync error, the GCE resource that hit the error along with the error type, metrics and more fields.
type L4NetLBSyncResult struct {
	Annotations        map[string]string
	Error              error
	GCEResourceInError string
	Status             *corev1.LoadBalancerStatus
	MetricsState       metrics.L4NetLBServiceState
	SyncType           string
	StartTime          time.Time
}

// NewL4NetLB creates a new Handler for the given L4NetLB service.
func NewL4NetLB(service *corev1.Service, cloud *gce.Cloud, scope meta.KeyType, namer namer.L4ResourcesNamer, recorder record.EventRecorder) *L4NetLB {
	l4netlb := &L4NetLB{cloud: cloud,
		scope:          scope,
		namer:          namer,
		recorder:       recorder,
		Service:        service,
		NamespacedName: types.NamespacedName{Name: service.Name, Namespace: service.Namespace},
		backendPool:    backends.NewPool(cloud, namer),
		l4HealthChecks: healthchecks.GetL4(),
	}
	portId := utils.ServicePortID{Service: l4netlb.NamespacedName}
	l4netlb.ServicePort = utils.ServicePort{
		ID:           portId,
		BackendNamer: l4netlb.namer,
		NodePort:     int64(service.Spec.Ports[0].NodePort),
		L4RBSEnabled: true,
	}

	return l4netlb
}

// createKey generates a meta.Key for a given GCE resource name.
func (l4netlb *L4NetLB) createKey(name string) (*meta.Key, error) {
	return composite.CreateKey(l4netlb.cloud, name, l4netlb.scope)
}

// EnsureFrontend ensures that all frontend resources for the given loadbalancer service have
// been created. It is health check, firewall rules, backend service and forwarding rule.
// It returns a LoadBalancerStatus with the updated ForwardingRule IP address.
// This function does not link instances to Backend Service.
func (l4netlb *L4NetLB) EnsureFrontend(nodeNames []string, svc *corev1.Service) *L4NetLBSyncResult {
	result := &L4NetLBSyncResult{
		Annotations: make(map[string]string),
		StartTime:   time.Now(),
		SyncType:    SyncTypeCreate}

	// If service already has an IP assigned, treat it as an update instead of a new Loadbalancer.
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		result.SyncType = SyncTypeUpdate
	}

	l4netlb.Service = svc

	sharedHC := !helpers.RequestsOnlyLocalTraffic(svc)
	hcLink, hcFwName, hcName, resourceInErr, err := l4netlb.l4HealthChecks.EnsureL4HealthCheck(l4netlb.Service, l4netlb.namer, sharedHC, l4netlb.scope, utils.XLB, nodeNames)

	if err != nil {
		result.GCEResourceInError = resourceInErr
		result.Error = fmt.Errorf("Failed to ensure health check %s - %w", hcName, err)
		return result
	}
	result.Annotations[annotations.HealthcheckKey] = hcName

	name := l4netlb.ServicePort.BackendName()
	protocol, res := l4netlb.createFirewalls(name, nodeNames)
	if res.Error != nil {
		return res
	}
	result.Annotations[annotations.FirewallRuleKey] = name
	result.Annotations[annotations.FirewallRuleForHealthcheckKey] = hcFwName

	bs, err := l4netlb.backendPool.EnsureL4BackendService(name, hcLink, protocol, string(l4netlb.Service.Spec.SessionAffinity), string(cloud.SchemeExternal), l4netlb.NamespacedName, meta.VersionGA)
	if err != nil {
		result.GCEResourceInError = annotations.BackendServiceResource
		result.Error = fmt.Errorf("Failed to ensure backend service %s - %w", name, err)
		return result
	}
	result.Annotations[annotations.BackendServiceKey] = name
	fr, ipAddrType, err := l4netlb.ensureExternalForwardingRule(bs.SelfLink)
	if err != nil {
		result.GCEResourceInError = annotations.ForwardingRuleResource
		result.Error = fmt.Errorf("Failed to ensure forwarding rule - %w", err)
		return result
	}
	if fr.IPProtocol == string(corev1.ProtocolTCP) {
		result.Annotations[annotations.TCPForwardingRuleKey] = fr.Name
	} else {
		result.Annotations[annotations.UDPForwardingRuleKey] = fr.Name
	}
	result.Status = &corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: fr.IPAddress}}}
	if fr.NetworkTier == cloud.NetworkTierPremium.ToGCEValue() {
		result.MetricsState.IsPremiumTier = true
	}
	if ipAddrType == IPAddrManaged {
		result.MetricsState.IsManagedIP = true
	}
	return result
}

// EnsureLoadBalancerDeleted performs a cleanup of all GCE resources for the given loadbalancer service.
// It is health check, firewall rules and backend service
func (l4netlb *L4NetLB) EnsureLoadBalancerDeleted(svc *corev1.Service) *L4NetLBSyncResult {
	result := &L4NetLBSyncResult{SyncType: SyncTypeDelete, StartTime: time.Now()}

	frName := l4netlb.GetFRName()
	key, err := l4netlb.createKey(frName)
	if err != nil {
		klog.Errorf("Failed to create key for forwarding rule resources with name %s for service %s - %v", frName, l4netlb.NamespacedName.String(), err)
		result.Error = err
		return result
	}
	// If any resource deletion fails, log the error and continue cleanup.
	if err = utils.IgnoreHTTPNotFound(composite.DeleteForwardingRule(l4netlb.cloud, key, meta.VersionGA)); err != nil {
		klog.Errorf("Failed to delete forwarding rule %s for service %s - %v", frName, l4netlb.NamespacedName.String(), err)
		result.Error = err
		result.GCEResourceInError = annotations.ForwardingRuleResource
	}
	name := l4netlb.ServicePort.BackendName()
	if err = ensureAddressDeleted(l4netlb.cloud, name, l4netlb.cloud.Region()); err != nil {
		klog.Errorf("Failed to delete address for service %s - %v", l4netlb.NamespacedName.String(), err)
		result.Error = err
		result.GCEResourceInError = annotations.AddressResource
	}

	err = l4netlb.deleteFirewall(name)
	if err != nil {
		klog.Errorf("Failed to delete firewall rule %s for service %s - %v", name, l4netlb.NamespacedName.String(), err)
		result.GCEResourceInError = annotations.FirewallRuleResource
		result.Error = err
	}
	// delete backend service
	err = utils.IgnoreHTTPNotFound(l4netlb.backendPool.Delete(name, meta.VersionGA, meta.Regional))
	if err != nil {
		klog.Errorf("Failed to delete backends for L4 External LoadBalancer service %s - %v", l4netlb.NamespacedName.String(), err)
		result.GCEResourceInError = annotations.BackendServiceResource
		result.Error = err
	}

	// Delete healthcheck
	// We don't delete health check during service update so
	// it is possible that there might be some health check leak
	// when externalTrafficPolicy is changed from Local to Cluster and new a health check was created.
	// When service is deleted we need to check both health checks shared and non-shared
	// and delete them if needed.
	for _, isShared := range []bool{true, false} {
		resourceInError, err := l4netlb.l4HealthChecks.DeleteHealthCheck(svc, l4netlb.namer, isShared, meta.Regional, utils.XLB)
		if err != nil {
			result.GCEResourceInError = resourceInError
			result.Error = err
		}
	}
	return result
}

func (l4netlb *L4NetLB) deleteFirewall(name string) error {
	err := firewalls.EnsureL4FirewallRuleDeleted(l4netlb.cloud, name)
	if err != nil {
		if fwErr, ok := err.(*firewalls.FirewallXPNError); ok {
			l4netlb.recorder.Eventf(l4netlb.Service, corev1.EventTypeNormal, "XPN", fwErr.Message)
			return nil
		}
		return err
	}
	return nil
}

// GetFRName returns the name of the forwarding rule for the given L4 External LoadBalancer service.
// This name should align with legacy forwarding rule name because we use forwarding rule to determine
// which controller should process the service Ingress-GCE or k/k service controller.
func (l4netlb *L4NetLB) GetFRName() string {
	return utils.LegacyForwardingRuleName(l4netlb.Service)
}

func (l4netlb *L4NetLB) createFirewalls(name string, nodeNames []string) (string, *L4NetLBSyncResult) {
	_, portRanges, _, protocol := utils.GetPortsAndProtocol(l4netlb.Service.Spec.Ports)
	result := &L4NetLBSyncResult{}
	sourceRanges, err := helpers.GetLoadBalancerSourceRanges(l4netlb.Service)
	if err != nil {
		result.Error = err
		return "", result
	}

	// Add firewall rule for L4 External LoadBalancer traffic to nodes
	nodesFWRParams := firewalls.FirewallParams{
		PortRanges:   portRanges,
		SourceRanges: sourceRanges.StringSlice(),
		Protocol:     string(protocol),
		Name:         name,
		IP:           l4netlb.Service.Spec.LoadBalancerIP,
		NodeNames:    nodeNames,
	}
	result.Error = firewalls.EnsureL4LBFirewallForNodes(l4netlb.Service, &nodesFWRParams, l4netlb.cloud, l4netlb.recorder)
	if result.Error != nil {
		result.GCEResourceInError = annotations.FirewallRuleResource
		result.Error = err
		return "", result
	}
	return string(protocol), result
}
