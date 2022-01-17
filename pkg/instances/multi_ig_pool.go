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
	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog"
	"k8s.io/ingress-gce/pkg/utils"
	"k8s.io/ingress-gce/pkg/utils/namer"
)

type MultiIGNodePool struct {
	*Instances
}

func NewMultiIGNodePool(cloud InstanceGroups, namer namer.BackendNamer, recorders recorderSource, basePath string, zl ZoneLister) NodePool {
	mignp := &MultiIGNodePool{
		Instances: &Instances{
			cloud:              cloud,
			namer:              namer,
			recorder:           recorders.Recorder(""), // No namespace
			instanceLinkFormat: basePath + "zones/%s/instances/%s",
			ZoneLister:         zl,
		},
	}
	return mignp
}

func (igc *MultiIGNodePool) DeleteInstanceGroup(name string) error {
	klog.Infof("DeleteInstanceGroup is a no-op. Instance groups will be deleted when the cluster is deleted.")
	return nil
}

func (igc *MultiIGNodePool) Sync(nodeNames []string) error {
	klog.Infof("Sync is a no-op. Instance groups will be synced in the separate cont roller.")
	return nil
}

func (igc *MultiIGNodePool) Get(name, zone string) (*compute.InstanceGroup, error) {
	// TODO (cezarygerard) reimplement. return all IGs from zone, do not use the Instances.Get
	// TODO (cezarygerard) change signature of the Get func to return a slice
	return igc.Instances.Get(name, zone)
}

func (igc *MultiIGNodePool) List() ([]string, error) {
	// TODO (cezarygerard) reimplement. return all IGs from cluster, do not use the Instances.List
	return igc.Instances.List()
}

func (igc *MultiIGNodePool) EnsureInstanceGroupsAndPorts(name string, ports []int64) (igs []*compute.InstanceGroup, err error) {
	// Instance groups need to be created only in zones that have ready nodes.
	zones, err := igc.Instances.ListZones(utils.CandidateNodesPredicate)
	if err != nil {
		return nil, err
	}

	for _, zone := range zones {

		ig, err := igc.Get(name, zone)
		if err != nil {
			return nil, err
		}
		err = igc.setPorts(ig, zone, ports)
		if err != nil {
			return nil, err
		}
		igs = append(igs, ig)
	}
	return igs, nil
}

func (igc *MultiIGNodePool) setPorts(ig *compute.InstanceGroup, zone string, ports []int64) error {
	// Build map of existing ports
	existingPorts := map[int64]bool{}
	for _, np := range ig.NamedPorts {
		existingPorts[np.Port] = true
	}

	// Determine which ports need to be added
	var newPorts []int64
	for _, p := range ports {
		if existingPorts[p] {
			continue
		}
		newPorts = append(newPorts, p)
	}

	// Build slice of NamedPorts for adding
	var newNamedPorts []*compute.NamedPort
	for _, port := range newPorts {
		newNamedPorts = append(newNamedPorts, &compute.NamedPort{Name: igc.namer.NamedPort(port), Port: port})
	}

	if len(newNamedPorts) > 0 {
		if err := igc.cloud.SetNamedPortsOfInstanceGroup(ig.Name, zone, append(ig.NamedPorts, newNamedPorts...)); err != nil {
			return err
		}
	}

	return nil
}
