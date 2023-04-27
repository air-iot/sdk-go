package entity

import "github.com/shopspring/decimal"

func ConvertValue(tagTemp *Tag, raw decimal.Decimal) (val decimal.Decimal) {
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

func ConvertRange(tagRange *Range, preVal, raw *decimal.Decimal) (newValue, rawValue *float64, isSave bool) {
	if raw == nil {
		return nil, nil, false
	}
	value, _ := raw.Float64()
	if tagRange == nil {
		return &value, nil, false
	}
	if tagRange.MinValue == nil || tagRange.MaxValue == nil {
		return convertConditions(tagRange, preVal, raw)
	}
	minValue := decimal.NewFromFloat(*tagRange.MinValue)
	maxValue := decimal.NewFromFloat(*tagRange.MaxValue)
	if raw.GreaterThanOrEqual(minValue) && raw.LessThanOrEqual(maxValue) {
		return &value, nil, false
	}
	switch tagRange.Active {
	case Active_Fixed:
		if tagRange.FixedValue == nil {
			return &value, nil, false
		}
		return tagRange.FixedValue, nil, false
	case Active_Boundary:
		if raw.LessThan(minValue) {
			return tagRange.MinValue, nil, false
		}
		if raw.GreaterThan(maxValue) {
			return tagRange.MaxValue, nil, false
		}
	case Active_Discard:
		return nil, nil, false
	case Active_Latest:
		if preVal == nil {
			return nil, nil, false
		}
		preValue, _ := preVal.Float64()
		return &preValue, nil, false
	}
	return &value, nil, false
}

func convertConditions(tagRange *Range, preVal, raw *decimal.Decimal) (newValue, rawValue *float64, isSave bool) {
	if raw == nil {
		return
	}
	value, _ := raw.Float64()
	if tagRange == nil {
		newValue = &value
		return
	}
	switch tagRange.InvalidAction {
	case InvalidAction_Save:
		rawValue = &value
	}
	if tagRange.Conditions == nil || len(tagRange.Conditions) == 0 {
		newValue = &value
		return
	}
	var defaultCondition *RangeCondition = nil
	for i, condition := range tagRange.Conditions {
		if condition.DefaultCondition {
			defaultCondition = &tagRange.Conditions[i]
		}
		var currentValue *decimal.Decimal = nil
		switch condition.Mode {
		case ConditionMode_Number:
			currentValue = raw
		case ConditionMode_Rate:
			if !isSave {
				isSave = true
			}
			if preVal == nil {
				continue
			}
			rateValue := ((raw.Sub(*preVal)).Div(*preVal)).Mul(decimal.NewFromInt(100))
			currentValue = &rateValue
		case ConditionMode_Delta:
			if !isSave {
				isSave = true
			}
			if preVal == nil {
				continue
			}
			deltaValue := raw.Sub(*preVal)
			currentValue = &deltaValue
		}
		if currentValue != nil {
			switch condition.Condition {
			case Condition_Range:
				if condition.MinValue != nil && condition.MaxValue != nil {
					minValue := decimal.NewFromFloat(*condition.MinValue)
					maxValue := decimal.NewFromFloat(*condition.MaxValue)
					if currentValue.GreaterThanOrEqual(minValue) && currentValue.LessThanOrEqual(maxValue) {
						newValue = &value
						return
					}
				}
			case Condition_Greater:
				if condition.Value != nil {
					valueTmp := decimal.NewFromFloat(*condition.Value)
					if currentValue.GreaterThan(valueTmp) {
						newValue = &value
						return
					}
				}
			case Condition_Less:
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
	switch tagRange.Active {
	case Active_Fixed:
		if tagRange.FixedValue == nil {
			newValue = nil
			return
		}
		newValue = tagRange.FixedValue
		return
	case Active_Boundary:
		if defaultCondition == nil {
			newValue = nil
			return
		}
		switch defaultCondition.Mode {
		case ConditionMode_Number:
			switch defaultCondition.Condition {
			case Condition_Range:
				if defaultCondition.MinValue != nil {
					minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
					if raw.LessThan(minValue) {
						newValue = defaultCondition.MinValue
						return
					}
				}
				if defaultCondition.MaxValue != nil {
					maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
					if raw.GreaterThan(maxValue) {
						newValue = defaultCondition.MaxValue
						return
					}
				}
			case Condition_Greater, Condition_Less:
				newValue = defaultCondition.Value
				return
			}
		case ConditionMode_Rate:
			if preVal == nil {
				newValue = nil
				return
			}
			rateValue := (raw.Sub(*preVal)).Div(*preVal).Mul(decimal.NewFromInt(100))
			one := decimal.NewFromInt(1)
			switch defaultCondition.Condition {
			case Condition_Range:
				// x = (min + 1) * pre
				if defaultCondition.MinValue != nil {
					minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
					if rateValue.LessThan(minValue) {
						sub, _ := (minValue.Add(one)).Mul(*preVal).Float64()
						newValue = &sub
						return
					}
				}
				if defaultCondition.MaxValue != nil {
					maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
					if rateValue.GreaterThan(maxValue) {
						sub, _ := (maxValue.Add(one)).Mul(*preVal).Float64()
						newValue = &sub
						return
					}
				}
			case Condition_Greater, Condition_Less:
				if defaultCondition.Value == nil {
					newValue = defaultCondition.Value
					return
				}
				defaultValue := decimal.NewFromFloat(*defaultCondition.Value)
				sub, _ := (defaultValue.Add(one)).Mul(*preVal).Float64()
				newValue = &sub
				return

			}
		case ConditionMode_Delta:
			if preVal == nil {
				newValue = nil
				return
			}
			deltaValue := raw.Sub(*preVal)
			switch defaultCondition.Condition {
			case Condition_Range:
				// x = (min +1)*pre
				if defaultCondition.MinValue != nil {
					minValue := decimal.NewFromFloat(*defaultCondition.MinValue)
					if deltaValue.LessThan(minValue) {
						sub, _ := (minValue.Add(*preVal)).Float64()
						newValue = &sub
						return
					}
				}
				if defaultCondition.MaxValue != nil {
					maxValue := decimal.NewFromFloat(*defaultCondition.MaxValue)
					if deltaValue.GreaterThan(maxValue) {
						sub, _ := (maxValue.Add(*preVal)).Float64()
						newValue = &sub
						return
					}
				}
			case Condition_Greater, Condition_Less:
				if defaultCondition.Value == nil {
					newValue = defaultCondition.Value
					return
				}
				defaultValue := decimal.NewFromFloat(*defaultCondition.Value)
				sub, _ := (defaultValue.Add(*preVal)).Float64()
				newValue = &sub
				return
			}
		}
	case Active_Discard:
		newValue = nil
		return
	case Active_Latest:
		if preVal == nil {
			newValue = nil
			return
		}
		preValue, _ := preVal.Float64()
		newValue = &preValue
		return
	}
	return
}
