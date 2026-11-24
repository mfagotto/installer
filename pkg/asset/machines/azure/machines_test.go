package azure

import (
	"testing"

	azureenv "github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"

	machineapi "github.com/openshift/api/machine/v1beta1"
	icazure "github.com/openshift/installer/pkg/asset/installconfig/azure"
	"github.com/openshift/installer/pkg/types"
	"github.com/openshift/installer/pkg/types/azure"
)

func TestMachinesNonZonalRegion(t *testing.T) {
	capabilities := map[string]string{
		"HyperVGenerations":            "V1,V2",
		"AcceleratedNetworkingEnabled": "True",
	}
	session := &icazure.Session{
		Credentials: icazure.Credentials{SubscriptionID: "sub"},
		Environment: azureenv.PublicCloud,
	}

	cases := []struct {
		name                string
		region              string
		zones               []string
		wantAvailabilitySet string
		wantZone            string
	}{
		{
			name:                "uaecentral uses availability set",
			region:              "uaecentral",
			zones:               []string{""},
			wantAvailabilitySet: "test-cluster_control-plane-as",
			wantZone:            "",
		},
		{
			name:                "uaenorth does not set availability set",
			region:              "uaenorth",
			zones:               []string{"1", "2", "3"},
			wantAvailabilitySet: "",
			wantZone:            "1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := &types.InstallConfig{
				ObjectMeta: metav1.ObjectMeta{Name: "test-cluster"},
				Platform: types.Platform{
					Azure: &azure.Platform{
						Region:    tc.region,
						CloudName: azure.PublicCloud,
					},
				},
			}
			pool := &types.MachinePool{
				Name:     "master",
				Replicas: pointer.Int64(3),
				Platform: types.MachinePoolPlatform{
					Azure: &azure.MachinePool{
						InstanceType: "Standard_D8s_v3",
						Zones:        append([]string(nil), tc.zones...),
						OSDisk: azure.OSDisk{
							DiskSizeGB: 1024,
							DiskType:   "Premium_LRS",
						},
						Identity: &azure.VMIdentity{
							Type: capz.VMIdentityNone,
						},
					},
				},
			}

			machines, _, err := Machines("test-cluster", config, pool, "rhcos", "master", "master-user-data", capabilities, session)
			assert.NoError(t, err)
			assert.Len(t, machines, 3)

			spec, ok := machines[0].Spec.ProviderSpec.Value.Object.(*machineapi.AzureMachineProviderSpec)
			assert.True(t, ok)
			assert.Equal(t, tc.wantZone, spec.Zone)
			assert.Equal(t, tc.wantAvailabilitySet, spec.AvailabilitySet)
		})
	}
}
