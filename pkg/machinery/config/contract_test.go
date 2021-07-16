// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/talos-systems/talos/pkg/machinery/config"
)

func TestContractGreater(t *testing.T) {
	assert.True(t, config.TalosVersion0_9.Greater(config.TalosVersion0_8))
	assert.True(t, config.TalosVersionCurrent.Greater(config.TalosVersion0_8))
	assert.True(t, config.TalosVersionCurrent.Greater(config.TalosVersion0_9))

	assert.False(t, config.TalosVersion0_8.Greater(config.TalosVersion0_9))
	assert.False(t, config.TalosVersion0_8.Greater(config.TalosVersion0_8))
	assert.False(t, config.TalosVersionCurrent.Greater(config.TalosVersionCurrent))
}

func TestContractParseVersion(t *testing.T) {
	t.Parallel()

	for v, expected := range map[string]*config.VersionContract{
		"v0.8":           config.TalosVersion0_8,
		"v0.8.":          config.TalosVersion0_8,
		"v0.8.1":         config.TalosVersion0_8,
		"v0.88":          {0, 88},
		"v0.8.3-alpha.4": config.TalosVersion0_8,
	} {
		v, expected := v, expected
		t.Run(v, func(t *testing.T) {
			t.Parallel()

			actual, err := config.ParseContractFromVersion(v)
			assert.NoError(t, err)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestContractCurrent(t *testing.T) {
	assert.True(t, config.TalosVersionCurrent.SupportsAggregatorCA())
	assert.True(t, config.TalosVersionCurrent.SupportsECDSAKeys())
	assert.True(t, config.TalosVersionCurrent.SupportsServiceAccount())
	assert.True(t, config.TalosVersionCurrent.SupportsRBACFeature())
	assert.True(t, config.TalosVersionCurrent.SupportsDynamicCertSANs())
}

func TestContract0_12(t *testing.T) {
	assert.True(t, config.TalosVersion0_12.SupportsAggregatorCA())
	assert.True(t, config.TalosVersion0_12.SupportsECDSAKeys())
	assert.True(t, config.TalosVersion0_12.SupportsServiceAccount())
	assert.True(t, config.TalosVersion0_12.SupportsRBACFeature())
	assert.True(t, config.TalosVersion0_12.SupportsDynamicCertSANs())
}

func TestContract0_11(t *testing.T) {
	assert.True(t, config.TalosVersion0_11.SupportsAggregatorCA())
	assert.True(t, config.TalosVersion0_11.SupportsECDSAKeys())
	assert.True(t, config.TalosVersion0_11.SupportsServiceAccount())
	assert.True(t, config.TalosVersion0_11.SupportsRBACFeature())
	assert.False(t, config.TalosVersion0_11.SupportsDynamicCertSANs())
}

func TestContract0_10(t *testing.T) {
	assert.True(t, config.TalosVersion0_10.SupportsAggregatorCA())
	assert.True(t, config.TalosVersion0_10.SupportsECDSAKeys())
	assert.True(t, config.TalosVersion0_10.SupportsServiceAccount())
	assert.False(t, config.TalosVersion0_10.SupportsRBACFeature())
	assert.False(t, config.TalosVersion0_10.SupportsDynamicCertSANs())
}

func TestContract0_9(t *testing.T) {
	assert.True(t, config.TalosVersion0_9.SupportsAggregatorCA())
	assert.True(t, config.TalosVersion0_9.SupportsECDSAKeys())
	assert.True(t, config.TalosVersion0_9.SupportsServiceAccount())
	assert.False(t, config.TalosVersion0_9.SupportsRBACFeature())
	assert.False(t, config.TalosVersion0_9.SupportsDynamicCertSANs())
}

func TestContract0_8(t *testing.T) {
	assert.False(t, config.TalosVersion0_8.SupportsAggregatorCA())
	assert.False(t, config.TalosVersion0_8.SupportsECDSAKeys())
	assert.False(t, config.TalosVersion0_8.SupportsServiceAccount())
	assert.False(t, config.TalosVersion0_8.SupportsRBACFeature())
	assert.False(t, config.TalosVersion0_8.SupportsDynamicCertSANs())
}
