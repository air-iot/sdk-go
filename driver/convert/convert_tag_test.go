package convert

import (
	"encoding/json"
	"github.com/air-iot/sdk-go/driver/entity"
	"github.com/shopspring/decimal"
	"testing"
)

func TestConvertRange1(t *testing.T) {
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10,"defaultCondition":true}],"active":"boundary","fixedValue":10,"invalidAction":"save"}`
	var tagRange entity.Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(120)
	gotNewValue, gotRawValue, _, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
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
	var tagRange entity.Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(120)
	gotNewValue, gotRawValue, _, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
	t.Logf("new %+v , old %+v ,%t", gotNewValue, gotRawValue, gotIsSave)
	if gotNewValue != nil {
		t.Log(*gotNewValue)
	}
	if gotRawValue != nil {
		t.Log(*gotRawValue)
	}
}

func Test_ConvertRange3(t *testing.T) {
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10,"defaultCondition":true}],"active":"fixed","fixedValue":10,"invalidAction":"save"}`
	var tagRange entity.Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(120)
	gotNewValue, gotRawValue, _, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
	t.Logf("new %+v , old %+v ,%t", gotNewValue, gotRawValue, gotIsSave)
	if gotNewValue != nil {
		t.Log(*gotNewValue)
	}
	if gotRawValue != nil {
		t.Log(*gotRawValue)
	}
}

func Test_ConvertRange4(t *testing.T) {
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":80,"value":30,"defaultCondition":true}],"active":"fixed","fixedValue":10,"invalidAction":"save"}`
	var tagRange entity.Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(120)
	gotNewValue, gotRawValue, _, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
	t.Logf("new %+v , old %+v ,%t", gotNewValue, gotRawValue, gotIsSave)
	if gotNewValue != nil {
		t.Log(*gotNewValue)
	}
	if gotRawValue != nil {
		t.Log(*gotRawValue)
	}
}

func Test_ConvertRange5(t *testing.T) {
	rangeStr := `{"active":"boundary","maxValue":9999,"minValue":0}`
	var tagRange entity.Range

	if err := json.Unmarshal([]byte(rangeStr), &tagRange); err != nil {
		t.Fatal(err)
	}
	preVal := decimal.NewFromInt(8)
	raw := decimal.NewFromFloat(12000)
	gotNewValue, gotRawValue, _, gotIsSave := ConvertRange(&tagRange, &preVal, &raw)
	t.Logf("new %+v , old %+v ,%t", gotNewValue, gotRawValue, gotIsSave)
	if gotNewValue != nil {
		t.Log(*gotNewValue)
	}
	if gotRawValue != nil {
		t.Log(*gotRawValue)
	}
}
