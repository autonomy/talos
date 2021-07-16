// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package config

import (
	"fmt"
	"regexp"
	"strconv"
)

// VersionContract describes Talos version to generate config for.
//
// Config generation only supports backwards compatibility (e.g. Talos 0.9 can generate configs for Talos 0.9 and 0.8).
// Matching version of the machinery package is required to generate configs for the current version of Talos.
//
// Nil value of *VersionContract always describes current version of Talos.
type VersionContract struct {
	Major int
	Minor int
}

// Well-known Talos version contracts.
var (
	TalosVersionCurrent = (*VersionContract)(nil)
	TalosVersion0_12    = &VersionContract{0, 12}
	TalosVersion0_11    = &VersionContract{0, 11}
	TalosVersion0_10    = &VersionContract{0, 10}
	TalosVersion0_9     = &VersionContract{0, 9}
	TalosVersion0_8     = &VersionContract{0, 8}
)

var versionRegexp = regexp.MustCompile(`^v(\d+)\.(\d+)($|\.)`)

// ParseContractFromVersion parses Talos version into VersionContract.
func ParseContractFromVersion(version string) (*VersionContract, error) {
	matches := versionRegexp.FindStringSubmatch(version)
	if len(matches) < 3 {
		return nil, fmt.Errorf("error parsing version %q", version)
	}

	var contract VersionContract

	contract.Major, _ = strconv.Atoi(matches[1]) //nolint:errcheck
	contract.Minor, _ = strconv.Atoi(matches[2]) //nolint:errcheck

	return &contract, nil
}

// Greater compares contract to another contract.
func (contract *VersionContract) Greater(other *VersionContract) bool {
	if contract == nil {
		return other != nil
	}

	if other == nil {
		return false
	}

	return contract.Major > other.Major || (contract.Major == other.Major && contract.Minor > other.Minor)
}

// SupportsECDSAKeys returns true if version of Talos supports ECDSA keys (vs. RSA keys).
func (contract *VersionContract) SupportsECDSAKeys() bool {
	return contract.Greater(TalosVersion0_8)
}

// SupportsAggregatorCA returns true if version of Talos supports AggregatorCA in the config.
func (contract *VersionContract) SupportsAggregatorCA() bool {
	return contract.Greater(TalosVersion0_8)
}

// SupportsServiceAccount returns true if version of Talos supports ServiceAccount in the config.
func (contract *VersionContract) SupportsServiceAccount() bool {
	return contract.Greater(TalosVersion0_8)
}

// SupportsRBACFeature returns true if version of Talos supports RBAC feature gate.
func (contract *VersionContract) SupportsRBACFeature() bool {
	return contract.Greater(TalosVersion0_10)
}

// SupportsDynamicCertSANs returns true if version of Talos supports dynamic certificate generation with SANs provided from resources.
func (contract *VersionContract) SupportsDynamicCertSANs() bool {
	return contract.Greater(TalosVersion0_11)
}
