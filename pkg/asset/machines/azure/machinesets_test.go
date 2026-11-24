package azure

import (
	"testing"

	azureenv "github.com/Azure/go-autorest/autorest/azure"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
	capz "sigs.k8s.io/cluster-api-provider-azure/api/v1beta1"

	machineapi "github.com/openshift/api/machine/v1beta1"
	"github.com/openshift/installer/pkg/asset/installconfig"
	icazure "github.com/openshift/installer/pkg/asset/installconfig/azure"
	"github.com/openshift/installer/pkg/types"
	"github.com/openshift/installer/pkg/types/azure"
)

func TestMachineSetsNonZonalRegion(t *testing.T) {
	capabilities := map[string]string{
		"HyperVGenerations":            "V1,V2",
		"AcceleratedNetworkingEnabled": "True",
	}
	session := &icazure.Session{
		Credentials: icazure.Credentials{SubscriptionID: "sub"},
		Environment: azureenv.PublicCloud,
	}

	cases := []struct {
		name         string
		region       string
		zones        []string
		replicas     int64
		wantSets     int
		wantReplicas int32
		wantZone     string
	}{
		{
			name:         "uaecentral single machineset",
			region:       "uaecentral",
			zones:        []string{""},
			replicas:     3,
			wantSets:     1,
			wantReplicas: 3,
			wantZone:     "",
		},
		{
			name:         "uaenorth one machineset per zone",
			region:       "uaenorth",
			zones:        []string{"1", "2", "3"},
			replicas:     3,
			wantSets:     3,
			wantReplicas: 1,
			wantZone:     "1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ic := installconfig.MakeAsset(&types.InstallConfig{
				Platform: types.Platform{
					Azure: &azure.Platform{
						Region:    tc.region,
						CloudName: azure.PublicCloud,
					},
				},
			})
			pool := &types.MachinePool{
				Name:     "worker",
				Replicas: pointer.Int64(tc.replicas),
				Platform: types.MachinePoolPlatform{
					Azure: &azure.MachinePool{
						InstanceType: "Standard_D4s_v3",
						Zones:        append([]string(nil), tc.zones...),
						OSDisk: azure.OSDisk{
							DiskSizeGB: 128,
							DiskType:   "Premium_LRS",
						},
						Identity: &azure.VMIdentity{
							Type: capz.VMIdentityNone,
						},
					},
				},
			}

			sets, err := MachineSets("test-cluster", ic, pool, "rhcos", "worker", "worker-user-data", capabilities, nil, session)
			assert.NoError(t, err)
			assert.Len(t, sets, tc.wantSets)
			assert.Equal(t, tc.wantReplicas, *sets[0].Spec.Replicas)

			spec, ok := sets[0].Spec.Template.Spec.ProviderSpec.Value.Object.(*machineapi.AzureMachineProviderSpec)
			assert.True(t, ok)
			assert.Equal(t, tc.wantZone, spec.Zone)
		})
	}
}
