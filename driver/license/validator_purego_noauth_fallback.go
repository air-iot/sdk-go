package license

import (
	"encoding/json"
	"fmt"
	"strings"
)

const freeTagLimitNoAuthFromLib = 20

func shouldTreatPathAsNoLicense(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "license path not found") ||
		strings.Contains(msg, "license path must be a directory") ||
		strings.Contains(msg, "license path is empty")
}

func fallbackVerifyAsNoLicense(data string) (bool, *DriverLicenseInfoFromLib, error) {
	tagCount, err := countUniqueTagIDsFromDriverData(data)
	if err != nil {
		return false, nil, err
	}
	if tagCount > freeTagLimitNoAuthFromLib {
		return false, nil, fmt.Errorf("license unavailable, tag count %d exceeds limit %d", tagCount, freeTagLimitNoAuthFromLib)
	}
	return true, nil, nil
}

func countUniqueTagIDsFromDriverData(dataJSON string) (int, error) {
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &root); err != nil {
		return 0, fmt.Errorf("invalid data json: %w", err)
	}

	rootDevice, _ := root["device"].(map[string]interface{})
	rootTags := extractTagIDsFromAny(rootDevice["tags"])

	tables, ok := root["tables"].([]interface{})
	if !ok {
		return 0, nil
	}

	total := 0
	for _, tableRaw := range tables {
		table, ok := tableRaw.(map[string]interface{})
		if !ok {
			continue
		}

		tableTags := cloneTagSet(rootTags)
		addTagSet(tableTags, extractTagIDsFromAny(table["tags"]))
		tableDevice, _ := table["device"].(map[string]interface{})
		addTagSet(tableTags, extractTagIDsFromAny(tableDevice["tags"]))

		devices, ok := table["devices"].([]interface{})
		if !ok {
			continue
		}
		for _, devRaw := range devices {
			dev, ok := devRaw.(map[string]interface{})
			if !ok {
				continue
			}

			deviceTags := cloneTagSet(tableTags)
			addTagSet(deviceTags, extractTagIDsFromAny(dev["tags"]))
			nestedDevice, _ := dev["device"].(map[string]interface{})
			addTagSet(deviceTags, extractTagIDsFromAny(nestedDevice["tags"]))
			total += len(deviceTags)
		}
	}

	return total, nil
}

func extractTagIDsFromAny(v interface{}) map[string]struct{} {
	out := make(map[string]struct{})
	arr, ok := v.([]interface{})
	if !ok {
		return out
	}
	for _, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id := strings.TrimSpace(fmt.Sprint(m["id"]))
		if id == "" || id == "<nil>" {
			continue
		}
		out[id] = struct{}{}
	}
	return out
}

func cloneTagSet(src map[string]struct{}) map[string]struct{} {
	dst := make(map[string]struct{}, len(src))
	for k := range src {
		dst[k] = struct{}{}
	}
	return dst
}

func addTagSet(dst, src map[string]struct{}) {
	for k := range src {
		dst[k] = struct{}{}
	}
}
