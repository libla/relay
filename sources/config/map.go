package config

import (
	"math"
	"relay"
	"time"
)

func FromMap(root map[string]any) relay.Config {
	return &configMap{root: root}
}

type configMap struct {
	root map[string]any
	keys []string
}

type keys struct {
	keys  []string
	index int
}

func (this *keys) Next() *string {
	if this.index == len(this.keys) {
		return nil
	}
	result := this.keys[this.index]
	this.index++
	return &result
}

func (this configMap) Empty() bool {
	return len(this.root) == 0
}

func (this *configMap) Keys() interface{ Next() *string } {
	if this.keys == nil {
		keys := make([]string, 0, len(this.root))
		for k := range this.root {
			keys = append(keys, k)
		}
		this.keys = keys
	}
	return &keys{keys: this.keys}
}

func (this configMap) GetBool(key string) *bool {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toBool(value)
}

func (this configMap) GetInt(key string) *int64 {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toInt(value)
}

func (this configMap) GetUInt(key string) *uint64 {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toUInt(value)
}

func (this configMap) GetFloat(key string) *float64 {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toFloat(value)
}

func (this configMap) GetString(key string) *string {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toString(value)
}

func (this configMap) GetTime(key string) *time.Time {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toTime(value)
}

func (this configMap) GetDuration(key string) *time.Duration {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toDuration(value)
}

func (this configMap) GetConfig(key string) relay.Config {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toConfig(value)
}

func (this configMap) GetArray(key string) relay.Array {
	value, ok := this.root[key]
	if !ok {
		return nil
	}
	return toArray(value)
}

type configArray struct {
	root []any
}

func (this configArray) Count() int {
	return len(this.root)
}

func (this configArray) GetBool(index int) *bool {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toBool(this.root[index])
}

func (this configArray) GetInt(index int) *int64 {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toInt(this.root[index])
}

func (this configArray) GetUInt(index int) *uint64 {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toUInt(this.root[index])
}

func (this configArray) GetFloat(index int) *float64 {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toFloat(this.root[index])
}

func (this configArray) GetString(index int) *string {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toString(this.root[index])
}

func (this configArray) GetTime(index int) *time.Time {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toTime(this.root[index])
}

func (this configArray) GetDuration(index int) *time.Duration {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toDuration(this.root[index])
}

func (this configArray) GetConfig(index int) relay.Config {
	if index < 0 || index >= len(this.root) {
		return nil
	}
	return toConfig(this.root[index])
}

func toBool(value any) *bool {
	result, ok := value.(bool)
	if !ok {
		return nil
	}
	return &result
}

func toInt(value any) *int64 {
	var result int64
	switch value.(type) {
	case int:
		result = int64(value.(int))
	case uint:
		result = int64(value.(uint))
	case int8:
		result = int64(value.(int8))
	case uint8:
		result = int64(value.(uint8))
	case int16:
		result = int64(value.(int16))
	case uint16:
		result = int64(value.(uint16))
	case int32:
		result = int64(value.(int32))
	case uint32:
		result = int64(value.(uint32))
	case int64:
		result = value.(int64)
	case uint64:
		uint := value.(uint64)
		if uint > math.MaxInt64 {
			return nil
		}
		result = int64(uint)
	case float32:
		float := value.(float32)
		if float > math.MaxInt64 || float < math.MinInt64 {
			return nil
		}
		if math.IsInf(float64(float), 0) || math.IsNaN(float64(float)) {
			return nil
		}
		result = int64(float)
	case float64:
		float := value.(float64)
		if float > math.MaxInt64 || float < math.MinInt64 {
			return nil
		}
		if math.IsInf(float, 0) || math.IsNaN(float) {
			return nil
		}
		result = int64(float)
	default:
		return nil
	}
	return &result
}

func toUInt(value any) *uint64 {
	var result uint64
	switch value.(type) {
	case int:
		int := value.(int)
		if int < 0 {
			return nil
		}
		result = uint64(int)
	case uint:
		result = uint64(value.(uint))
	case int8:
		int := value.(int8)
		if int < 0 {
			return nil
		}
		result = uint64(int)
	case uint8:
		result = uint64(value.(uint8))
	case int16:
		int := value.(int16)
		if int < 0 {
			return nil
		}
		result = uint64(int)
	case uint16:
		result = uint64(value.(uint16))
	case int32:
		int := value.(int32)
		if int < 0 {
			return nil
		}
		result = uint64(int)
	case uint32:
		result = uint64(value.(uint32))
	case int64:
		int := value.(int64)
		if int < 0 {
			return nil
		}
		result = uint64(int)
	case uint64:
		result = value.(uint64)
	case float32:
		float := value.(float32)
		if float > math.MaxUint64 || float < 0 {
			return nil
		}
		if math.IsInf(float64(float), 0) || math.IsNaN(float64(float)) {
			return nil
		}
		result = uint64(float)
	case float64:
		float := value.(float64)
		if float > math.MaxUint64 || float < 0 {
			return nil
		}
		if math.IsInf(float, 0) || math.IsNaN(float) {
			return nil
		}
		result = uint64(float)
	default:
		return nil
	}
	return &result
}

func toFloat(value any) *float64 {
	var result float64
	switch value.(type) {
	case int:
		result = float64(value.(int))
	case uint:
		result = float64(value.(uint))
	case int8:
		result = float64(value.(int8))
	case uint8:
		result = float64(value.(uint8))
	case int16:
		result = float64(value.(int16))
	case uint16:
		result = float64(value.(uint16))
	case int32:
		result = float64(value.(int32))
	case uint32:
		result = float64(value.(uint32))
	case int64:
		result = float64(value.(int64))
	case uint64:
		result = float64(value.(uint64))
	case float32:
		result = float64(value.(float32))
	case float64:
		result = value.(float64)
	default:
		return nil
	}
	return &result
}

func toString(value any) *string {
	result, ok := value.(string)
	if !ok {
		return nil
	}
	return &result
}

func toTime(value any) *time.Time {
	result, ok := value.(time.Time)
	if !ok {
		return nil
	}
	if year, month, day := result.Date(); year == 0 && month == 1 && day == 1 {
		return nil
	}
	return &result
}

func toDuration(value any) *time.Duration {
	result, ok := value.(time.Time)
	if !ok {
		return nil
	}
	if year, month, day := result.Date(); year == 0 && month == 1 && day == 1 {
		duration := result.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, result.Location()))
		return &duration
	}
	return nil
}

func toConfig(value any) relay.Config {
	result, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	return &configMap{root: result}
}

func toArray(value any) relay.Array {
	result, ok := value.([]any)
	if !ok {
		return nil
	}
	return configArray{root: result}
}
