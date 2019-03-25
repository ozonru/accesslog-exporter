package net

import (
	"net"
)

// IsSubnetContainsIP checks if the passed ip is in subnet
func IsSubnetContainsIP(ip string, subNets []string) (bool, error) {
	for _, subNet := range subNets {
		_, IPNet, err := net.ParseCIDR(subNet)
		if err != nil {
			return false, err
		}

		if IPNet.Contains(net.ParseIP(ip)) {
			return true, nil
		}
	}

	return false, nil
}
