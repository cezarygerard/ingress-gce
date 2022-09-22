/*
Copyright 2015 The Kubernetes Authors.

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

package firewalls

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/api/compute/v1"
	api_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cloud-provider-gcp/crd/apis/gcpfirewall/v1beta1"
	firewallclient "k8s.io/cloud-provider-gcp/crd/client/gcpfirewall/clientset/versioned"
	"k8s.io/ingress-gce/pkg/utils"
	namer_util "k8s.io/ingress-gce/pkg/utils/namer"
	"k8s.io/klog/v2"
	"k8s.io/legacy-cloud-providers/gce"
	netset "k8s.io/utils/net"
)

const (
	// DefaultFirewallName is the name to use for firewall rules created
	// by an L7 controller when --firewall-rule is not used.
	DefaultFirewallName   = ""
	FirewallReasonAdded   = "Added Firewall CR(%v)"
	FirewallReasonUpdated = "Updated Firewall CR(%v), Status:%v, Reason:%v"
	FirewallReasonDeleted = "Deleted Firewall CR(%v)"
)

// FirewallRules manages firewall rules.
type FirewallRules struct {
	cloud     Firewall
	namer     *namer_util.Namer
	srcRanges []string
	// TODO(rramkumar): Eliminate this variable. We should just pass in
	// all the port ranges to open with each call to Sync()
	nodePortRanges       []string
	firewallClient       firewallclient.Interface
	enableFirewallCR     bool
	disableFWEnforcement bool
}

// NewFirewallPool creates a new firewall rule manager.
// cloud: the cloud object implementing Firewall.
// namer: cluster namer.
func NewFirewallPool(client firewallclient.Interface, cloud Firewall, namer *namer_util.Namer, l7SrcRanges []string, nodePortRanges []string, enableCR, disableFWEnforcement bool) SingleFirewallPool {
	_, err := netset.ParseIPNets(l7SrcRanges...)
	if err != nil {
		klog.Fatalf("Could not parse L7 src ranges %v for firewall rule: %v", l7SrcRanges, err)
	}
	return &FirewallRules{
		cloud:                cloud,
		namer:                namer,
		srcRanges:            l7SrcRanges,
		nodePortRanges:       nodePortRanges,
		firewallClient:       client,
		enableFirewallCR:     enableCR,
		disableFWEnforcement: disableFWEnforcement,
	}
}

// Sync firewall rules with the cloud.
func (fr *FirewallRules) Sync(nodeNames, additionalPorts, additionalRanges []string, allowNodePort bool) error {
	klog.V(4).Infof("Sync(%v)", nodeNames)
	name := fr.namer.FirewallRule()

	// Retrieve list of target tags from node names. This may be configured in
	// gce.conf or computed by the GCE cloudprovider package.
	targetTags, err := fr.cloud.GetNodeTags(nodeNames)
	if err != nil {
		return err
	}
	sort.Strings(targetTags)

	// De-dupe ports
	ports := sets.NewString()
	if allowNodePort {
		ports.Insert(fr.nodePortRanges...)
	}
	ports.Insert(additionalPorts...)

	// De-dupe srcRanges
	ranges := sets.NewString(fr.srcRanges...)
	ranges.Insert(additionalRanges...)

	expectedFirewall := &compute.Firewall{
		Name:         name,
		Description:  "GCE L7 firewall rule",
		SourceRanges: ranges.UnsortedList(),
		Network:      fr.cloud.NetworkURL(),
		Allowed: []*compute.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports:      ports.List(),
			},
		},
		TargetTags: targetTags,
	}
	if fr.enableFirewallCR {
		klog.V(3).Infof("Firewall CR is enabled.")
		expectedFirewallCR, err := NewFirewallCR(name, ports.List(), ranges.UnsortedList(), []string{}, fr.disableFWEnforcement)
		if err != nil {
			return err
		}
		err = ensureFirewallCR(fr.firewallClient, expectedFirewallCR)
		if err != nil && errors.Is(err, fmt.Errorf(string(v1beta1.FirewallRuleReasonXPNPermissionError))) {
			gcloudCmd := gce.FirewallToGCloudCreateCmd(expectedFirewall, fr.cloud.NetworkProjectID())
			klog.V(3).Infof("Could not create L7 firewall on XPN cluster: %v. Raising event for cmd: %q", err, gcloudCmd)
			err = newFirewallXPNError(err, gcloudCmd)
		}
		if fr.disableFWEnforcement {
			return err
		}
	}

	existingFirewall, _ := fr.cloud.GetFirewall(name)

	if existingFirewall == nil {
		klog.V(3).Infof("Creating firewall rule %q", name)
		return fr.createFirewall(expectedFirewall)
	}

	// Early return if an update is not required.
	if equal(expectedFirewall, existingFirewall) {
		klog.V(4).Info("Firewall does not need update of ports or source ranges")
		return nil
	}

	klog.V(3).Infof("Updating firewall rule %q", name)
	return fr.updateFirewall(expectedFirewall)
}

// ensureFirewallCR creates/updates the firewall CR
// On CR update, it will read the conditions to see if there are errors updated by PFW controller.
// If the Spec was updated by others, it will reconcile the Spec.
func ensureFirewallCR(client firewallclient.Interface, expectedFWCR *v1beta1.GCPFirewall) error {
	fw := client.NetworkingV1beta1().GCPFirewalls()
	currentFWCR, err := fw.Get(context.Background(), expectedFWCR.Name, metav1.GetOptions{})
	klog.V(3).Infof("ensureFirewallCR Get CR :%+v, err :%v", currentFWCR, err)
	if err != nil {
		// Create the CR if it is not found.
		if api_errors.IsNotFound(err) {
			klog.V(3).Infof("The CR is not found. ensureFirewallCR Create CR :%+v", expectedFWCR)
			_, err = fw.Create(context.Background(), expectedFWCR, metav1.CreateOptions{})
		}
		return err
	}
	if !reflect.DeepEqual(currentFWCR.Spec, expectedFWCR.Spec) {
		// Update the current firewall CR
		klog.V(3).Infof("ensureFirewallCR Update CR currentFW.Spec: %+v, expectedFW.Spec: %+v", currentFWCR.Spec, expectedFWCR.Spec)
		currentFWCR.Spec = expectedFWCR.Spec
		_, err = fw.Update(context.Background(), currentFWCR, metav1.UpdateOptions{})
		return err
	}
	for _, con := range currentFWCR.Status.Conditions {
		if con.Reason == string(v1beta1.FirewallRuleReasonXPNPermissionError) ||
			con.Reason == string(v1beta1.FirewallRuleReasonInvalid) ||
			con.Reason == string(v1beta1.FirewallRuleReasonSyncError) {
			// Use recorder to emit the cmd in Sync()
			klog.V(3).Infof("ensureFirewallCR(%v): Could not enforce Firewall CR Reason:%v", currentFWCR.Name, con.Reason)
			return fmt.Errorf(con.Reason)
		}
	}
	return nil
}

// deleteFirewallCR deletes the firewall CR
func deleteFirewallCR(client firewallclient.Interface, name string) error {
	fw := client.NetworkingV1beta1().GCPFirewalls()
	klog.V(3).Infof("Delete CR :%v", name)
	return fw.Delete(context.Background(), name, metav1.DeleteOptions{})
}

// NewFirewallCR constructs the firewall CR from name, ports and ranges
func NewFirewallCR(name string, ports, srcRanges, dstRanges []string, enforced bool) (*v1beta1.GCPFirewall, error) {
	firewallCR := &v1beta1.GCPFirewall{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1beta1.GCPFirewallSpec{
			Action:   v1beta1.ActionAllow,
			Disabled: !enforced,
		},
	}
	var protocolPorts []v1beta1.ProtocolPort

	for _, item := range ports {
		s := strings.Split(item, "-")
		var protocolPort v1beta1.ProtocolPort
		startport, err := strconv.Atoi(s[0])
		if err != nil {
			return nil, fmt.Errorf("failed to convert startport to an integer: %w", err)
		}
		sp32 := int32(startport)
		if len(s) > 1 {
			endport, err := strconv.Atoi(s[1])
			if err != nil {
				return nil, fmt.Errorf("failed to convert endport to an integer: %w", err)
			}
			ep32 := int32(endport)
			protocolPort = v1beta1.ProtocolPort{
				Protocol:  v1beta1.ProtocolTCP,
				StartPort: &sp32,
				EndPort:   &ep32,
			}
		} else {
			protocolPort = v1beta1.ProtocolPort{
				Protocol:  v1beta1.ProtocolTCP,
				StartPort: &sp32,
			}
		}
		protocolPorts = append(protocolPorts, protocolPort)
	}
	firewallCR.Spec.Ports = protocolPorts

	var src_cidrs, dst_cidrs []v1beta1.CIDR

	for _, item := range srcRanges {
		src_cidrs = append(src_cidrs, v1beta1.CIDR(item))
	}

	for _, item := range dstRanges {
		dst_cidrs = append(dst_cidrs, v1beta1.CIDR(item))
	}
	firewallCR.Spec.Ingress = &v1beta1.GCPFirewallIngress{
		Source: &v1beta1.IngressSource{
			IPBlocks: src_cidrs,
		},
		Destination: &v1beta1.IngressDestination{
			IPBlocks: dst_cidrs,
		},
	}

	return firewallCR, nil
}

// GC deletes the firewall rule.
func (fr *FirewallRules) GC() error {
	name := fr.namer.FirewallRule()
	klog.V(3).Infof("Deleting firewall %q", name)
	return fr.deleteFirewall(name)
}

// GCFirewallCR deletes the firewall CR
// For the upgraded clusters with EnableFirewallCR = true, the firewall CR and the firewall co-exist.
// We need to delete both of them every time.
func (fr *FirewallRules) GCFirewallCR() error {
	name := fr.namer.FirewallRule()
	klog.V(3).Infof("Deleting firewall CR %q", name)
	return deleteFirewallCR(fr.firewallClient, name)
}

// GetFirewall just returns the firewall object corresponding to the given name.
// TODO: Currently only used in testing. Modify so we don't leak compute
// objects out of this interface by returning just the (src, ports, error).
func (fr *FirewallRules) GetFirewall(name string) (*compute.Firewall, error) {
	return fr.cloud.GetFirewall(name)
}

func (fr *FirewallRules) createFirewall(f *compute.Firewall) error {
	err := fr.cloud.CreateFirewall(f)
	if utils.IsForbiddenError(err) && fr.cloud.OnXPN() {
		gcloudCmd := gce.FirewallToGCloudCreateCmd(f, fr.cloud.NetworkProjectID())
		klog.V(3).Infof("Could not create L7 firewall on XPN cluster: %v. Raising event for cmd: %q", err, gcloudCmd)
		return newFirewallXPNError(err, gcloudCmd)
	}
	return err
}

func (fr *FirewallRules) updateFirewall(f *compute.Firewall) error {
	err := fr.cloud.UpdateFirewall(f)
	if utils.IsForbiddenError(err) && fr.cloud.OnXPN() {
		gcloudCmd := gce.FirewallToGCloudUpdateCmd(f, fr.cloud.NetworkProjectID())
		klog.V(3).Infof("Could not update L7 firewall on XPN cluster: %v. Raising event for cmd: %q", err, gcloudCmd)
		return newFirewallXPNError(err, gcloudCmd)
	}
	return err
}

func (fr *FirewallRules) deleteFirewall(name string) error {
	err := fr.cloud.DeleteFirewall(name)
	if utils.IsNotFoundError(err) {
		klog.Infof("Firewall with name %v didn't exist when attempting delete.", name)
		return nil
	} else if utils.IsForbiddenError(err) && fr.cloud.OnXPN() {
		gcloudCmd := gce.FirewallToGCloudDeleteCmd(name, fr.cloud.NetworkProjectID())
		klog.V(3).Infof("Could not attempt delete of L7 firewall on XPN cluster: %v. %q needs to be ran.", err, gcloudCmd)
		return newFirewallXPNError(err, gcloudCmd)
	}
	return err
}

func newFirewallXPNError(internal error, cmd string) *FirewallXPNError {
	return &FirewallXPNError{
		Internal: internal,
		Message:  fmt.Sprintf("Firewall change required by security admin: `%v`", cmd),
	}
}

type FirewallXPNError struct {
	Internal error
	Message  string
}

func (f *FirewallXPNError) Error() string {
	return f.Message
}

func equal(expected *compute.Firewall, existing *compute.Firewall) bool {
	if !sets.NewString(expected.TargetTags...).Equal(sets.NewString(existing.TargetTags...)) {
		klog.V(5).Infof("Expected target tags %v, actually %v", expected.TargetTags, existing.TargetTags)
		return false
	}

	expectedAllowed := allowedToStrings(expected.Allowed)
	existingAllowed := allowedToStrings(existing.Allowed)
	if !sets.NewString(expectedAllowed...).Equal(sets.NewString(existingAllowed...)) {
		klog.V(5).Infof("Expected allowed rules %v, actually %v", expectedAllowed, existingAllowed)
		return false
	}

	if !sets.NewString(expected.SourceRanges...).Equal(sets.NewString(existing.SourceRanges...)) {
		klog.V(5).Infof("Expected source ranges %v, actually %v", expected.SourceRanges, existing.SourceRanges)
		return false
	}

	// Ignore other firewall properties as the controller does not set them.
	return true
}

func allowedToStrings(allowed []*compute.FirewallAllowed) []string {
	var allowedStrs []string
	for _, v := range allowed {
		sort.Strings(v.Ports)
		s := strings.ToUpper(v.IPProtocol) + ":" + strings.Join(v.Ports, ",")
		allowedStrs = append(allowedStrs, s)
	}
	return allowedStrs
}
