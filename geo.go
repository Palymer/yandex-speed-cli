package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

// geoLookup — данные ipwho.is (HTTPS, без ключа).
type geoLookup struct {
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	OK      bool   `json:"-"`
	Err     string `json:"-"`
}

func lookupGeoIP(ctx context.Context, client *http.Client, ip string) *geoLookup {
	out := &geoLookup{}
	ip = strings.TrimSpace(ip)
	if ip == "" {
		out.Err = "пустой адрес"
		return out
	}
	p := net.ParseIP(ip)
	if p == nil {
		out.Err = "некорректный IP"
		return out
	}
	if p.IsLoopback() || p.IsPrivate() || p.IsUnspecified() || p.IsLinkLocalUnicast() || p.IsLinkLocalMulticast() {
		out.Err = "локальный адрес"
		return out
	}

	url := "https://ipwho.is/" + ip
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		out.Err = err.Error()
		return out
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; yandex-speed-cli)")
	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	res, err := client.Do(req)
	if err != nil {
		out.Err = err.Error()
		return out
	}
	defer res.Body.Close()

	var raw struct {
		Success bool   `json:"success"`
		City    string `json:"city"`
		Region  string `json:"region"`
		Country string `json:"country"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		out.Err = err.Error()
		return out
	}
	if !raw.Success {
		if strings.TrimSpace(raw.Message) != "" {
			out.Err = raw.Message
		} else {
			out.Err = "geo: success=false"
		}
		return out
	}
	out.City = strings.TrimSpace(raw.City)
	out.Region = strings.TrimSpace(raw.Region)
	out.Country = strings.TrimSpace(raw.Country)
	out.OK = true
	return out
}

func classifyPublicIP(ip string) (ipv4, ipv6 string) {
	ipv4, ipv6 = "не определен", "не определен"
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ipv4, ipv6
	}
	p := net.ParseIP(ip)
	if p == nil {
		return ipv4, ipv6
	}
	if v4 := p.To4(); v4 != nil {
		return v4.String(), "не определен"
	}
	return "не определен", p.String()
}

func regionByIPLabel(g *geoLookup) string {
	if g == nil || !g.OK {
		return "не определен"
	}
	if g.City != "" {
		return g.City
	}
	if g.Region != "" {
		return g.Region
	}
	if g.Country != "" {
		return g.Country
	}
	return "не определен"
}

func regionLabel(g *geoLookup) string {
	if g == nil || !g.OK {
		return "не определен"
	}
	if g.Region != "" {
		return g.Region
	}
	if g.City != "" {
		return g.City
	}
	if g.Country != "" {
		return g.Country
	}
	return "не определен"
}
