//go:build darwin

package license

import "runtime"

func getLibName() string {
	switch runtime.GOARCH {
	case "amd64":
		return "license_core_darwin_amd64.dylib"
	case "arm64":
		return "license_core_darwin_arm64.dylib"
	default:
		return "license_core.dylib"
	}
}

func getLibNameFallbacks() []string {
	primary := getLibName()
	if primary == "license_core.dylib" {
		return []string{primary}
	}
	return []string{primary, "license_core.dylib"}
}
