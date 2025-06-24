package utils

import (
	"fmt"
	"net"

	"github.com/madlabx/pkgx/log"
)

func GetIpAddr(deviceName string) (string, error) {

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Name == deviceName {

			addrs, err := iface.Addrs()
			if err != nil {
				log.Errorf("Ignore error getting addresses for interface %v, err:%v", iface.Name, err)
				continue
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
					return ipnet.IP.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("cannot find " + deviceName)
}
