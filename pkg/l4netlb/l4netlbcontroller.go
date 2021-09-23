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

package l4netlb

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/GoogleCloudPlatform/k8s-cloud-provider/pkg/cloud/meta"
	v1 "k8s.io/api/core/v1"
	listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/cloud-provider/service/helpers"
	"k8s.io/ingress-gce/pkg/annotations"
	"k8s.io/ingress-gce/pkg/backends"
	"k8s.io/ingress-gce/pkg/context"
	"k8s.io/ingress-gce/pkg/controller/translator"
	"k8s.io/ingress-gce/pkg/instances"
	"k8s.io/ingress-gce/pkg/loadbalancers"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/ingress-gce/pkg/utils/namer"
	"k8s.io/ingress-gce/pkg/utils/patch"
	"k8s.io/klog"
)

type L4NetLbController struct{
	ctx *context.ControllerContext
	svcQueue      utils.TaskQueue
	numWorkers    int
	serviceLister cache.Indexer
	nodeLister    listers.NodeLister
	stopCh        chan struct{}

	translator *translator.Translator
	backendPool *backends.Backends
	namer       namer.L4ResourcesNamer
	// enqueueTracker tracks the latest time an update was enqueued
	enqueueTracker utils.TimeTracker
	// syncTracker tracks the latest time an enqueued service was synced
	syncTracker         utils.TimeTracker
	sharedResourcesLock sync.Mutex

	instancePool instances.NodePool
	igLinker  	 backends.Linker
}

// NewL4NetLbController creates a controller for l4 external loadbalancer.
func NewL4NetLbController(
		ctx *context.ControllerContext,
		stopCh chan struct{}) *L4NetLbController {
	if ctx.NumL4Workers <= 0 {
		klog.Infof("L4 Worker count has not been set, setting to 1")
		ctx.NumL4Workers = 1
	}

	backendPool := backends.NewPool(ctx.Cloud, ctx.L4Namer)
	instancePool := instances.NewNodePool(ctx.Cloud, ctx.ClusterNamer, ctx, utils.GetBasePath(ctx.Cloud))
	l4netLb := &L4NetLbController{
		ctx:           ctx,
		numWorkers:    ctx.NumL4Workers,
		serviceLister: ctx.ServiceInformer.GetIndexer(),
		nodeLister:    listers.NewNodeLister(ctx.NodeInformer.GetIndexer()),
		stopCh:        stopCh,
		translator:		 translator.NewTranslator(ctx),
		backendPool: 	 backendPool,
		namer:				 ctx.L4Namer,
		instancePool: instancePool,
		igLinker:      backends.NewInstanceInternalGroupLinker(instancePool, backendPool),
	}
	l4netLb.svcQueue = utils.NewPeriodicTaskQueueWithMultipleWorkers("l4netLb", "services", l4netLb.numWorkers, l4netLb.sync)

	ctx.ServiceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			addSvc := obj.(*v1.Service)
			svcKey := utils.ServiceKeyFunc(addSvc.Namespace, addSvc.Name)
			needsNetLB, svcType := annotations.WantsNewL4NetLb(addSvc)
			// Check for deletion since updates or deletes show up as Add when controller restarts.
			if needsNetLB || needsDeletion(addSvc) {
				klog.V(3).Infof("NetLB Service %s added, enqueuing", svcKey)
				l4netLb.ctx.Recorder(addSvc.Namespace).Eventf(addSvc, v1.EventTypeNormal, "ADD", svcKey)
				l4netLb.svcQueue.Enqueue(addSvc)
				l4netLb.enqueueTracker.Track()
			} else {
				klog.V(4).Infof("L4NetLb Ignoring add for non-lb service %s based on %v", svcKey, svcType)
			}
		},
		// Deletes will be handled in the Update when the deletion timestamp is set.
		UpdateFunc: func(old, cur interface{}) {
			curSvc := cur.(*v1.Service)
			svcKey := utils.ServiceKeyFunc(curSvc.Namespace, curSvc.Name)
			// oldSvc := old.(*v1.Service)
			needsUpdate := false // l4netLb.needsUpdate(oldSvc, curSvc)
			needsDeletion := needsDeletion(curSvc)
			if needsUpdate || needsDeletion {
				klog.V(3).Infof(" L4NetLb Service %v changed, needsUpdate %v, needsDeletion %v, enqueuing", svcKey, needsUpdate, needsDeletion)
				l4netLb.svcQueue.Enqueue(curSvc)
				l4netLb.enqueueTracker.Track()
				return
			}
			// Enqueue ILB services periodically for reasserting that resources exist.
			needsILB, _ := annotations.WantsNewL4NetLb(curSvc)
			if needsILB && reflect.DeepEqual(old, cur) {
				// this will happen when informers run a resync on all the existing services even when the object is
				// not modified.
				klog.V(3).Infof(" L4NetLb Periodic enqueueing of %v", svcKey)
				l4netLb.svcQueue.Enqueue(curSvc)
				l4netLb.enqueueTracker.Track()
			}
		},
	})
	klog.Infof("l4NetLbIgController started")
	ctx.AddHealthCheck("service-controller health", l4netLb.checkHealth)
	return l4netLb
}

