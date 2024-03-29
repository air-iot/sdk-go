package convert

import (
	"github.com/air-iot/sdk-go/v4/driver/entity"
	"github.com/shopspring/decimal"
)

func Value(tagTemp *entity.Tag, raw decimal.Decimal) (val decimal.Decimal) {
	var value = raw
	if tagTemp.TagValue != nil {
		if tagTemp.TagValue.MinRaw != nil {
			minRaw := decimal.NewFromFloat(*tagTemp.TagValue.MinRaw)
			if value.LessThan(minRaw) {
				value = minRaw
			}
		}

		if tagTemp.TagValue.MaxRaw != nil {
			maxRaw := decimal.NewFromFloat(*tagTemp.TagValue.MaxRaw)
			if value.GreaterThan(maxRaw) {
				value = maxRaw
			}
		}

		if tagTemp.TagValue.MinRaw != nil && tagTemp.TagValue.MaxRaw != nil && tagTemp.TagValue.MinValue != nil && tagTemp.TagValue.MaxValue != nil {
			//value = (((rawTmp - minRaw) / (maxRaw - minRaw)) * (maxValue - minValue)) + minValue
			minRaw := decimal.NewFromFloat(*tagTemp.TagValue.MinRaw)
			maxRaw := decimal.NewFromFloat(*tagTemp.TagValue.MaxRaw)
			minValue := decimal.NewFromFloat(*tagTemp.TagValue.MinValue)
			maxValue := decimal.NewFromFloat(*tagTemp.TagValue.MaxValue)
			if !maxRaw.Equal(minRaw) {
				value = raw.Sub(minRaw).Div(maxRaw.Sub(minRaw)).Mul(maxValue.Sub(minValue)).Add(minValue)
			}
		}
	}

	if tagTemp.Fixed != nil {
		value = value.Round(*tagTemp.Fixed)
	}

	if tagTemp.Mod != nil {
		value = value.Mul(decimal.NewFromFloat(*tagTemp.Mod))
	}

	return value
}

func Range(tagRange *entity.Range, preVal, raw *decimal.Decimal) (newValue, rawValue *float64, invalidType string, isSave bool) {
	if raw == nil {
		return
	}
	value, _ := raw.Float64()
	if tagRange == nil {
		newValue = &value
		return
	}
	if tagRange.MinValue == nil || tagRange.MaxValue == nil {
		switch tagRange.Method {
		case entity.RangeMethod_Valid:
			newValue, rawValue, isSave = conditions(tagRange, preVal, raw)
			return
		case entity.RangeMethod_Invalid:
			return invalidConditions(tagRange, preVal, raw)
		default:
			newValue = &value
			return
		}
	}
	minValue := decimal.NewFromFloat(*tagRange.MinValue)
	maxValue := decimal.NewFromFloat(*tagRange.MaxValue)
	if raw.GreaterThanOrEqual(minValue) && raw.LessThanOrEqual(maxValue) {
		newValue = &value
		return
	}
	switch tagRange.Active {
	case entity.Active_Fixed:
		if tagRange.FixedValue == nil {
			newValue = &value
			return
		}
		newValue = tagRange.FixedValue
		return
	case entity.Active_Boundary:
		if raw.LessThan(minValue) {
			newValue = tagRange.MinValue
			return
		}
		if raw.GreaterThan(maxValue) {
			newValue = tagRange.MaxValue
			return
		}
	case entity.Active_Discard:
		return
	case entity.Active_Latest:
		if preVal == nil {
			return
		}
		preValue, _ := preVal.Float64()
		newValue = &preValue
		return
	}
	newValue = &value
	return
}

