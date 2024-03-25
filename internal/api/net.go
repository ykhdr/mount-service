package api

import (
	"errors"
	"mount-service/internal/log"
	"mount-service/internal/models"
	"net"
	"net/http"
	"strings"
)

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Logger.WithError(err).Fatal("Error on connecting DNS server")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func findUserByUsername(users []*models.User, username string) (*models.User, int) {
	for i, user := range users {
		if user.Username == username {
			return user, i
		}
	}
	return nil, -1
}

func readUserIP(r *http.Request) (net.IP, error) {
	ipAddr := r.Header.Get("X-Real-Ip")
	if ipAddr == "" {
		ipAddr = r.Header.Get("X-Forwarded-For")
	}
	if ipAddr == "" {
		ipAddr = r.RemoteAddr
	}

	if strings.Contains(ipAddr, "[::1]") || ipAddr == "127.0.0.1" {
		return getOutboundIP(), nil
	}

	resolvedAddr, err := net.ResolveTCPAddr("tcp4", ipAddr)
	if err != nil {
		return nil, errors.New("error on resolving user IP addr")
	}

	return resolvedAddr.IP, nil
}
