package convert

import (
	"github.com/air-iot/json"
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"github.com/shopspring/decimal"
	"testing"
)

func Test_ConvertRange_1(t *testing.T) {
	type args struct {
		tagRange *entity.Range
		preVal   *decimal.Decimal
		raw      *decimal.Decimal
	}
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10,"defaultCondition":true}],"active":"boundary","fixedValue":10,"invalidAction":"save"}`
	var tagRange entity.Range

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

func Test_ConvertRange_2(t *testing.T) {
	type args struct {
		tagRange *entity.Range
		preVal   *decimal.Decimal
		raw      *decimal.Decimal
	}
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10},{"mode":"rate","condition":"range","minValue":0,"maxValue":1000,"value":10,"defaultCondition":true}],"active":"boundary","fixedValue":10,"invalidAction":"save"}`
	var tagRange entity.Range

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

func Test_ConvertRange_3(t *testing.T) {
	type args struct {
		tagRange *entity.Range
		preVal   *decimal.Decimal
		raw      *decimal.Decimal
	}
	rangeStr := `{"conditions":[{"mode":"number","condition":"range","minValue":0,"maxValue":10,"value":10,"defaultCondition":true},{"mode":"rate","condition":"range","minValue":0,"maxValue":1000,"value":10,"defaultCondition":false}],"active":"boundary","fixedValue":10,"invalidAction":"save"}`
	var tagRange entity.Range

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