func conditions(tagRange *entity.Range, preVal, raw *decimal.Decimal) (newValue, rawValue *float64, isSave bool) {
	if raw == nil {
		return
	}
	isSave = true
	value, _ := raw.Float64()
	if tagRange == nil {
		newValue = &value
		return
	}

	if tagRange.Conditions == nil || len(tagRange.Conditions) == 0 {
		newValue = &value
		return
	}
	var defaultCondition *entity.RangeCondition = nil
	for i, condition := range tagRange.Conditions {
		if condition.DefaultCondition {
			defaultCondition = &tagRange.Conditions[i]
		}
		var currentValue *decimal.Decimal = nil
		switch condition.Mode {
		case entity.ConditionMode_Number:
			currentValue = raw
		case entity.ConditionMode_Rate:
			if preVal == nil {
				continue
			}
			pf, _ := preVal.Float64()
			if pf == 0 {
				continue
			}
			rateValue := ((raw.Sub(*preVal)).Div(*preVal)).Mul(decimal.NewFromInt(100))
			currentValue = &rateValue
		case entity.ConditionMode_Delta:
			if preVal == nil {
				continue
			}
			deltaValue := raw.Sub(*preVal)
			currentValue = &deltaValue
		}
		if currentValue != nil {
			switch condition.Condition {
			case entity.Condition_Range:
				if condition.MinValue != nil && condition.MaxValue != nil {
					minValue := decimal.NewFromFloat(*condition.MinValue)
					maxValue := decimal.NewFromFloat(*condition.MaxValue)
					if currentValue.GreaterThanOrEqual(minValue) && currentValue.LessThanOrEqual(maxValue) {
						newValue = &value
						return
					}
				}
			case entity.Condition_Greater:
				if condition.Value != nil {
					valueTmp := decimal.NewFromFloat(*condition.Value)
					if currentValue.GreaterThan(valueTmp) {
						newValue = &value
						return
					}
				}
			case entity.Condition_Less:
				if condition.Value != nil {
					valueTmp := decimal.NewFromFloat(*condition.Value)
					if currentValue.LessThan(valueTmp) {
						newValue = &value
						return
					}
				}
			}
		}
	}
	switch tagRange.InvalidAction {
	case entity.InvalidAction_Save:
		rawValue = &value
	}
	switch tagRange.Active {
	case entity.Active_Fixed:
		if tagRange.FixedValue == nil {
			newValue = nil
			return
		}
		newValue = tagRange.FixedValue
		return
	case entity.Active_Boundary:
		if defaultCondition == nil {
			newValue = nil
			return
		}
		switch defaultCondition.Mode {
		case entity.ConditionMode_Number:
			switch defaultCondition.Condition {
			case entity.Condition_Range:
				if defaultCondition.MinValue != nil && defaultCondition.MaxValue != nil {
					minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
					if raw.LessThan(minValue) {
						newValue = defaultCondition.MinValue
						return
					}
					maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
					if raw.GreaterThan(maxValue) {
						newValue = defaultCondition.MaxValue
						return
					}
				} else {
					newValue = nil
					return
				}
			case entity.Condition_Greater, entity.Condition_Less:
				newValue = defaultCondition.Value
				return
			}
		case entity.ConditionMode_Rate:
			if preVal == nil {
				newValue = &value
				return
			}
			pf, _ := preVal.Float64()
			if pf == 0 {
				newValue = &value
				return
			}
			rateValue := (raw.Sub(*preVal)).Div(*preVal).Mul(decimal.NewFromInt(100))
			one := decimal.NewFromInt(1)
			switch defaultCondition.Condition {
			case entity.Condition_Range:
				// x = (min + 1) * pre
				if defaultCondition.MinValue != nil && defaultCondition.MaxValue != nil {
					minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
					if rateValue.LessThan(minValue) {
						sub, _ := (minValue.Div(decimal.NewFromInt(100)).Add(one)).Mul(*preVal).Float64()
						newValue = &sub
						return
					}
					maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
					if rateValue.GreaterThan(maxValue) {
						sub, _ := (maxValue.Div(decimal.NewFromInt(100)).Add(one)).Mul(*preVal).Float64()
						newValue = &sub
						return
					}
				}
			case entity.Condition_Greater, entity.Condition_Less:
				if defaultCondition.Value == nil {
					newValue = nil
					return
				}
				defaultValue := decimal.NewFromFloat(*defaultCondition.Value)
				sub, _ := (defaultValue.Div(decimal.NewFromInt(100)).Add(one)).Mul(*preVal).Float64()
				newValue = &sub
				return

			}
		case entity.ConditionMode_Delta:
			if preVal == nil {
				newValue = &value
				return
			}
			deltaValue := raw.Sub(*preVal)
			switch defaultCondition.Condition {
			case entity.Condition_Range:
				// x = (min +1)*pre
				if defaultCondition.MinValue != nil && defaultCondition.MaxValue != nil {
					minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
					if deltaValue.LessThan(minValue) {
						sub, _ := (minValue.Add(*preVal)).Float64()
						newValue = &sub
						return
					}
					maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
					if deltaValue.GreaterThan(maxValue) {
						sub, _ := (maxValue.Add(*preVal)).Float64()
						newValue = &sub
						return
					}
				}
			case entity.Condition_Greater, entity.Condition_Less:
				if defaultCondition.Value == nil {
					newValue = nil
					return
				}
				defaultValue := decimal.NewFromFloat(*defaultCondition.Value)
				sub, _ := (defaultValue.Add(*preVal)).Float64()
				newValue = &sub
				return
			}
		}
	case entity.Active_Discard:
		newValue = nil
		return
	case entity.Active_Latest:
		if preVal == nil {
			newValue = &value
			return
		}
		preValue, _ := preVal.Float64()
		newValue = &preValue
		return
	}
	return
}