func (l4netlbc *L4NetLbController) checkHealth() error {
	//TODO
	return nil
}

//Init inits instance Pool
func (l4netlbc *L4NetLbController) Init() {
	l4netlbc.instancePool.Init(l4netlbc.translator)
}

// Run starts the loadbalancer controller.
func (l4netlbc *L4NetLbController) Run() {
	klog.Infof("Starting l4NetLbController")
	l4netlbc.svcQueue.Run()

	<-l4netlbc.stopCh
	klog.Infof("Shutting down l4NetLbController")
}

// This should only be called when the process is being terminated.
func (l4netlbc *L4NetLbController) shutdown() {
	klog.Infof("Shutting down l4NetLbController")
	l4netlbc.svcQueue.Shutdown()
}

func (l4netlbc *L4NetLbController) sync(key string) error {
	klog.V(3).Infof("L4NetLbController Sync start")
	l4netlbc.syncTracker.Track()
	svc, exists, err := l4netlbc.ctx.Services().GetByKey(key)
	if err != nil {
		return fmt.Errorf("Failed to lookup NetLb service for key %s : %w", key, err)
	}
	if !exists || svc == nil {

		klog.V(3).Infof("Ignoring delete of service %s not managed by L4NetLbController", key)
		return nil
	}
	var result *loadbalancers.SyncResultNetLb
	if wantsNetLb, _ := annotations.WantsNewL4NetLb(svc); wantsNetLb {
		result = l4netlbc.processServiceCreateOrUpdate(key, svc)
		if result == nil {
			// result will be nil if the service was ignored(due to presence of service controller finalizer).
			return nil
		}
		return result.Error
	}
	klog.V(3).Infof("Ignoring sync of service %s, neither delete nor ensure needed.", key)
	return nil
}

func (l4netlbc *L4NetLbController) processServiceDeletion(key string, svc *v1.Service) *loadbalancers.SyncResultNetLb {
	klog.V(0).Infof(" L4netLbController processServiceDeletion")
	l4netlb := loadbalancers.NewL4NetLbHandler(svc, l4netlbc.ctx.Cloud, meta.Regional, l4netlbc.namer, l4netlbc.ctx.Recorder(svc.Namespace), l4netlbc.igLinker, &l4netlbc.sharedResourcesLock)
	l4netlbc.ctx.Recorder(svc.Namespace).Eventf(svc, v1.EventTypeNormal, "DeletingLoadBalancer", "Deleting load balancer for %s", key)
	igName := l4netlbc.ctx.ClusterNamer.InstanceGroup()
	klog.Infof("Deleting instance group %v", igName)
	var result loadbalancers.SyncResultNetLb
	if err := l4netlbc.instancePool.DeleteInstanceGroup(igName); err != err {
		l4netlbc.ctx.Recorder(svc.Namespace).Eventf(svc, v1.EventTypeWarning, "DeleteLoadBalancer",
			"Error reseting load balancer status to empty: %v", err)
		result.Error = fmt.Errorf("failed to reset ILB status, err: %w", err)
		return &result
	}
	result = *l4netlb.EnsureNetLbLoadBalancerDeleted(svc)
	if result.Error != nil {
		l4netlbc.ctx.Recorder(svc.Namespace).Eventf(svc, v1.EventTypeWarning, "DeleteLoadBalancerFailed", "Error deleting load balancer: %v", result.Error)
		return &result
	}
	return nil
}

