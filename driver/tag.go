package driver

type Tag struct {
	ID   string `json:"id" description:"ID"`
	Name string `json:"name" description:"自定义名称"`

	//以下为通用值计算相关属性
	TagValue *TagValue   `json:"tagValue"`
	Fixed    interface{} `json:"fixed"`
	Mod      interface{} `json:"mod"`
}

type TagValue struct {
	MinValue float64 `json:"minValue"`
	MaxValue float64 `json:"maxValue"`
	MinRaw   float64 `json:"minRaw"`
	MaxRaw   float64 `json:"maxRaw"`
}
