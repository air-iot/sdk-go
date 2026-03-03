//go:build linux

package license

import "runtime"

func getLibName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "license_core_linux_amd64.so"
	case "arm64":
		return "license_core_linux_arm64.so"
	case "loong64":
		return "license_core_linux_loong64.so"
	default:
		return "license_core.so"
	}
}

func getLibNameFallbacks() []string {
	primary := getLibName()
	if primary == "license_core.so" {
		return []string{primary}
	}
	return []string{primary, "license_core.so"}
}
