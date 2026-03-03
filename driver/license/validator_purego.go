//go:build !windows

package license

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/air-iot/errors"
	"github.com/air-iot/logger"
	"github.com/ebitengine/purego"
)

var (
	libMu      sync.Mutex
	libHandle  uintptr
	libLoadErr error
)

var (
	fnGetMachineCode        func(*byte, int) int
	fnGetLastError          func() string
	fnVerifyLicense         func(*byte, *byte, *byte) int
	fnVerifyDriverLicenseJS func(*byte, int, *byte, int, *byte, *byte, int, *byte, int) int
)

func loadLibrary() error {
	libMu.Lock()
	defer libMu.Unlock()

	if libHandle != 0 && libLoadErr == nil {
		return nil
	}

	libPath := findLibrary()
	handle, err := purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		libLoadErr = err
		logger.Errorf("load driver license library failed: %v", err)
		return err
	}

	getMachineCodeAddr, err := purego.Dlsym(handle, "get_machine_code")
	if err != nil {
		_ = purego.Dlclose(handle)
		libLoadErr = err
		return err
	}

	verifyAddr, err := purego.Dlsym(handle, "verify_driver_license_json")
	if err != nil {
		_ = purego.Dlclose(handle)
		libLoadErr = err
		return err
	}

	verifyLicenseAddr, err := purego.Dlsym(handle, "VerifyLicense")
	if err != nil {
		verifyLicenseAddr = 0
	}

	getLastErrorAddr, err := purego.Dlsym(handle, "get_last_error")
	if err != nil {
		getLastErrorAddr = 0
	}

	var getMachineCode func(*byte, int) int
	purego.RegisterFunc(&getMachineCode, getMachineCodeAddr)

	var verifyDriverLicenseJSON func(*byte, int, *byte, int, *byte, *byte, int, *byte, int) int
	purego.RegisterFunc(&verifyDriverLicenseJSON, verifyAddr)

	var verifyLicense func(*byte, *byte, *byte) int
	if verifyLicenseAddr != 0 {
		purego.RegisterFunc(&verifyLicense, verifyLicenseAddr)
	}

	var getLastError func() string
	if getLastErrorAddr != 0 {
		purego.RegisterFunc(&getLastError, getLastErrorAddr)
	}

	libHandle = handle
	fnGetMachineCode = getMachineCode
	fnVerifyDriverLicenseJS = verifyDriverLicenseJSON
	fnVerifyLicense = verifyLicense
	fnGetLastError = getLastError
	libLoadErr = nil
	return nil
}

func findLibrary() string {
	for _, libName := range getLibNameFallbacks() {
		searchPaths := []string{
			libName,
			filepath.Join("lib", libName),
			filepath.Join("license", "lib", libName),
			filepath.Join(filepath.Dir(os.Args[0]), libName),
			filepath.Join(filepath.Dir(os.Args[0]), "lib", libName),
			libName,
		}
		for _, path := range searchPaths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
	}
	return getLibName()
}

func GetMachineCodeFromLib() (string, error) {
	if err := loadLibrary(); err != nil {
		return "", fmt.Errorf("load driver license library failed: %w", err)
	}

	var output [64]byte
	ret := fnGetMachineCode(&output[0], len(output))
	if ret != 0 {
		errMsg := GetLastErrorFromLib()
		return "", errors.New400Response(100050049, "get machine code failed: %s", errMsg)
	}

	return goStringFromBytes(output[:]), nil
}

func GetLastErrorFromLib() string {
	if fnGetLastError == nil {
		return "unknown error"
	}
	return fnGetLastError()
}

func goStringFromBytes(b []byte) string {
	var i int
	for i = 0; i < len(b) && b[i] != 0; i++ {
	}
	return string(b[:i])
}

func verifyDriverLicenseJSONRaw(payload, signature []byte, payloadJSON string) (string, error) {
	if err := loadLibrary(); err != nil {
		return "", fmt.Errorf("load driver license library failed: %w", err)
	}
	if fnVerifyDriverLicenseJS == nil {
		return "", fmt.Errorf("verify_driver_license_json symbol unavailable")
	}

	payloadJSONBuf := cStringBytes(payloadJSON)
	var resultBuf [4096]byte
	var errorBuf [256]byte

	ret := fnVerifyDriverLicenseJS(
		sliceDataOrNil(payload),
		len(payload),
		sliceDataOrNil(signature),
		len(signature),
		&payloadJSONBuf[0],
		&resultBuf[0],
		len(resultBuf),
		&errorBuf[0],
		len(errorBuf),
	)
	result := goStringFromBytes(resultBuf[:])
	if ret != 0 {
		errStr := goStringFromBytes(errorBuf[:])
		if errStr == "" {
			errStr = GetLastErrorFromLib()
		}
		if errStr == "" {
			errStr = "driver license verification failed"
		}
		if result == "" {
			result = fmt.Sprintf(`{"ok":false,"message":"%s"}`, errStr)
		}
		return result, fmt.Errorf("driver license verification failed(code %d): %s", ret, errStr)
	}
	return result, nil
}

func sliceDataOrNil(b []byte) *byte {
	if len(b) == 0 {
		return nil
	}
	return &b[0]
}

func cStringBytes(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

func verifyLicenseRawFromLib(licensePath, driverID, data string) (*DriverLicenseVerifyResultFromLib, error) {
	if err := loadLibrary(); err != nil {
		return nil, fmt.Errorf("load driver license library failed: %w", err)
	}
	if fnVerifyLicense == nil {
		return nil, fmt.Errorf("VerifyLicense symbol unavailable")
	}

	licensePathBuf := cStringBytes(licensePath)
	driverIDBuf := cStringBytes(driverID)
	dataBuf := cStringBytes(data)

	ret := fnVerifyLicense(
		&licensePathBuf[0],
		&driverIDBuf[0],
		&dataBuf[0],
	)
	raw := GetLastErrorFromLib()
	logger.Infof("VerifyLicenseFromLib ret=%d raw=%s", ret, raw)
	return parseVerifyLicenseResultFromLib(raw, ret)
}

func VerifyLicenseFromLib(licensePath, driverID, data string) (bool, *DriverLicenseInfoFromLib, error) {
	lp := strings.TrimSpace(licensePath)
	if lp == "" {
		logger.Warnf("VerifyLicenseFromLib treat empty path as no license")
		return fallbackVerifyAsNoLicense(data)
	}
	stat, statErr := os.Stat(lp)
	if statErr != nil || !stat.IsDir() {
		logger.Warnf("VerifyLicenseFromLib treat invalid path as no license, path=%s err=%v", licensePath, statErr)
		return fallbackVerifyAsNoLicense(data)
	}

	result, err := verifyLicenseRawFromLib(licensePath, driverID, data)
	if err != nil && shouldTreatPathAsNoLicense(err) {
		logger.Warnf("VerifyLicenseFromLib treat invalid path as no license, path=%s err=%v", licensePath, err)
		return fallbackVerifyAsNoLicense(data)
	}
	if result == nil {
		return false, nil, err
	}
	return result.OK, result.License, err
}

func CloseLibrary() {
	libMu.Lock()
	defer libMu.Unlock()

	if libHandle != 0 {
		_ = purego.Dlclose(libHandle)
	}
	libHandle = 0
	fnGetMachineCode = nil
	fnVerifyDriverLicenseJS = nil
	fnVerifyLicense = nil
	fnGetLastError = nil
	libLoadErr = nil
}

func IsLibraryLoaded() bool {
	libMu.Lock()
	defer libMu.Unlock()
	return libHandle != 0 && libLoadErr == nil
}
