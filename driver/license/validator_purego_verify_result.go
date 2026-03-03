package license

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DriverLicenseInfoFromLib struct {
	Path           string   `json:"path,omitempty"`
	LicenseName    string   `json:"licenseName,omitempty"`
	LicenseType    string   `json:"licenseType,omitempty"`
	MacAddr        []string `json:"macAddr,omitempty"`
	CreateTime     string   `json:"createTime,omitempty"`
	ValidityPeriod int      `json:"validityPeriod,omitempty"`
	Deadline       string   `json:"deadline,omitempty"`
}

type DriverLicenseVerifyResultFromLib struct {
	OK       bool                      `json:"ok"`
	TagCount int                       `json:"tagCount"`
	License  *DriverLicenseInfoFromLib `json:"license,omitempty"`
	Error    string                    `json:"error,omitempty"`
}

func parseVerifyLicenseResultFromLib(raw string, code int) (*DriverLicenseVerifyResultFromLib, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if code == 0 {
			return &DriverLicenseVerifyResultFromLib{OK: true}, nil
		}
		return &DriverLicenseVerifyResultFromLib{OK: false}, fmt.Errorf("verify license failed(code %d): empty result", code)
	}

	out := new(DriverLicenseVerifyResultFromLib)
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		if code == 0 {
			return nil, fmt.Errorf("parse verify result failed: %w", err)
		}
		return &DriverLicenseVerifyResultFromLib{OK: false, Error: raw}, fmt.Errorf("verify license failed(code %d): %s", code, raw)
	}

	if code != 0 || !out.OK {
		msg := out.Error
		if msg == "" {
			msg = fmt.Sprintf("verify license failed(code %d)", code)
		}
		return out, fmt.Errorf("%s", msg)
	}

	return out, nil
}