// processServiceCreateOrUpdate ensures load balancer resources for the given service, as needed.
// Returns an error if processing the service update failed.
func (l4netlbc *L4NetLbController) processServiceCreateOrUpdate(key string, service *v1.Service) *loadbalancers.SyncResultNetLb {
	klog.V(0).Infof(" L4 netLbIg controller processServiceCreateOrUpdate")
	l4netlb := loadbalancers.NewL4NetLbHandler(service, l4netlbc.ctx.Cloud, meta.Regional, l4netlbc.namer, l4netlbc.ctx.Recorder(service.Namespace), l4netlbc.igLinker, &l4netlbc.sharedResourcesLock)
	if !l4netlbc.shouldProcessService(service, l4netlb) {
		return nil
	}

	// #TODO Add ensure finalizer for NetLB
	nodeNames, err := utils.GetReadyNodeNames(l4netlbc.nodeLister)
	if err != nil {
		return &loadbalancers.SyncResultNetLb{Error: err}
	}
	klog.V(3).Infof(" L4netLbController EnsureInsranceGroup %v", nodeNames)
	// Add or create instance groups
	if err:= l4netlbc.EnsureInsranceGroup(service, nodeNames); err != nil {
		l4netlbc.ctx.Recorder(service.Namespace).Eventf(service, v1.EventTypeWarning, "SyncInstanceagroupsFailed",
			"Error syncing load balancer: %v", err)
		return &loadbalancers.SyncResultNetLb{Error: err}
	}

	// Use the same function for both create and updates. If controller crashes and restarts,
	// all existing services will show up as Service Adds.
	syncResult := l4netlb.EnsureNetLoadBalancer(nodeNames, service)
	if syncResult.Error != nil {
		l4netlbc.ctx.Recorder(service.Namespace).Eventf(service, v1.EventTypeWarning, "SyncLoadBalancerFailed",
			"Error syncing l4 net load balancer: %v", syncResult.Error)
		return syncResult
	}
	// link ig to backend service
	zones, err := l4netlbc.translator.ListZones()
	if err != nil {
		return nil
	}
	var groupKeys []backends.GroupKey
	for _, zone := range zones {
		groupKeys = append(groupKeys, backends.GroupKey{Zone: zone})
	}
	if err = l4netlbc.igLinker.Link(l4netlb.ServicePort, groupKeys); err != nil {
		syncResult.Error = err
		return syncResult
	}

	err = l4netlbc.updateServiceStatus(service, syncResult.Status)
	if err != nil {
		l4netlbc.ctx.Recorder(service.Namespace).Eventf(service, v1.EventTypeWarning, "SyncLoadBalancerFailed",
			"Error updating l4 net load balancer status: %v", err)
		syncResult.Error = err
		return syncResult
	}
	l4netlbc.ctx.Recorder(service.Namespace).Eventf(service, v1.EventTypeNormal, "SyncLoadBalancerSuccessful",
		"Successfully ensured l4 net load balancer resources")
	return nil
}

// shouldProcessService returns if the given LoadBalancer service should be processed by this controller.
func (l4netlbc *L4NetLbController) shouldProcessService(service *v1.Service, l4 *loadbalancers.L4NetLb) bool {
	//TODO
	return true
}

// EnsureInsranceGroup adds or deletes instances to pool
func (l4netlbc *L4NetLbController) EnsureInsranceGroup(service *v1.Service, nodeNames []string) error {
	_, _, nodePorts, _ := utils.GetPortsAndProtocol(service.Spec.Ports)

	klog.V(3).Infof("Syncing Instance Group L4Netlb %s with nodeports %v", service.Name, nodePorts)
	_, err := l4netlbc.instancePool.EnsureInstanceGroupsAndPorts(l4netlbc.ctx.ClusterNamer.InstanceGroup(), nodePorts)
	if err != nil {
		klog.V(2).Infof("EnsureInstanceGroupsAndPortse Group L4Netlb %s with nodeports Error %v", l4netlbc.ctx.ClusterNamer.InstanceGroup(), err)
		return err
	}

	// Add/remove instances to the instance groups.
	if err = l4netlbc.instancePool.Sync(nodeNames); err != nil {
		klog.V(2).Infof("Sync Group L4Netlb %v with nodeports Error %v", nodeNames, err)
		return err
	}
	return nil
}

func needsDeletion(svc *v1.Service) bool {
	//TODO
	needsILB, _ := annotations.WantsNewL4NetLb(svc)
	return !needsILB
}

// needsUpdate checks if load balancer needs to be updated due to change in attributes.
func (l4netlbc *L4NetLbController) needsUpdate(oldService *v1.Service, newService *v1.Service) bool {
	//TODO
	return false
}

// updateServiceStatus checks if status changed and then updates it
func (l4netlbc *L4NetLbController) updateServiceStatus(svc *v1.Service, newStatus *v1.LoadBalancerStatus) error {
	if helpers.LoadBalancerStatusEqual(&svc.Status.LoadBalancer, newStatus) {
		return nil
	}
	return patch.PatchServiceLoadBalancerStatus(l4netlbc.ctx.KubeClient.CoreV1(), svc, *newStatus)
}