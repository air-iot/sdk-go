package entity

type Tag struct {
	ID   string `json:"id" description:"ID"`
	Name string `json:"name" description:"自定义名称"`

	//以下为通用值计算相关属性
	TagValue *TagValue `json:"tagValue"`
	Fixed    *int32    `json:"fixed"`
	Mod      *float64  `json:"mod"`
	Range    *Range    `json:"range"`
}

type Active string

const (
	Active_Fixed    Active = "fixed"
	Active_Boundary Active = "boundary"
	Active_Discard  Active = "discard"
	Active_Latest   Active = "latest"
)

type InvalidAction string

const (
	InvalidAction_Save InvalidAction = "save"
)

type Range struct {
	MinValue      *float64         `json:"minValue"`
	MaxValue      *float64         `json:"maxValue"`
	Conditions    []RangeCondition `json:"conditions"`
	Active        Active           `json:"active"`
	FixedValue    *float64         `json:"fixedValue"`
	InvalidAction InvalidAction    `json:"invalidAction"`
}

type ConditionMode string

const (
	ConditionMode_Number ConditionMode = "number"
	ConditionMode_Rate   ConditionMode = "rate"
	ConditionMode_Delta  ConditionMode = "delta"
)

type Condition string

const (
	Condition_Range   Condition = "range"
	Condition_Greater Condition = "greater"
	Condition_Less    Condition = "less"
)

type RangeCondition struct {
	Mode             ConditionMode `json:"mode"`
	Condition        Condition     `json:"condition"`
	MinValue         *float64      `json:"minValue"`
	MaxValue         *float64      `json:"maxValue"`
	Value            *float64      `json:"value"`
	DefaultCondition bool          `json:"defaultCondition"`
}

type TagValue struct {
	MinValue *float64 `json:"minValue"`
	MaxValue *float64 `json:"maxValue"`
	MinRaw   *float64 `json:"minRaw"`
	MaxRaw   *float64 `json:"maxRaw"`
}

//type Range struct {
//	MinValue   *float64 `json:"minValue"`
//	MaxValue   *float64 `json:"maxValue"`
//	Active     *string  `json:"active"`
//	FixedValue *float64 `json:"fixedValue"`
//}
