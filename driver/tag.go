package driver

type Tag struct {
	ID   string `json:"id" description:"ID"`
	Name string `json:"name" description:"自定义名称"`
	//以下为通用值计算相关属性
	TagValue *TagValue `json:"tagValue"`
	Fixed    *int32    `json:"fixed"`
	Mod      *float64  `json:"mod"`
	Range    *Range    `json:"range"`
}

type TagValue struct {
	MinValue *float64 `json:"minValue"`
	MaxValue *float64 `json:"maxValue"`
	MinRaw   *float64 `json:"minRaw"`
	MaxRaw   *float64 `json:"maxRaw"`
}

type Range struct {
	MinValue   *float64 `json:"minValue"`
	MaxValue   *float64 `json:"maxValue"`
	Active     *string  `json:"active"`
	FixedValue *float64 `json:"fixedValue"`
}