func invalidConditions(tagRange *entity.Range, preVal, raw *decimal.Decimal) (newValue, invalidValue *float64, invalidType string, isSave bool) {
	if raw == nil {
		return
	}
	isSave = true
	value, _ := raw.Float64()
	if tagRange == nil {
		newValue = &value
		return
	}

	if tagRange.Conditions == nil || len(tagRange.Conditions) == 0 {
		newValue = &value
		return
	}
	//var isInvalid bool
	//var defaultCondition *entity.RangeCondition = nil
loop:
	for _, condition := range tagRange.Conditions {
		//if condition.DefaultCondition {
		//	defaultCondition = &tagRange.Conditions[i]
		//}
		var currentValue *decimal.Decimal = nil
		switch condition.Mode {
		case entity.ConditionMode_Number:
			currentValue = raw
		case entity.ConditionMode_Rate:
			if preVal == nil {
				continue
			}
			pf, _ := preVal.Float64()
			if pf == 0 {
				continue
			}
			rateValue := ((raw.Sub(*preVal)).Div(*preVal)).Mul(decimal.NewFromInt(100))
			currentValue = &rateValue
		case entity.ConditionMode_Delta:
			if preVal == nil {
				continue
			}
			deltaValue := raw.Sub(*preVal)
			currentValue = &deltaValue
		}
		if currentValue != nil {
			switch condition.Condition {
			case entity.Condition_Range:
				if condition.MinValue != nil && condition.MaxValue != nil {
					minValue := decimal.NewFromFloat(*condition.MinValue)
					maxValue := decimal.NewFromFloat(*condition.MaxValue)
					if currentValue.GreaterThanOrEqual(minValue) && currentValue.LessThanOrEqual(maxValue) {
						invalidValue = &value
						invalidType = condition.InvalidType
						break loop
					}
				}
			case entity.Condition_Greater:
				if condition.Value != nil {
					valueTmp := decimal.NewFromFloat(*condition.Value)
					if currentValue.GreaterThan(valueTmp) {
						invalidValue = &value
						invalidType = condition.InvalidType
						break loop
					}
				}
			case entity.Condition_Less:
				if condition.Value != nil {
					valueTmp := decimal.NewFromFloat(*condition.Value)
					if currentValue.LessThan(valueTmp) {
						invalidValue = &value
						invalidType = condition.InvalidType
						break loop
					}
				}
			}
		}
	}
	if invalidValue != nil {
		switch tagRange.InvalidAction {
		case entity.InvalidAction_Save:
		//invalidValue = &value
		default:
			invalidValue = nil
		}
		switch tagRange.Active {
		case entity.Active_Fixed:
			if tagRange.FixedValue == nil {
				newValue = nil
				return
			}
			newValue = tagRange.FixedValue
			return
		//case entity.Active_Boundary:
		//	if defaultCondition == nil {
		//		newValue = nil
		//		return
		//	}
		//	switch defaultCondition.Mode {
		//	case entity.ConditionMode_Number:
		//		switch defaultCondition.Condition {
		//		case entity.Condition_Range:
		//			if defaultCondition.MinValue != nil && defaultCondition.MaxValue != nil {
		//				minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
		//				if raw.LessThan(minValue) {
		//					newValue = defaultCondition.MinValue
		//					return
		//				}
		//				maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
		//				if raw.GreaterThan(maxValue) {
		//					newValue = defaultCondition.MaxValue
		//					return
		//				}
		//			} else {
		//				newValue = nil
		//				return
		//			}
		//		case entity.Condition_Greater, entity.Condition_Less:
		//			newValue = defaultCondition.Value
		//			return
		//		}
		//	case entity.ConditionMode_Rate:
		//		if preVal == nil {
		//			newValue = &value
		//			return
		//		}
		//		pf, _ := preVal.Float64()
		//		if pf == 0 {
		//			newValue = &value
		//			return
		//		}
		//		rateValue := (raw.Sub(*preVal)).Div(*preVal).Mul(decimal.NewFromInt(100))
		//		one := decimal.NewFromInt(1)
		//		switch defaultCondition.Condition {
		//		case entity.Condition_Range:
		//			// x = (min + 1) * pre
		//			if defaultCondition.MinValue != nil && defaultCondition.MaxValue != nil {
		//				minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
		//				if rateValue.LessThan(minValue) {
		//					sub, _ := (minValue.Div(decimal.NewFromInt(100)).Add(one)).Mul(*preVal).Float64()
		//					newValue = &sub
		//					return
		//				}
		//				maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
		//				if rateValue.GreaterThan(maxValue) {
		//					sub, _ := (maxValue.Div(decimal.NewFromInt(100)).Add(one)).Mul(*preVal).Float64()
		//					newValue = &sub
		//					return
		//				}
		//			}
		//		case entity.Condition_Greater, entity.Condition_Less:
		//			if defaultCondition.Value == nil {
		//				newValue = nil
		//				return
		//			}
		//			defaultValue := decimal.NewFromFloat(*defaultCondition.Value)
		//			sub, _ := (defaultValue.Div(decimal.NewFromInt(100)).Add(one)).Mul(*preVal).Float64()
		//			newValue = &sub
		//			return
		//
		//		}
		//	case entity.ConditionMode_Delta:
		//		if preVal == nil {
		//			newValue = &value
		//			return
		//		}
		//		deltaValue := raw.Sub(*preVal)
		//		switch defaultCondition.Condition {
		//		case entity.Condition_Range:
		//			// x = (min +1)*pre
		//			if defaultCondition.MinValue != nil && defaultCondition.MaxValue != nil {
		//				minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
		//				if deltaValue.LessThan(minValue) {
		//					sub, _ := (minValue.Add(*preVal)).Float64()
		//					newValue = &sub
		//					return
		//				}
		//				maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
		//				if deltaValue.GreaterThan(maxValue) {
		//					sub, _ := (maxValue.Add(*preVal)).Float64()
		//					newValue = &sub
		//					return
		//				}
		//			}
		//		case entity.Condition_Greater, entity.Condition_Less:
		//			if defaultCondition.Value == nil {
		//				newValue = nil
		//				return
		//			}
		//			defaultValue := decimal.NewFromFloat(*defaultCondition.Value)
		//			sub, _ := (defaultValue.Add(*preVal)).Float64()
		//			newValue = &sub
		//			return
		//		}
		//	}
		case entity.Active_Discard:
			newValue = nil
			return
		case entity.Active_Latest:
			if preVal == nil {
				newValue = &value
				return
			}
			preValue, _ := preVal.Float64()
			newValue = &preValue
			return
		}
	} else {
		newValue = &value
	}

	return
}
