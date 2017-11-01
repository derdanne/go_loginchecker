package main

import "net"

func isNotAllowedIp(hostname string, allowedNetworks []string) bool {
	parsedAddress := net.ParseIP(hostname)
	if parsedAddress == nil {
		for _, allowedHostname := range allowedNetworks {
			if allowedHostname == hostname {
				return false
			}
		}
	} else {
		for _, allowedNetwork := range allowedNetworks {
			_, parsedAllowedNetworkCIDR, _ := net.ParseCIDR(allowedNetwork)
			if parsedAllowedNetworkCIDR == nil {
			} else {
				if parsedAllowedNetworkCIDR.Contains(parsedAddress) {
					return false
				}
			}
		}
	}
	return true
}

func isNotAllowedUser(username string, allowedUsers []string) bool {
	for _, allowedUser := range allowedUsers {
		if allowedUser == username {
			return false
		}
	}
	return true
}
