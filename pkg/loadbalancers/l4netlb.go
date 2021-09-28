/*
Copyright 2020 The Kubernetes Authors.

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
	"strconv"
	"strings"
	"sync"
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
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/ingress-gce/pkg/utils/namer"
	"k8s.io/klog"
	"k8s.io/legacy-cloud-providers/gce"
)

// L4NetLb handles the resource creation/deletion/update for a given L4NetLb service.
type L4NetLb struct {
	cloud       *gce.Cloud
	backendPool *backends.Backends
	scope       meta.KeyType
	// TODO(52752) change namer for proper NetLb Namer
	namer namer.L4ResourcesNamer
	// recorder is used to generate k8s Events.
	recorder            record.EventRecorder
	Service             *corev1.Service
	ServicePort         utils.ServicePort
	NamespacedName      types.NamespacedName
	sharedResourcesLock *sync.Mutex
}

// SyncResultNetLb contains information about the outcome of an L4NetLb sync. It stores the list of resource name annotations,
// sync error, the GCE resource that hit the error along with the error type and more fields.
type SyncResultNetLb struct {
	Annotations        map[string]string
	Error              error
	GCEResourceInError string
	Status             *corev1.LoadBalancerStatus
	SyncType           string
	StartTime          time.Time
}

// NewL4NetLbHandler creates a new Handler for the given L4NetLb service.
func NewL4NetLbHandler(service *corev1.Service, cloud *gce.Cloud, scope meta.KeyType, namer namer.L4ResourcesNamer, recorder record.EventRecorder, lock *sync.Mutex) *L4NetLb {
	l4netlb := &L4NetLb{cloud: cloud,
		scope:               scope,
		namer:               namer,
		recorder:            recorder,
		Service:             service,
		sharedResourcesLock: lock,
		NamespacedName:      types.NamespacedName{Name: service.Name, Namespace: service.Namespace},
		backendPool:         backends.NewPool(cloud, namer),
	}
	portId := utils.ServicePortID{Service: l4netlb.NamespacedName}
	l4netlb.ServicePort = utils.ServicePort{ID: portId,
		BackendNamer: l4netlb.namer,
		//TODO(kl52752) just for igLinker to work to be removed later!
		VMIPNEGEnabled: true,
	}
	return l4netlb
}

// CreateKey generates a meta.Key for a given GCE resource name.
func (l *L4NetLb) CreateKey(name string) (*meta.Key, error) {
	return composite.CreateKey(l.cloud, name, l.scope)
}

// EnsureNetLbDeleted performs a cleanup of all GCE resources for the given loadbalancer service.
func (l *L4NetLb) EnsureNetLbDeleted(svc *corev1.Service) *SyncResultNetLb {
	klog.V(2).Infof("EnsureNetLbDeleted(%s): attempting delete of load balancer resources", l.NamespacedName.String())
	sharedHC := true // TODO(52752) set shared based on service params
	result := &SyncResultNetLb{SyncType: SyncTypeDelete, StartTime: time.Now()}
	name, _ := l.namer.VMIPNEG(svc.Namespace, svc.Name)
	frName := l.GetFRName()
	key, err := l.CreateKey(frName)
	if err != nil {
		klog.Errorf("Failed to create key for NetLb resources with name %s for service %s, err %v", frName, l.NamespacedName.String(), err)
		result.Error = err
		return result
	}
	// If any resource deletion fails, log the error and continue cleanup.
	if err = utils.IgnoreHTTPNotFound(composite.DeleteForwardingRule(l.cloud, key, meta.VersionGA)); err != nil {
		klog.Errorf("Failed to delete forwarding rule for NetLb service %s, err %v", l.NamespacedName.String(), err)
		result.Error = err
		result.GCEResourceInError = annotations.ForwardingRuleResource
	}
	if err = ensureAddressDeleted(l.cloud, name, l.cloud.Region()); err != nil {
		klog.Errorf("Failed to delete address for NetLb service %s, err %v", l.NamespacedName.String(), err)
		result.Error = err
		result.GCEResourceInError = annotations.AddressResource
	}
	hcName, hcFwName := l.namer.L4HealthCheck(svc.Namespace, svc.Name, sharedHC)
	deleteFunc := func(name string) error {
		err := firewalls.EnsureL4InternalFirewallRuleDeleted(l.cloud, name)
		if err != nil {
			if fwErr, ok := err.(*firewalls.FirewallXPNError); ok {
				l.recorder.Eventf(l.Service, corev1.EventTypeNormal, "XPN", fwErr.Message)
				return nil
			}
			return err
		}
		return nil
	}
	// delete firewall rule allowing load balancer source ranges
	err = deleteFunc(name)
	if err != nil {
		klog.Errorf("Failed to delete firewall rule %s for NetLb service %s, err %v", name, l.NamespacedName.String(), err)
		result.GCEResourceInError = annotations.FirewallRuleResource
		result.Error = err
	}

	// delete firewall rule allowing healthcheck source ranges
	err = deleteFunc(hcFwName)
	if err != nil {
		klog.Errorf("Failed to delete firewall rule %s for NetLB service %s, err %v", hcFwName, l.NamespacedName.String(), err)
		result.GCEResourceInError = annotations.FirewallForHealthcheckResource
		result.Error = err
	}
	// delete backend service
	err = utils.IgnoreHTTPNotFound(l.backendPool.Delete(name, meta.VersionGA, meta.Regional))
	if err != nil {
		klog.Errorf("Failed to delete backends for netlb loadbalancer service %s, err  %v", l.NamespacedName.String(), err)
		result.GCEResourceInError = annotations.BackendServiceResource
		result.Error = err
	}

	// Delete healthcheck
	if sharedHC {
		l.sharedResourcesLock.Lock()
		defer l.sharedResourcesLock.Unlock()
	}
	err = utils.IgnoreHTTPNotFound(healthchecks.DeleteHealthCheck(l.cloud, hcName, meta.Regional))
	if err != nil {
		if !utils.IsInUsedByError(err) {
			klog.Errorf("Failed to delete healthcheck for NetLb service %s, err %v", l.NamespacedName.String(), err)
			result.GCEResourceInError = annotations.HealthcheckResource
			result.Error = err
			return result
		}
		// Ignore deletion error due to health check in use by another resource.
		// This will be hit if this is a shared healthcheck.
		klog.V(2).Infof("Failed to delete healthcheck %s: health check in use.", hcName)
	}
	return result
}

// GetFRName returns the name of the forwarding rule for the given NetLB service.
// This appends the protocol to the forwarding rule name, which will help supporting multiple protocols in the same service.
func (l *L4NetLb) GetFRName() string {
	_, _, _, protocol := utils.GetPortsAndProtocol(l.Service.Spec.Ports)
	return l.getFRNameWithProtocol(string(protocol))
}

func (l *L4NetLb) getFRNameWithProtocol(protocol string) string {
	return l.namer.L4ForwardingRule(l.Service.Namespace, l.Service.Name, strings.ToLower(protocol))
}

// EnsureNetLoadBalancer ensures that all GCE resources for the given loadbalancer service have
// been created. It returns a LoadBalancerStatus with the updated ForwardingRule IP address.
func (l *L4NetLb) EnsureNetLoadBalancer(nodeNames []string, svc *corev1.Service) *SyncResultNetLb {
	result := &SyncResultNetLb{
		Annotations: make(map[string]string),
		StartTime:   time.Now(),
		SyncType:    SyncTypeCreate}

	// If service already has an IP assigned, treat it as an update instead of a new Loadbalancer.
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		result.SyncType = SyncTypeUpdate
	}

	l.Service = svc
	name, _ := l.namer.VMIPNEG(l.Service.Namespace, l.Service.Name)

	// create healthcheck
	// TODO(52752) set shared based on service params
	sharedHC := true

	hcName, hcFwName := l.namer.L4HealthCheck(svc.Namespace, svc.Name, sharedHC)
	hcPath, hcPort := gce.GetNodesHealthCheckPath(), gce.GetNodesHealthCheckPort()
	if !sharedHC {
		hcPath, hcPort = helpers.GetServiceHealthCheckPathPort(l.Service)
	} else {
		// Take the lock when creating the shared healthcheck
		l.sharedResourcesLock.Lock()
	}

	_, hcLink, err := healthchecks.EnsureL4HealthCheck(l.cloud, hcName, l.NamespacedName, sharedHC, hcPath, hcPort, meta.Regional)
	if sharedHC {
		// unlock here so rest of the resource creation API can be called without unnecessarily holding the lock.
		l.sharedResourcesLock.Unlock()
	}
	if err != nil {
		result.GCEResourceInError = annotations.HealthcheckResource
		result.Error = err
		return result
	}

	_, portRanges, _, protocol := utils.GetPortsAndProtocol(l.Service.Spec.Ports)

	// ensure firewalls
	sourceRanges, err := helpers.GetLoadBalancerSourceRanges(l.Service)
	if err != nil {
		result.Error = err
		return result
	}
	hcSourceRanges := gce.L4LoadBalancerSrcRanges()
	ensureFunc := func(name, IP string, sourceRanges, portRanges []string, proto string, shared bool) error {
		if shared {
			l.sharedResourcesLock.Lock()
			defer l.sharedResourcesLock.Unlock()
		}
		nsName := utils.ServiceKeyFunc(l.Service.Namespace, l.Service.Name)
		err := firewalls.EnsureL4InternalFirewallRule(l.cloud, name, IP, nsName, sourceRanges, portRanges, nodeNames, proto, shared)
		if err != nil {
			if fwErr, ok := err.(*firewalls.FirewallXPNError); ok {
				l.recorder.Eventf(l.Service, corev1.EventTypeNormal, "XPN", fwErr.Message)
				return nil
			}
			return err
		}
		return nil
	}

	//// Add firewall rule for NetLB traffic to nodes
	err = ensureFunc(name, hcLink, sourceRanges.StringSlice(), portRanges, string(protocol), false)
	if err != nil {
		result.GCEResourceInError = annotations.FirewallRuleResource
		result.Error = err
		return result
	}

	// Add firewall rule for healthchecks to nodes
	err = ensureFunc(hcFwName, "", hcSourceRanges, []string{strconv.Itoa(int(hcPort))}, string(corev1.ProtocolTCP), false)
	if err != nil {
		result.GCEResourceInError = annotations.FirewallForHealthcheckResource
		result.Error = err
		return result
	}

	//Check if protocol has changed for this service. In this case, forwarding rule should be deleted before
	//the backend service can be updated.
	existingBS, err := l.backendPool.Get(name, meta.VersionGA, l.scope)
	err = utils.IgnoreHTTPNotFound(err)
	if err != nil {
		klog.Errorf("Failed to lookup existing backend service, ignoring err: %v", err)
	}
	existingFR := l.GetForwardingRule(l.GetFRName(), meta.VersionGA)
	if existingBS != nil && existingBS.Protocol != string(protocol) {
		klog.Infof("Protocol changed from %q to %q for service %s", existingBS.Protocol, string(protocol), l.NamespacedName)
		// Delete forwarding rule if it exists
		existingFR = l.GetForwardingRule(l.getFRNameWithProtocol(existingBS.Protocol), meta.VersionGA)
		l.deleteForwardingRule(l.getFRNameWithProtocol(existingBS.Protocol), meta.VersionGA)
	}

	// ensure backend service
	bs, err := l.backendPool.EnsureL4BackendService(name, hcLink, string(protocol), string(l.Service.Spec.SessionAffinity),
		string(cloud.SchemeExternal), l.NamespacedName, meta.VersionGA)
	if err != nil {
		result.GCEResourceInError = annotations.BackendServiceResource
		result.Error = err
		return result
	}
	// create fr rule
	frName := l.GetFRName()
	fr, err := l.ensureForwardingRule(frName, bs.SelfLink, existingFR, l.cloud)
	if err != nil {
		klog.Errorf("EnsureNetLoadBalancer: Failed to create forwarding rule - %v", err)
		result.GCEResourceInError = annotations.ForwardingRuleResource
		result.Error = err
		return result
	}
	if fr.IPProtocol == string(corev1.ProtocolTCP) {
	} else {
	}

	result.Status = &corev1.LoadBalancerStatus{Ingress: []corev1.LoadBalancerIngress{{IP: fr.IPAddress}}}
	return result
}
