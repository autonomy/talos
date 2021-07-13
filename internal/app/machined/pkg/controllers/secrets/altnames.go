// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package secrets

import "net"

// AltNames defines certificate alternative names.
type AltNames struct {
	IPs      []net.IP
	DNSNames []string
}

// Append list of SANs splitting into IPs/DNS names.
func (altNames *AltNames) Append(sans ...string) {
	for _, san := range sans {
		if ip := net.ParseIP(san); ip != nil {
			altNames.AppendIPs(ip)
		} else {
			altNames.AppendDNSNames(san)
		}
	}
}

// AppendIPs skipping duplicates.
func (altNames *AltNames) AppendIPs(ips ...net.IP) {
	for _, ip := range ips {
		found := false

		for _, addr := range altNames.IPs {
			if addr.Equal(ip) {
				found = true

				break
			}
		}

		if !found {
			altNames.IPs = append(altNames.IPs, ip)
		}
	}
}

// AppendDNSNames skipping duplicates.
func (altNames *AltNames) AppendDNSNames(dnsNames ...string) {
	for _, dnsName := range dnsNames {
		found := false

		for _, name := range altNames.DNSNames {
			if name == dnsName {
				found = true

				break
			}
		}

		if !found {
			altNames.DNSNames = append(altNames.DNSNames, dnsName)
		}
	}
}
