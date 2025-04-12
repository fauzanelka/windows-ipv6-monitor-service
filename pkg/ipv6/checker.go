package ipv6

import (
	"io"
	"net/http"
	"strings"
)

// GetGlobalIPv6 retrieves the global IPv6 address using an external service
func GetGlobalIPv6() (string, error) {
	// Using ipv6.icanhazip.com to get the IPv6 address
	resp, err := http.Get("https://ipv6.icanhazip.com")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Clean the response
	ipv6 := strings.TrimSpace(string(body))
	return ipv6, nil
} 