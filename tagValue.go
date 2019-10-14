package sdk

import (
	"fmt"
	"reflect"
	"strconv"
)

func ConvertValue(tagValue TagValue, fixedIntf interface{}, modIntf interface{}, value interface{}) interface{} {

	if modIntf == nil {
		modIntf = "0"
	}
	if fixedIntf == nil {
		fixedIntf = "0"
	}
	modStr, ok := modIntf.(string)
	if !ok {
		fmt.Println("mod转string失败", modIntf)
	}
	mod, err := strconv.ParseFloat(modStr, 64)
	if err != nil {
		fmt.Println("mod转float失败")
	}

	fixedStr, ok := fixedIntf.(string)
	if !ok {
		fmt.Println("fix转string失败", fixedIntf)
	}
	fixed, err := strconv.Atoi(fixedStr)
	if err != nil {
		fmt.Println("fix转int失败")
	}

	vType := reflect.TypeOf(value).String()
	if vType == "uint" || vType == "uintptr" || vType == "uint8" || vType == "uint16" || vType == "uint32" || vType == "uint64" {
		vDetail := reflect.ValueOf(value).Uint()
		if tagValue.MinValue != tagValue.MaxValue {
			if uint64(tagValue.MinValue) > vDetail {
				vDetail = uint64(tagValue.MinValue)
			}
			if uint64(tagValue.MaxValue) < vDetail {
				vDetail = uint64(tagValue.MaxValue)
			}
			if tagValue.MinRaw != tagValue.MaxRaw {
				vDetail = ((vDetail-uint64(tagValue.MinRaw))/(uint64(tagValue.MaxRaw)-uint64(tagValue.MinRaw)))*(uint64(tagValue.MaxValue)-uint64(tagValue.MinValue)) + uint64(tagValue.MinValue)
			}
		}
		if mod != 0 {
			vDetail = vDetail * uint64(mod)
		}
		return vDetail
	} else if vType == "int" || vType == "int8" || vType == "int16" || vType == "int32" || vType == "int64" {
		vDetail := reflect.ValueOf(value).Int()
		if tagValue.MinValue != tagValue.MaxValue {
			if int64(tagValue.MinValue) > vDetail {
				vDetail = int64(tagValue.MinValue)
			}
			if int64(tagValue.MaxValue) < vDetail {
				vDetail = int64(tagValue.MaxValue)
			}
			if tagValue.MinRaw != tagValue.MaxRaw {
				vDetail = ((vDetail-int64(tagValue.MinRaw))/(int64(tagValue.MaxRaw)-int64(tagValue.MinRaw)))*(int64(tagValue.MaxValue)-int64(tagValue.MinValue)) + int64(tagValue.MinValue)
			}
		}
		if mod != 0 {
			vDetail = vDetail * int64(mod)
		}
		return vDetail
	} else if vType == "float32" || vType == "float64" {
		vDetail := reflect.ValueOf(value).Float()
		if tagValue.MinValue != tagValue.MaxValue {
			if float64(tagValue.MinValue) > vDetail {
				vDetail = float64(tagValue.MinValue)
			}
			if float64(tagValue.MaxValue) < vDetail {
				vDetail = float64(tagValue.MaxValue)
			}
			if tagValue.MinRaw != tagValue.MaxRaw {
				vDetail = ((vDetail-float64(tagValue.MinRaw))/(float64(tagValue.MaxRaw)-float64(tagValue.MinRaw)))*(float64(tagValue.MaxValue)-float64(tagValue.MinValue)) + float64(tagValue.MinValue)
			}
		}
		if mod != 0 {
			vDetail = vDetail * float64(mod)
		}
		if fixed != 0 {
			rs, err := strconv.ParseFloat(fmt.Sprintf("%.2f", vDetail), 64)
			if err != nil {
				return value
			}
			vDetail = rs
		}
		return vDetail
	} else {
		return value
	}
}
