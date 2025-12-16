package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// toma una lista de targets y devuelve una lista de IPs
func ParseTarget(target string) ([]string, error) {
	//CIDR
	if strings.Contains(target, "/") {
		return parseCIDR(target)
	}

	//rango manual
	if strings.Contains(target, "-") {
		return parseRange(target)
	}

	//IP Unica (o host)
	ip := net.ParseIP(target)
	if ip != nil {
		return []string{ip.String()}, nil
	}

	//si no es IP, puede ser hostname
	ips, err := net.LookupHost(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target or hostname resolution failed: %s", target)
	}
	// retornamos la primera IP resuelta por simplicidad
	if len(ips) > 0 {
		return []string{ips[0]}, nil
	}

	return nil, fmt.Errorf("could not resolve target: %s", target)
}

// parseCIDR toma una CIDR y devuelve una lista de IPs
func parseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	//escanear todo el rango es mas seguro para discovery
	if len(ips) > 2 {
		return ips[1 : len(ips)-1], nil
	}
	return ips, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// establece rango manual
func parseRange(rangeStr string) ([]string, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format")
	}

	startIPStr := parts[0]
	endPart := parts[1] // puede ser una IP completa o solo el ultimo octeto

	startIP := net.ParseIP(startIPStr)
	if startIP == nil {
		return nil, fmt.Errorf("invalid start IP")
	}

	//determinar si endPart es numero o IP
	if strings.Contains(endPart, ".") {
		//PENDIENTE: incremento de IP
		return nil, fmt.Errorf("full IP range not supported yet, use last octet range (e.g. 192.168.1.1-50)")
	}

	//solo octeto final, por eso To4()
	//ParseIP devuelve 16 bytes para IPv6
	v4 := startIP.To4()
	if v4 == nil {
		return nil, fmt.Errorf("IPv6 ranges not supported yet")
	}

	startVal := int(v4[3])
	endVal, err := strconv.Atoi(endPart)
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %v", err)
	}

	if endVal < startVal {
		return nil, fmt.Errorf("end range smaller than start")
	}

	var ips []string
	base := v4[:3] //<- primeros 3 bytes
	for i := startVal; i <= endVal; i++ {
		if i > 255 {
			break
		}
		newIP := net.IPv4(base[0], base[1], base[2], byte(i))
		ips = append(ips, newIP.String())
	}

	return ips, nil
}
