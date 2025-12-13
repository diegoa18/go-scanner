package utils

import (
	"fmt"
	"net"
)

// intenta resolver una IP, si esta "cruda" la utiliza como taL
func Resolve(target string) (string, error) {
	if IsValidIP(target) { //si es una IP valida
		return target, nil
	}

	//intentar resolver el hostname
	ips, err := net.LookupIP(target)
	if err != nil {
		return "", fmt.Errorf("failed to resolve host: %w", err)
	}

	//priorizar IPv4
	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			return ipv4.String(), nil
		}
	}

	//si no hay IPv4, devolver la primera IP (que seria IPv6)
	if len(ips) > 0 {
		return ips[0].String(), nil
	}

	return "", fmt.Errorf("no IP addresses found for host: %s", target)
}
