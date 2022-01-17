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

package instances

import (
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/ingress-gce/pkg/context"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/klog"
	"time"
)

type InstancesGroupController struct {
	pool NodePool
	// lister is a cache of the k8s Node resources.
	lister cache.Indexer
	// queue is the TaskQueue used to manage the node worker updates.
	queue utils.TaskQueue
	// hasSynced returns true if relevant caches have done their initial
	// synchronization.
	hasSynced func() bool

	stopCh chan struct{}
}

// NewNodeController returns a new node update controller.
func NewInstancesGroupController(ctx *context.ControllerContext, stopCh chan struct{}) *InstancesGroupController {
	igc := &InstancesGroupController{
		lister:    ctx.NodeInformer.GetIndexer(),
		hasSynced: ctx.HasSynced,
		stopCh:    stopCh,
	}
	igc.queue = utils.NewPeriodicTaskQueue("", "nodes", igc.sync)

	// Cast or die
	pool, ok := ctx.InstancePool.(*MultiIGNodePool)
	if !ok {
		klog.Fatalf("provided InstancePool must be of type MultiIGNodePool")
	}
	igc.pool = pool

	ctx.NodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			igc.queue.Enqueue(obj)
		},
		DeleteFunc: func(obj interface{}) {
			igc.queue.Enqueue(obj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			igc.queue.Enqueue(newObj)
		},
	})
	return igc
}

// Run the queue to process updates for the controller. This must be run in a
// separate goroutine (method will block until queue shutdown).
func (igc *InstancesGroupController) Run() {
	start := time.Now()
	for !igc.hasSynced() {
		klog.V(2).Infof("Waiting for hasSynced (%s elapsed)", time.Now().Sub(start))
		time.Sleep(1 * time.Second)
	}
	klog.V(2).Infof("Caches synced (took %s)", time.Now().Sub(start))
	go igc.queue.Run()
	<-igc.stopCh
	igc.Shutdown()
}

// Shutdown shuts down the goroutine that processes node updates.
func (igc *InstancesGroupController) Shutdown() {
	igc.queue.Shutdown()
}

func (igc *InstancesGroupController) sync(key string) error {
	_, err := utils.GetReadyNodeNames(listers.NewNodeLister(igc.lister))
	if err != nil {
		return err
	}
	//TODO (panslava) do the actual IG sync
	return nil
}
