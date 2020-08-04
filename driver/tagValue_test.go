package driver

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestConvertValue(t *testing.T) {

	tv := TagValue{MinValue: 1.2, MaxValue: 10, MinRaw: 0, MaxRaw: 0}
	rs := ConvertValue(tv, 0, 0, 100)
	fmt.Println("结果", rs)

	tv = TagValue{MinValue: 1, MaxValue: 100, MinRaw: 0, MaxRaw: 5}
	rs = ConvertValue(tv, 0, 0, 50)
	fmt.Println("结果", rs)

}

func TestConvertValue2(t *testing.T) {
	tv := TagValue{MinValue: 0, MaxValue: 0, MinRaw: 0, MaxRaw: 1}
	rs := ConvertValue(tv, 2, 0, 50)
	fmt.Println("结果", rs)
}

func TestConvertValue3(t *testing.T) {
	tv := TagValue{MinValue: 1, MaxValue: 10000, MinRaw: 0, MaxRaw: 1000}
	rs := ConvertValue(tv, 2, 0, 50.22)
	fmt.Println("结果", rs)
}

func TestJson(t *testing.T) {
	type T struct {
		V1 float64 `json:"v_1"`
	}
	j := `{"v_1":12.1}`

	r := new(T)
	err := json.Unmarshal([]byte(j), r)
	if err != nil {
		fmt.Println("what s wrong with you", err)
	}
	fmt.Println(r)
}
