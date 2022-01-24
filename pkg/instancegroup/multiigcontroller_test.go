package instancegroup

import (
	"time"

	api_v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes/fake"
	backendconfigclient "k8s.io/ingress-gce/pkg/backendconfig/client/clientset/versioned/fake"
	"k8s.io/ingress-gce/pkg/context"
	"k8s.io/ingress-gce/pkg/instances"
	"k8s.io/ingress-gce/pkg/test"
	"k8s.io/ingress-gce/pkg/utils"
	namer_util "k8s.io/ingress-gce/pkg/utils/namer"
	"k8s.io/legacy-cloud-providers/gce"
)

var (
	clusterUID = "aaaaa"
	fakeZone   = "zone-a"
)

func newMultiInstancesGroupController() *MultiIGController {
	kubeClient := fake.NewSimpleClientset()
	backendConfigClient := backendconfigclient.NewSimpleClientset()
	fakeGCE := gce.NewFakeGCECloud(gce.DefaultTestClusterValues())
	fakeZL := &instances.FakeZoneLister{Zones: []string{fakeZone}}
	namer := namer_util.NewNamer(clusterUID, "")

	stopCh := make(chan struct{})
	ctxConfig := context.ControllerContextConfig{
		Namespace:             api_v1.NamespaceAll,
		ResyncPeriod:          1 * time.Minute,
		DefaultBackendSvcPort: test.DefaultBeSvcPort,
		HealthCheckPath:       "/",
	}
	ctx := context.NewControllerContext(nil, kubeClient, backendConfigClient, nil, nil, nil, nil, fakeGCE, namer, "", ctxConfig)
	ctx.InstancePool = instances.NewMultiIGNodePool(instances.NewFakeInstanceGroups(sets.NewString(), namer), namer, &test.FakeRecorderSource{}, utils.GetBasePath(fakeGCE), fakeZL)
	igc := NewMultiInstancesGroupController(ctx, stopCh)

	return igc
}
