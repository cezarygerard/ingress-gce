package instances

import (
	"fmt"
	compute "google.golang.org/api/compute/v1"
	"k8s.io/ingress-gce/pkg/test"
	"k8s.io/ingress-gce/pkg/utils/namer"
	"testing"
)

const (
	clusterID         = "clusterUID"
	igName            = "igname"
	zone              = "zone1"
	emptyBasePath     = ""
	emptyFirewallName = ""
)

func TestGetInstanceGroups(t *testing.T) {
	recorder := &test.FakeRecorderSource{}
	backendNamer := namer.NewNamer(clusterID, emptyFirewallName)
	zones := []string{zone}
	zoneLister := &FakeZoneLister{zones}
	fakeCloud := NewFakeInstanceGroups(nil, defaultNamer)
	multiIGInst := NewMultiIGInstances(fakeCloud, backendNamer, recorder, emptyBasePath, zoneLister)

	testCases := []struct {
		name           string
		instanceGroups []*compute.InstanceGroup
		expecetedIGs   int
		expectedError  bool
	}{
		{
			name:           "Single IG in cluster",
			instanceGroups: []*compute.InstanceGroup{{Name: "k8s-ig--clusterUIDdef"}},
			expecetedIGs:   1,
			expectedError:  false,
		},

		{
			name:           "Single IG, not in cluster",
			instanceGroups: []*compute.InstanceGroup{{Name: "k8s-ig--cluster2UIDdef-1"}},
			expecetedIGs:   0,
			expectedError:  true,
		},
		{
			name: "Multiple IGs in cluster",
			instanceGroups: []*compute.InstanceGroup{
				{Name: "k8s-ig--clusterUIDdef"},
				{Name: "k8s-ig--clusterUIDdef-1"},
				{Name: "k8s-ig--clusterUIDdef-2"},
				{Name: "k8s-ig--clusterUIDdef-3"},
				{Name: "k8s-ig--clusterUIDdef-5"},
			},
			expecetedIGs:  5,
			expectedError: false,
		},
		{
			name: "Multiple IGs, some in cluster",
			instanceGroups: []*compute.InstanceGroup{
				{Name: "k8s-ig--cluster1UIDdef"},
				{Name: "k8s-ig--cluster2UIDdef-1"},
				{Name: "k8s-ig--clusterUIDdef-2"},
				{Name: "k8s-ig--clusterUIDdef-3"},
				{Name: "k8s-ig--cluster1UIDdef-5"},
			},
			expecetedIGs:  2,
			expectedError: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCloud.instanceGroups = tc.instanceGroups
			result, err := multiIGInst.Get(igName, "zone1")
			if len(result) != tc.expecetedIGs {
				t.Errorf("Incorrect number of instance grouos, got: %v, want: %v", len(result), tc.expecetedIGs)
			}
			gotError := (err != nil)
			if gotError != tc.expectedError {
				t.Errorf("Unexpected error, got: %v, want no error", err)
			}
		})
	}
}

func TestEnsureInstanceGroupsAndPorts(t *testing.T) {
	recorder := &test.FakeRecorderSource{}
	backendNamer := namer.NewNamer(clusterID, emptyFirewallName)
	zones := []string{"zone1"}
	zoneLister := &FakeZoneLister{zones}
	fakeCloud := NewFakeInstanceGroups(nil, defaultNamer)
	multiIGInst := NewMultiIGInstances(fakeCloud, backendNamer, recorder, emptyBasePath, zoneLister)

	testCases := []struct {
		name           string
		instanceGroups []*compute.InstanceGroup
		ports          []int64
		expecetedIGs   int
		expectedError  bool
	}{
		{
			name:           "No instance groups",
			instanceGroups: []*compute.InstanceGroup{},
			expecetedIGs:   0,
			expectedError:  true,
		},
		{
			name:          "Nil instance groups",
			expecetedIGs:  0,
			expectedError: true,
		},
		{
			name:           "Nil ports",
			instanceGroups: []*compute.InstanceGroup{{Name: "k8s-ig--clusterUIDdef", Zone: zone}},
			expecetedIGs:   1,
			expectedError:  false,
		},
		{
			name:           "No instance groups for cluster",
			instanceGroups: []*compute.InstanceGroup{{Name: "k8s-ig--cluster2UIDdef", Zone: zone}},
			ports:          []int64{22, 80, 8888},
			expecetedIGs:   0,
			expectedError:  true,
		},
		{
			name: "Ports set to 2 instance groups",
			instanceGroups: []*compute.InstanceGroup{
				{Name: "k8s-ig--clusterUIDdef", Zone: zone},
				{Name: "k8s-ig--clusterUIDdef-3", Zone: zone},
			},
			ports:         []int64{22, 80, 8888},
			expecetedIGs:  2,
			expectedError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeCloud.instanceGroups = tc.instanceGroups
			result, err := multiIGInst.EnsureInstanceGroupsAndPorts(igName, tc.ports)
			if len(result) != tc.expecetedIGs {
				t.Errorf("Incorrect number of instance grouos, got: %v, want: %v", len(result), tc.expecetedIGs)
			}
			gotError := (err != nil)
			if gotError != tc.expectedError {
				t.Errorf("Unexpected error, got: %v, want no error", err)
			}
			for _, resultIG := range result {
				if len(resultIG.NamedPorts) != len(tc.ports) {
					t.Errorf("unexpected number of named ports. Got: %+v, want: %v", len(resultIG.NamedPorts), len(tc.ports))
				}
				for i, np := range resultIG.NamedPorts {
					expectedPort := tc.ports[i]
					if np.Port != expectedPort || np.Name != fmt.Sprint("port", expectedPort) {
						t.Errorf("invalid ports set on instance group. IG named port: %+v, want: %v", expectedPort, tc.ports)
					}
				}
			}
		})
	}
}
