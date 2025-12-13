package utils

import (
	"fmt"
	"net" //para validacion de IP
	"strconv"
	"strings"
)

// verifica si la estructura IP es valida
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// parsea un string de puertos a escanear
func ParsePortRange(portStr string) ([]int, error) {
	var ports []int //slice de puertos validos
	parts := strings.Split(portStr, ",")

	for _, part := range parts {
		if strings.Contains(part, "-") {
			//manejo de rangos
			rangeParts := strings.Split(part, "-")

			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %s", part)
			}

			start, err := strconv.Atoi(rangeParts[0])

			if err != nil {
				return nil, fmt.Errorf("invalid initial port: %s", rangeParts[0])
			}

			end, err := strconv.Atoi(rangeParts[1])

			if err != nil {
				return nil, fmt.Errorf("invalid final port: %s", rangeParts[1])
			}

			if start > end {
				return nil, fmt.Errorf("initial port is greater than final port: %s", part)
			}

			for i := start; i <= end; i++ {
				if isValidPort(i) {
					ports = append(ports, i)
				}
			}
		} else {
			//manejo de puertos individuales
			port, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid port: %s", part)
			}
			if isValidPort(port) {
				ports = append(ports, port)
			}
		}
	}
	return ports, nil
}

// verifica si el numero del puerto es valido
func isValidPort(port int) bool {
	return port > 0 && port <= 65535
}
