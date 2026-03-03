//go:build windows

package license

import "runtime"

func getLibName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "license_core_windows_amd64.dll"
	case "arm64":
		return "license_core_windows_arm64.dll"
	default:
		return "license_core.dll"
	}
}

func getLibNameFallbacks() []string {
	primary := getLibName()
	if primary == "license_core.dll" {
		return []string{primary}
	}
	return []string{primary, "license_core.dll"}
}
