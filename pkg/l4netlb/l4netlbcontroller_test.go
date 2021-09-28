package l4netlb

import (
	context2 "context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"k8s.io/ingress-gce/pkg/context"
	"k8s.io/ingress-gce/pkg/test"
	"k8s.io/ingress-gce/pkg/utils/common"
	"k8s.io/ingress-gce/pkg/utils/namer"
	"k8s.io/klog"
	"k8s.io/legacy-cloud-providers/gce"
	"testing"
	"time"
)

const (
	clusterUID = "aaaaa"
	// This is one of the zones used in gce_fake.go
	testGCEZone = "us-central1-b"
)

func TestProcessServiceCreateOrUpdate(t *testing.T) {
	l4netController := newServiceController()
	svc := test.NewL4NetLBService(8080)

	l4netController.ctx.KubeClient.CoreV1().Services(svc.Namespace).Create(context2.TODO(), svc, metav1.CreateOptions{})
	l4netController.ctx.ServiceInformer.GetIndexer().Add(svc)
	key, _ := common.KeyFunc(svc)
	err := l4netController.sync(key)
	if err != nil {
		t.Errorf("Failed to sync newly added service %s, err %v", svc.Name, err)
	}
	newSvc, err := l4netController.ctx.KubeClient.CoreV1().Services(svc.Namespace).Get(context2.TODO(), svc.Name, metav1.GetOptions{})
	if err != nil {
		t.Errorf("Failed to lookup service %s, err: %v", svc.Name, err)
	}
	t.Logf("New service: %+v", newSvc) //

	// test Check instances and instance groups and nodes
	// test: check IG link
	// test: check status
}

func newServiceController() *L4NetLbController {
	vals := gce.DefaultTestClusterValues()
	fakeGCE := gce.NewFakeGCECloud(vals)
	kubeClient := fake.NewSimpleClientset()
	namer := namer.NewNamer(clusterUID, "")

	stopCh := make(chan struct{})
	ctxConfig := context.ControllerContextConfig{
		Namespace:    v1.NamespaceAll,
		ResyncPeriod: 1 * time.Minute,
		NumL4Workers: 5,
	}
	ctx := context.NewControllerContext(nil, kubeClient, nil, nil, nil, nil, nil, fakeGCE, namer, "" /*kubeSystemUID*/, ctxConfig)
	nodes, err := test.CreateAndInsertNodes(ctx.Cloud, []string{"instance-1"}, vals.ZoneName)
	if err != nil {
		klog.Fatalf("Failed to add new nodes, err  %v", err)
	}
	for _, n := range nodes {
		ctx.NodeInformer.GetIndexer().Add(n)
	}
	return NewL4NetLbController(ctx, stopCh)
}

func validateSvcStatus(svc *v1.Service, t *testing.T) {

}
