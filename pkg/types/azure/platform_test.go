package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetBaseDomain(t *testing.T) {
	platform := Platform{}
	zoneID := "/subscriptions/<subid>/resourceGroups/<rg_name>/providers/Microsoft.Network/dnszones/<zone_name>"
	platform.SetBaseDomain(zoneID)
	assert.Equal(t, "<rg_name>", platform.BaseDomainResourceGroupName)
}

func TestIsNonZonalRegion(t *testing.T) {
	cases := []struct {
		name     string
		region   string
		expected bool
	}{
		{name: "uaecentral", region: "uaecentral", expected: true},
		{name: "uaecentral mixed case", region: "UAECentral", expected: true},
		{name: "uaenorth", region: "uaenorth", expected: false},
		{name: "eastus", region: "eastus", expected: false},
		{name: "empty", region: "", expected: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Platform{Region: tc.region}
			assert.Equal(t, tc.expected, p.IsNonZonalRegion())
		})
	}
}
