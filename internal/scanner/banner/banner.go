package banner

import (
	"bufio"
	"net"
	"strings"
	"time"
)

// intenta leer un banner de manera pasiva, solo a puertos conocidos
func Grab(conn net.Conn, port int) (string, error) {
	allowedPorts := map[int]bool{
		21:  true, //FTP
		22:  true, //SSH
		25:  true, //SMTP
		110: true, //POP3
		143: true, //IMAP
	}

	if !allowedPorts[port] {
		return "", nil
	}

	//timeout estricto para lectura
	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return "", err
	}

	//buffer para lectura
	reader := bufio.NewReader(conn)
	//limite de 1024 bytes (evita DoS o problemas de memoria)
	//el banner es solo informacion ligera
	buffer := make([]byte, 1024)
	n, err := reader.Read(buffer)

	if err != nil {
		//normal que falle :p
		return "", nil
	}

	//limpiar el banner
	banner := string(buffer[:n])
	banner = strings.TrimSpace(banner) //eliminacion de espacios

	return banner, nil
}
