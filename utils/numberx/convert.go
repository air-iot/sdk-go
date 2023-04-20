package numberx

import (
	"fmt"
	"reflect"
	"strconv"
)

type FieldType string

const (
	String  FieldType = "string"
	Float   FieldType = "float"
	Int     FieldType = "integer"
	Bool    FieldType = "boolean"
	UNKNOWN FieldType = "UNKNOWN"
)

func (v FieldType) String() string {
	switch v {
	case String:
		return "string"
	case Float:
		return "float"
	case Int:
		return "integer"
	case Bool:
		return "boolean"
	default:
		return "UNKNOWN"
	}
}

func GetValueByType(valueType FieldType, v interface{}) (interface{}, error) {
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		v = reflect.ValueOf(v).Elem().Interface()
	}
	switch valueType {
	case String:
		val, err := GetString(v)
		if err != nil {
			return nil, err
		}
		return val, nil
	case Float:
		val, err := GetFloat(v)
		if err != nil {
			return nil, err
		}
		return val, nil
	case Int:
		val, err := GetInt(v)
		if err != nil {
			return nil, err
		}
		return val, nil
	case Bool:
		val, err := GetBool(v)
		if err != nil {
			return nil, err
		}
		return val, nil
	default:
		switch r := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
			return GetFloat(r)
		case string:
			return r, nil
		//case bool:
		//	if r {
		//		return 1, nil
		//	} else {
		//		return 0, nil
		//	}
		default:
			return nil, fmt.Errorf("数据类型非int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string, bool")
		}
	}
}

func GetString(v interface{}) (string, error) {
	switch r := v.(type) {
	case int:
		return strconv.Itoa(r), nil
	case int8:
		return strconv.FormatInt(int64(r), 10), nil
	case int16:
		return strconv.FormatInt(int64(r), 10), nil
	case int32:
		return strconv.FormatInt(int64(r), 10), nil
	case int64:
		return strconv.FormatInt(r, 10), nil
	case uint:
		return strconv.FormatUint(uint64(r), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(r), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(r), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(r), 10), nil
	case uint64:
		return strconv.FormatUint(r, 10), nil
	case float32:
		return strconv.FormatFloat(float64(r), 'f', -1, 64), nil
	case float64:
		return strconv.FormatFloat(r, 'f', -1, 32), nil
	case string:
		return r, nil
	case bool:
		if r {
			return "1", nil
		} else {
			return "0", nil
		}
	}
	return "", fmt.Errorf("不能转字符串,数据类型未知或错误")
}

func GetFloat(v interface{}) (float64, error) {
	switch r := v.(type) {
	case int:
		return float64(r), nil
	case int8:
		return float64(r), nil
	case int16:
		return float64(r), nil
	case int32:
		return float64(r), nil
	case int64:
		return float64(r), nil
	case uint:
		return float64(r), nil
	case uint8:
		return float64(r), nil
	case uint16:
		return float64(r), nil
	case uint32:
		return float64(r), nil
	case uint64:
		return float64(r), nil
	case float32:
		return float64(r), nil
	case float64:
		return r, nil
	case string:
		s, err := strconv.ParseFloat(r, 64)
		if err != nil {
			return 0, fmt.Errorf("转换值 %s 到 %s 错误,%s", v, Float, err)
		}
		return s, nil
	case bool:
		if r {
			return float64(1), nil
		} else {
			return float64(0), nil
		}
	}
	return 0, fmt.Errorf("不能转浮点数,数据类型未知或错误")
}

func GetInt(v interface{}) (int, error) {
	switch r := v.(type) {
	case int:
		return r, nil
	case int8:
		return int(r), nil
	case int16:
		return int(r), nil
	case int32:
		return int(r), nil
	case int64:
		return int(r), nil
	case uint:
		return int(r), nil
	case uint8:
		return int(r), nil
	case uint16:
		return int(r), nil
	case uint32:
		return int(r), nil
	case uint64:
		return int(r), nil
	case float32:
		return int(r), nil
	case float64:
		return int(r), nil
	case string:
		s, err := strconv.Atoi(r)
		if err != nil {
			return 0, fmt.Errorf("转换值 %s 到 %s 错误,%s", v, Int, err)
		}
		return s, nil
	case bool:
		if r {
			return 1, nil
		} else {
			return 0, nil
		}
	}
	return 0, fmt.Errorf("不能转整型,数据类型未知或错误")
}

func GetBool(v interface{}) (int, error) {
	switch r := v.(type) {
	case int:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case int8:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case int16:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case int32:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case int64:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case uint:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case uint8:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case uint16:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case uint32:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case uint64:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case float32:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case float64:
		if r != 0 {
			return 1, nil
		} else {
			return 0, nil
		}
	case string:
		if r == "true" || r == "false" {
			b, err := strconv.ParseBool(r)
			if err != nil {
				return 0, fmt.Errorf("转换值 %s 到 %s 错误,%s", v, Bool, err)
			}
			if b {
				return 1, nil
			} else {
				return 0, nil
			}

		} else {
			s, err := strconv.ParseFloat(r, 64)
			if err != nil {
				return 0, fmt.Errorf("转换值 %s 解析浮点到 %s 错误,%s", v, Bool, err)
			}
			if s != 0 {
				return 1, nil
			} else {
				return 0, nil
			}
		}
	case bool:
		if r {
			return 1, nil
		} else {
			return 0, nil
		}
	}
	return 0, fmt.Errorf("不能转布尔类型,数据类型未知或错误")
}
