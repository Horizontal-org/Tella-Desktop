package network

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func GetWiFiNetworkName() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getWiFiNetworkNameMacOS()
	case "windows":
		return getWiFiNetworkNameWindows()
	case "linux":
		return getWiFiNetworkNameLinux()
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func getWiFiNetworkNameMacOS() (string, error) {
	methods := []func() (string, error){
		ssidViaNetworksetup,
		ssidViaAirport,
		ssidViaSystemProfiler,
	}

	for _, method := range methods {
		if ssid, _ := method(); ssid != "" {
			return ssid, nil
		}
	}

	return "", fmt.Errorf("no Wi-Fi network detected")
}

func ssidViaNetworksetup() (string, error) {
	out, err := exec.Command("networksetup", "-listallhardwareports").Output()
	if err != nil {
		return "", err
	}

	iface := findWiFiInterface(out)
	air, err := exec.Command("networksetup", "-getairportnetwork", iface).Output()
	if err != nil {
		return "", err
	}

	return parseNetworksetupOutput(string(air)), nil
}

func findWiFiInterface(output []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		if strings.Contains(sc.Text(), "Wi-Fi") || strings.Contains(sc.Text(), "AirPort") {
			if sc.Scan() {
				fields := strings.Fields(sc.Text())
				if len(fields) == 2 {
					return fields[1]
				}
			}
			break
		}
	}
	return "en0"
}

func parseNetworksetupOutput(output string) string {
	if strings.Contains(output, "Current Wi-Fi Network:") {
		parts := strings.SplitN(output, ":", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

func ssidViaAirport() (string, error) {
	path := "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	out, err := exec.Command(path, "-I").Output()
	if err != nil {
		return "", err
	}

	return parseAirportOutput(out), nil
}

func parseAirportOutput(output []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "SSID:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		}
	}
	return ""
}

func ssidViaSystemProfiler() (string, error) {
	out, err := exec.Command("system_profiler", "SPAirPortDataType").Output()
	if err != nil {
		return "", err
	}

	return parseSystemProfilerOutput(out), nil
}

func parseSystemProfilerOutput(output []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		if strings.HasPrefix(line, "SSID:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "SSID:"))
		}

		if strings.HasPrefix(line, "Current Network Information:") && sc.Scan() {
			ssid := strings.TrimSpace(strings.TrimSuffix(sc.Text(), ":"))
			if ssid != "" {
				return ssid
			}
		}
	}
	return ""
}

func getWiFiNetworkNameWindows() (string, error) {
	out, err := exec.Command("netsh", "wlan", "show", "interfaces").Output()
	if err != nil {
		return "", err
	}

	return parseWindowsNetshOutput(out), nil
}

func parseWindowsNetshOutput(output []byte) string {
	sc := bufio.NewScanner(bytes.NewReader(output))
	for sc.Scan() {
		line := sc.Text()
		if strings.Contains(line, "SSID") && !strings.Contains(line, "BSSID") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func getWiFiNetworkNameLinux() (string, error) {
	if ssid := tryIwgetid(); ssid != "" {
		return ssid, nil
	}

	return tryNetworkManager()
}

func tryIwgetid() string {
	if path, _ := exec.LookPath("iwgetid"); path != "" {
		if out, err := exec.Command("iwgetid", "-r").Output(); err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return ""
}

func tryNetworkManager() (string, error) {
	out, err := exec.Command("nmcli", "-t", "-f", "active,ssid", "dev", "wifi").Output()
	if err != nil {
		return "", err
	}

	for _, line := range bytes.Split(out, []byte{'\n'}) {
		if bytes.HasPrefix(line, []byte("yes:")) {
			return string(bytes.TrimPrefix(line, []byte("yes:"))), nil
		}
	}

	return "", nil
}
