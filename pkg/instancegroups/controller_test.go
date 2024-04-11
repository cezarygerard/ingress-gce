package instancegroups

import (
	"testing"
	"time"

	api_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informerv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/klog/v2"
)

func TestSync(t *testing.T) {
	config := &ControllerConfig{}
	fakeKubeClient := fake.NewSimpleClientset()
	informer := informerv1.NewNodeInformer(fakeKubeClient, 60*time.Second, utils.NewNamespaceIndexer())

	config.NodeInformer = informer
	fakeSyncer := &IGSyncerFake{}
	config.IGSyncer = fakeSyncer
	config.HasSyncedFunc = func() bool {
		return true
	}
	controller := NewController(config, klog.TODO())

	channel := make(chan struct{})
	go informer.Run(channel)
	go controller.Run()

	node1 := testNode()
	node2 := testNode()
	node2.Name = "n2"

	informer.GetIndexer().Add(node1)
	informer.GetIndexer().Add(node2)
	//fakeKubeClient.CoreV1().Nodes().Create(context.TODO(), node1, meta_v1.CreateOptions{})
	//time.Sleep(2 * time.Second)
	//fakeKubeClient.CoreV1().Nodes().Create(context.TODO(), node2, meta_v1.CreateOptions{})
	time.Sleep(1 * time.Second)
	if len(fakeSyncer.syncedNodes) != 1 {
		t.Errorf("1 synced too many times %v", len(fakeSyncer.syncedNodes))
	}
	//klog.Warningf("hp: 1 synced times %v", len(fakeSyncer.syncedNodes))
	//
	//node1.Annotations["dupa"] = "true"
	//node1.Spec.PodCIDR = "1.1.1.1/28"
	//node2.Annotations["dupa"] = "true"
	//node2.Spec.PodCIDR = "1.1.3.1/28"
	//informer.GetIndexer().Update(node1)
	//informer.GetIndexer().Update(node2)
	////fakeKubeClient.CoreV1().Nodes().Update(context.TODO(), node1, meta_v1.UpdateOptions{})
	////fakeKubeClient.CoreV1().Nodes().Update(context.TODO(), node2, meta_v1.UpdateOptions{})
	//time.Sleep(5 * time.Second)
	//////informer.GetIndexer().Update(node2)
	//if len(fakeSyncer.syncedNodes) != 1 {
	//	t.Errorf("2 synced too many times %v", len(fakeSyncer.syncedNodes))
	//}
	//klog.Warningf("hp: 2 synced times %v", len(fakeSyncer.syncedNodes))

}

type IGSyncerFake struct {
	syncedNodes [][]string
}

func (igsf *IGSyncerFake) Sync(nodeNames []string) error {
	klog.Warningf("synced %v", nodeNames)
	igsf.syncedNodes = append(igsf.syncedNodes, nodeNames)
	return nil
}

func TestNodeStatusChanged(t *testing.T) {
	testCases := []struct {
		desc   string
		mutate func(node *api_v1.Node)
		expect bool
	}{
		{
			"no change",
			func(node *api_v1.Node) {},
			false,
		},
		{
			"unSchedulable changes",
			func(node *api_v1.Node) {
				node.Spec.Unschedulable = true
			},
			true,
		},
		{
			"readiness changes",
			func(node *api_v1.Node) {
				node.Status.Conditions[0].Status = api_v1.ConditionFalse
				node.Status.Conditions[0].LastTransitionTime = meta_v1.NewTime(time.Now())
			},
			true,
		},
		{
			"new heartbeat",
			func(node *api_v1.Node) {
				node.Status.Conditions[0].LastHeartbeatTime = meta_v1.NewTime(time.Now())
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			node := testNode()
			tc.mutate(node)
			res := nodeStatusChanged(testNode(), node)
			if res != tc.expect {
				t.Fatalf("Test case %q got: %v, expected: %v", tc.desc, res, tc.expect)
			}
		})
	}
}

func testNode() *api_v1.Node {
	return &api_v1.Node{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: "node",
			Annotations: map[string]string{
				"key1": "value1",
			},
		},
		Spec: api_v1.NodeSpec{
			Unschedulable: false,
		},
		Status: api_v1.NodeStatus{
			Conditions: []api_v1.NodeCondition{
				{
					Type:               api_v1.NodeReady,
					Status:             api_v1.ConditionTrue,
					LastHeartbeatTime:  meta_v1.NewTime(time.Date(2000, 01, 1, 1, 0, 0, 0, time.UTC)),
					LastTransitionTime: meta_v1.NewTime(time.Date(2000, 01, 1, 1, 0, 0, 0, time.UTC)),
				},
			},
		},
	}
}
