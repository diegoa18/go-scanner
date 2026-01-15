package service

import (
	"strings"
)

// tipo de servicio detectado en un string
type ServiceType string

const (
	ServiceUnknown ServiceType = "unknown"
	ServiceHTTP    ServiceType = "HTTP"
	ServiceHTTPS   ServiceType = "HTTPS"
	ServiceSSH     ServiceType = "SSH"
	ServiceFTP     ServiceType = "FTP"
	ServiceSMTP    ServiceType = "SMTP"
	ServicePOP3    ServiceType = "POP3"
	ServiceIMAP    ServiceType = "IMAP"
	ServiceDNS     ServiceType = "DNS"
	ServiceLDP     ServiceType = "LDP"
)

// como fue detectado el servicio
type Method string

const (
	MethodBanner Method = "banner"
	MethodPort   Method = "port" //heuristica por puerto
	MethodNone   Method = "none"
)

// contiene la informacion del servicio detectado
type ServiceInfo struct {
	Type   ServiceType
	Method Method
}

// mapa de puertos comunes por defecto
var commonPorts = map[int]ServiceType{
	21:  ServiceFTP,
	22:  ServiceSSH,
	25:  ServiceSMTP,
	53:  ServiceDNS,
	80:  ServiceHTTP,
	110: ServicePOP3,
	143: ServiceIMAP,
	443: ServiceHTTPS,
	465: ServiceSMTP, //SMTPS
	646: ServiceLDP,
	993: ServiceIMAP, //IMAPS
	995: ServicePOP3, //POP3S
}

// inferir el servicio a raiz del puerto y banner
func Detect(port int, banner string) ServiceInfo {
	//mediante banner
	if banner != "" {
		lowerBanner := strings.ToLower(banner)
		if strings.HasPrefix(lowerBanner, "ssh-") {
			return ServiceInfo{Type: ServiceSSH, Method: MethodBanner}
		}
		if strings.HasPrefix(lowerBanner, "220 ") {
			//FTP y SMTP empiezan en 220 normalmente, asi que...
			if strings.Contains(lowerBanner, "ftp") {
				return ServiceInfo{Type: ServiceFTP, Method: MethodBanner}
			}
			if strings.Contains(lowerBanner, "smtp") || strings.Contains(lowerBanner, "mail") {
				return ServiceInfo{Type: ServiceSMTP, Method: MethodBanner}
			}
		}
	}

	//mediante puerto
	if svc, ok := commonPorts[port]; ok {
		return ServiceInfo{Type: svc, Method: MethodPort}
	}

	//desconocido
	return ServiceInfo{Type: ServiceUnknown, Method: MethodNone}
}
