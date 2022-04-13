package namer

import (
	"strings"
)


const (
	reginalHc = "regional"
)

// L4Namer implements naming scheme for L4 LoadBalancer resources.
// This uses the V2 Naming scheme
// Example:
// For Service - namespace/svc, clusterUID/clusterName - uid01234, prefix - k8s2, protocol TCP
// Assume that hash("uid01234;svc;namespace") = cysix1wq
// The resource names are -
// TCP Forwarding Rule : k8s2-tcp-uid01234-namespace-svc-cysix1wq
// UDP Forwarding Rule : k8s2-udp-uid01234-namespace-svc-cysix1wq
// All other resources : k8s2-uid01234-namespace-svc-cysix1wq
// The "namespace-svc" part of the string will be trimmed as needed.
type L4NetLBNamer struct {
	// Namer is needed to implement all methods required by BackendNamer interface.
	*L4Namer

}

func NewL4NetLBNamer(kubeSystemUID string, namer *Namer) *L4NetLBNamer {
	old :=NewL4Namer(kubeSystemUID, namer)
	l4n := &L4NetLBNamer{
		L4Namer: old,
	}
	return l4n
}

// L4HealthCheck returns the name of the L4 LB Healthcheck and the associated firewall rule.
func (namer *L4NetLBNamer) L4HealthCheck(namespace, name string, shared bool) (string, string) {
	if !shared {
		l4Name, _ := namer.L4Backend(namespace, name)
		return l4Name, namer.hcFirewallName(l4Name)
	}
	return strings.Join([]string{namer.v2Prefix, namer.v2ClusterUID, sharedHcSuffix}, "-"),
		strings.Join([]string{namer.v2Prefix, namer.v2ClusterUID, sharedFirewallHcSuffix}, "-")
}
 v