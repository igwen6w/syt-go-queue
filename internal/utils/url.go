package utils

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// 允许的回调URL方案
var allowedSchemes = map[string]bool{
	"http":  true,
	"https": true,
}

// 禁止的IP范围
var forbiddenIPRanges = []struct {
	start net.IP
	end   net.IP
}{
	{net.ParseIP("10.0.0.0"), net.ParseIP("10.255.255.255")},                           // 10.0.0.0/8
	{net.ParseIP("172.16.0.0"), net.ParseIP("172.31.255.255")},                         // 172.16.0.0/12
	{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},                       // 192.168.0.0/16
	{net.ParseIP("127.0.0.0"), net.ParseIP("127.255.255.255")},                         // 127.0.0.0/8
	{net.ParseIP("0.0.0.0"), net.ParseIP("0.255.255.255")},                             // 0.0.0.0/8
	{net.ParseIP("169.254.0.0"), net.ParseIP("169.254.255.255")},                       // 169.254.0.0/16
	{net.ParseIP("192.0.2.0"), net.ParseIP("192.0.2.255")},                             // 192.0.2.0/24
	{net.ParseIP("198.51.100.0"), net.ParseIP("198.51.100.255")},                       // 198.51.100.0/24
	{net.ParseIP("203.0.113.0"), net.ParseIP("203.0.113.255")},                         // 203.0.113.0/24
	{net.ParseIP("224.0.0.0"), net.ParseIP("239.255.255.255")},                         // 224.0.0.0/4
	{net.ParseIP("240.0.0.0"), net.ParseIP("255.255.255.255")},                         // 240.0.0.0/4
	{net.ParseIP("100.64.0.0"), net.ParseIP("100.127.255.255")},                        // 100.64.0.0/10
	{net.ParseIP("192.0.0.0"), net.ParseIP("192.0.0.255")},                             // 192.0.0.0/24
	{net.ParseIP("198.18.0.0"), net.ParseIP("198.19.255.255")},                         // 198.18.0.0/15
	{net.ParseIP("2001:db8::"), net.ParseIP("2001:db8:ffff:ffff:ffff:ffff:ffff:ffff")}, // 2001:db8::/32
	{net.ParseIP("fc00::"), net.ParseIP("fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},    // fc00::/7
	{net.ParseIP("fe80::"), net.ParseIP("febf:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},    // fe80::/10
	{net.ParseIP("::1"), net.ParseIP("::1")},                                           // ::1/128
	{net.ParseIP("::"), net.ParseIP("::")},                                             // ::/128
}

// 禁止的主机名
var forbiddenHostnames = map[string]bool{
	"localhost": true,
	"127.0.0.1": true,
	"::1":       true,
	"0.0.0.0":   true,
	"[::1]":     true,
	"[::0]":     true,
}

// ipInRange 检查IP是否在指定范围内
func ipInRange(ip net.IP, start net.IP, end net.IP) bool {
	if ip.To4() != nil {
		ip = ip.To4()
		start = start.To4()
		end = end.To4()
	}

	if ip == nil || start == nil || end == nil {
		return false
	}

	for i := 0; i < len(ip); i++ {
		if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
		if ip[i] != start[i] && ip[i] != end[i] {
			break
		}
	}
	return true
}

// ValidateCallbackURL 验证回调URL是否安全
// 防止SSRF攻击
func ValidateCallbackURL(callbackURL string) error {
	// 解析URL
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// 检查URL方案
	if !allowedSchemes[parsedURL.Scheme] {
		return fmt.Errorf("URL scheme not allowed: %s", parsedURL.Scheme)
	}

	// 检查主机名
	hostname := parsedURL.Hostname()
	if forbiddenHostnames[strings.ToLower(hostname)] {
		return fmt.Errorf("hostname not allowed: %s", hostname)
	}

	// 解析IP地址
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	// 检查每个IP是否在禁止范围内
	for _, ip := range ips {
		for _, ipRange := range forbiddenIPRanges {
			if ipInRange(ip, ipRange.start, ipRange.end) {
				return fmt.Errorf("IP address in forbidden range: %s", ip.String())
			}
		}
	}

	return nil
}
