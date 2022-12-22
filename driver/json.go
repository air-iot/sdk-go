package driver

import (
	"encoding/json"
	"fmt"

	jsonpatch "gopkg.in/evanphx/json-patch.v5"
)

func Merge(original, target, result interface{}) error {
	originalBts, err := json.Marshal(original)
	if err != nil {
		return fmt.Errorf("origin err,%s", err)
	}
	targetBts, err := json.Marshal(target)
	if err != nil {
		return fmt.Errorf("target err,%s", err)
	}
	data, err := jsonpatch.MergePatch(originalBts, targetBts)
	if err != nil {
		return fmt.Errorf("merge err,%s", err)
	}
	if err := json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("unmarshal err,%s", err)
	}
	return nil
}
