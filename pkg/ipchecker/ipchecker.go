package ipchecker

import "net"

type IPChecker struct {
	trustedSubnet *net.IPNet
}

func New(subnet string) *IPChecker {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil
	}
	return &IPChecker{trustedSubnet: ipNet}
}

func (ic *IPChecker) Contains(ip *net.IP) bool {
	return ic.trustedSubnet.Contains(*ip)
}
