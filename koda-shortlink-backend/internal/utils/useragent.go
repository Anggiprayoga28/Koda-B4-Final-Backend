package utils

import (
	"strings"
)

type DeviceInfo struct {
	DeviceType string
	Browser    string
	OS         string
}

func ParseUserAgent(userAgent string) *DeviceInfo {
	info := &DeviceInfo{
		DeviceType: detectDeviceType(userAgent),
		Browser:    detectBrowser(userAgent),
		OS:         detectOS(userAgent),
	}
	return info
}

func detectDeviceType(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") ||
		strings.Contains(ua, "iphone") || strings.Contains(ua, "ipod") {
		return "mobile"
	}

	if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		return "tablet"
	}

	return "desktop"
}

func detectBrowser(userAgent string) string {
	ua := strings.ToLower(userAgent)

	browsers := map[string]string{
		"edg/":     "Edge",
		"chrome/":  "Chrome",
		"safari/":  "Safari",
		"firefox/": "Firefox",
		"opera/":   "Opera",
		"opr/":     "Opera",
		"trident/": "Internet Explorer",
	}

	for key, browser := range browsers {
		if strings.Contains(ua, key) {
			if browser == "Safari" && strings.Contains(ua, "chrome/") {
				return "Chrome"
			}
			return browser
		}
	}

	return "Unknown"
}

func detectOS(userAgent string) string {
	ua := strings.ToLower(userAgent)

	if strings.Contains(ua, "windows nt 10.0") {
		return "Windows 10"
	}
	if strings.Contains(ua, "windows nt 6.3") {
		return "Windows 8.1"
	}
	if strings.Contains(ua, "windows nt 6.2") {
		return "Windows 8"
	}
	if strings.Contains(ua, "windows nt 6.1") {
		return "Windows 7"
	}
	if strings.Contains(ua, "windows") {
		return "Windows"
	}

	if strings.Contains(ua, "mac os x") {
		return "macOS"
	}

	if strings.Contains(ua, "android") {
		return "Android"
	}

	if strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ipod") {
		return "iOS"
	}

	if strings.Contains(ua, "linux") {
		return "Linux"
	}

	return "Unknown"
}
