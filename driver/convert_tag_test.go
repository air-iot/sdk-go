package driver

import (
	"encoding/json"
	"github.com/shopspring/decimal"
	"testing"
)

func TestConvertRange1(t *testing.T) {
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10,"defaultCondition":true}],"active":"boundary","fixedValue":10,"invalidAction":"save"}`
	var tagRange Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(120)
	gotNewValue, gotRawValue, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
	t.Logf("new %+v , old %+v ,%t", gotNewValue, gotRawValue, gotIsSave)
	if gotNewValue != nil {
		t.Log(*gotNewValue)
	}
	if gotRawValue != nil {
		t.Log(*gotRawValue)
	}
}

func TestConvertRange2(t *testing.T) {
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10,"defaultCondition":true}],"active":"boundary","fixedValue":10,"invalidAction":"save"}`
	var tagRange Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(120)
	gotNewValue, gotRawValue, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
	t.Logf("new %+v , old %+v ,%t", gotNewValue, gotRawValue, gotIsSave)
	if gotNewValue != nil {
		t.Log(*gotNewValue)
	}
	if gotRawValue != nil {
		t.Log(*gotRawValue)
	}
}
