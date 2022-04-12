package relay

import (
	"math"
	"sort"
	"time"
)

type Configurable interface {
	LoadConfig(config Config) bool
}

type Config interface {
	Empty() bool
	Keys() interface{ Next() *string }
	GetBool(key string) *bool
	GetInt(key string) *int64
	GetUInt(key string) *uint64
	GetFloat(key string) *float64
	GetString(key string) *string
	GetTime(key string) *time.Time
	GetDuration(key string) *time.Duration
	GetConfig(key string) Config
	GetArray(key string) Array
}

type Array interface {
	Count() int
	GetBool(index int) *bool
	GetInt(index int) *int64
	GetUInt(index int) *uint64
	GetFloat(index int) *float64
	GetString(index int) *string
	GetTime(index int) *time.Time
	GetDuration(index int) *time.Duration
	GetConfig(index int) Config
}

func EmptyConfig() Config {
	return &emptyConfig
}

func ConfigSkip(config Config, skips ...string) Config {
	sort.Strings(skips)
	return skipConfig{config: config, skips: skips}
}

func ConfigCombine(first Config, rest ...Config) Config {
	configs := append([]Config{first}, rest...)
	return combineConfig{configs: configs}
}

func ConfigGetBool(config Config, key string) *bool {
	return config.GetBool(key)
}

func ConfigGetInt8(config Config, key string) *int8 {
	result := config.GetInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt8 || *result < math.MinInt8 {
		return nil
	}
	value := int8(*result)
	return &value
}

func ConfigGetUInt8(config Config, key string) *uint8 {
	result := config.GetUInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint8 {
		return nil
	}
	value := uint8(*result)
	return &value
}

func ConfigGetInt16(config Config, key string) *int16 {
	result := config.GetInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt16 || *result < math.MinInt16 {
		return nil
	}
	value := int16(*result)
	return &value
}

func ConfigGetUInt16(config Config, key string) *uint16 {
	result := config.GetUInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint16 {
		return nil
	}
	value := uint16(*result)
	return &value
}

func ConfigGetInt32(config Config, key string) *int32 {
	result := config.GetInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt32 || *result < math.MinInt32 {
		return nil
	}
	value := int32(*result)
	return &value
}

func ConfigGetUInt32(config Config, key string) *uint32 {
	result := config.GetUInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint32 {
		return nil
	}
	value := uint32(*result)
	return &value
}

func ConfigGetInt64(config Config, key string) *int64 {
	return config.GetInt(key)
}

func ConfigGetUInt64(config Config, key string) *uint64 {
	return config.GetUInt(key)
}

func ConfigGetInt(config Config, key string) *int {
	result := config.GetInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt || *result < math.MinInt {
		return nil
	}
	value := int(*result)
	return &value
}

func ConfigGetUInt(config Config, key string) *uint {
	result := config.GetUInt(key)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint {
		return nil
	}
	value := uint(*result)
	return &value
}

func ConfigGetFloat32(config Config, key string) *float32 {
	result := config.GetFloat(key)
	if result == nil {
		return nil
	}
	if math.Abs(*result) > math.MaxFloat32 {
		return nil
	}
	value := float32(*result)
	return &value
}

func ConfigGetFloat64(config Config, key string) *float64 {
	return config.GetFloat(key)
}

func ConfigGetString(config Config, key string) *string {
	return config.GetString(key)
}

func ConfigGetTime(config Config, key string) *time.Time {
	return config.GetTime(key)
}

func ConfigGetDuration(config Config, key string) *time.Duration {
	return config.GetDuration(key)
}

func ConfigGetConfig(config Config, key string) Config {
	return config.GetConfig(key)
}

func ConfigGetArray(config Config, key string) Array {
	return config.GetArray(key)
}

func ConfigGetEnum[T any](config Config, key string, enum func(string) *T) *T {
	result := config.GetString(key)
	if result == nil {
		return nil
	}
	return enum(*result)
}

func ConfigGetType[T Configurable](config Config, key string) *T {
	result1 := config.GetConfig(key)
	if result1 == nil {
		return nil
	}
	result2 := new(T)
	if !(*result2).LoadConfig(result1) {
		return nil
	}
	return result2
}

func ConfigGetBoolArray(config Config, key string) []bool {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsBoolArray(array)
}

func ConfigGetInt8Array(config Config, key string) []int8 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsInt8Array(array)
}

func ConfigGetUInt8Array(config Config, key string) []uint8 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsUInt8Array(array)
}

func ConfigGetInt16Array(config Config, key string) []int16 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsInt16Array(array)
}

func ConfigGetUInt16Array(config Config, key string) []uint16 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsUInt16Array(array)
}

func ConfigGetInt32Array(config Config, key string) []int32 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsInt32Array(array)
}

func ConfigGetUInt32Array(config Config, key string) []uint32 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsUInt32Array(array)
}

func ConfigGetInt64Array(config Config, key string) []int64 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsInt64Array(array)
}

func ConfigGetUInt64Array(config Config, key string) []uint64 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsUInt64Array(array)
}

func ConfigGetIntArray(config Config, key string) []int {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsIntArray(array)
}

func ConfigGetUIntArray(config Config, key string) []uint {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsUIntArray(array)
}

func ConfigGetFloat32Array(config Config, key string) []float32 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsFloat32Array(array)
}

func ConfigGetFloat64Array(config Config, key string) []float64 {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsFloat64Array(array)
}

func ConfigGetStringArray(config Config, key string) []string {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsStringArray(array)
}

func ConfigGetTimeArray(config Config, key string) []time.Time {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsTimeArray(array)
}

func ConfigGetDurationArray(config Config, key string) []time.Duration {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsDurationArray(array)
}

func ConfigGetConfigArray(config Config, key string) []Config {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsConfigArray(array)
}

func ConfigGetEnumArray[T any](config Config, key string, enum func(string) *T) []T {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsEnumArray(array, enum)
}

func ConfigGetTypeArray[T Configurable](config Config, key string) []T {
	array := config.GetArray(key)
	if array == nil {
		return nil
	}
	return ArrayAsTypeArray[T](array)
}

func ArrayGetBool(array Array, index int) *bool {
	return array.GetBool(index)
}

func ArrayGetInt8(array Array, index int) *int8 {
	result := array.GetInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt8 || *result < math.MinInt8 {
		return nil
	}
	value := int8(*result)
	return &value
}

func ArrayGetUInt8(array Array, index int) *uint8 {
	result := array.GetUInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint {
		return nil
	}
	value := uint8(*result)
	return &value
}

func ArrayGetInt16(array Array, index int) *int16 {
	result := array.GetInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt16 || *result < math.MinInt16 {
		return nil
	}
	value := int16(*result)
	return &value
}

func ArrayGetUInt16(array Array, index int) *uint16 {
	result := array.GetUInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint {
		return nil
	}
	value := uint16(*result)
	return &value
}

func ArrayGetInt32(array Array, index int) *int32 {
	result := array.GetInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt32 || *result < math.MinInt32 {
		return nil
	}
	value := int32(*result)
	return &value
}

func ArrayGetUInt32(array Array, index int) *uint32 {
	result := array.GetUInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint {
		return nil
	}
	value := uint32(*result)
	return &value
}

func ArrayGetInt64(array Array, index int) *int64 {
	return array.GetInt(index)
}

func ArrayGetUInt64(array Array, index int) *uint64 {
	return array.GetUInt(index)
}

func ArrayGetInt(array Array, index int) *int {
	result := array.GetInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxInt || *result < math.MinInt {
		return nil
	}
	value := int(*result)
	return &value
}

func ArrayGetUInt(array Array, index int) *uint {
	result := array.GetUInt(index)
	if result == nil {
		return nil
	}
	if *result > math.MaxUint {
		return nil
	}
	value := uint(*result)
	return &value
}

func ArrayGetFloat32(array Array, index int) *float32 {
	result := array.GetFloat(index)
	if result == nil {
		return nil
	}
	if math.Abs(*result) > math.MaxFloat32 {
		return nil
	}
	value := float32(*result)
	return &value
}

func ArrayGetFloat64(array Array, index int) *float64 {
	return array.GetFloat(index)
}

func ArrayGetString(array Array, index int) *string {
	return array.GetString(index)
}

func ArrayGetTime(array Array, index int) *time.Time {
	return array.GetTime(index)
}

func ArrayGetDuration(array Array, index int) *time.Duration {
	return array.GetDuration(index)
}

func ArrayGetConfig(array Array, index int) Config {
	return array.GetConfig(index)
}

func ArrayAsBoolArray(array Array) []bool {
	count := array.Count()
	results := make([]bool, count)
	for i := 0; i < count; i++ {
		result := ArrayGetBool(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsInt8Array(array Array) []int8 {
	count := array.Count()
	results := make([]int8, count)
	for i := 0; i < count; i++ {
		result := ArrayGetInt8(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsUInt8Array(array Array) []uint8 {
	count := array.Count()
	results := make([]uint8, count)
	for i := 0; i < count; i++ {
		result := ArrayGetUInt8(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsInt16Array(array Array) []int16 {
	count := array.Count()
	results := make([]int16, count)
	for i := 0; i < count; i++ {
		result := ArrayGetInt16(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsUInt16Array(array Array) []uint16 {
	count := array.Count()
	results := make([]uint16, count)
	for i := 0; i < count; i++ {
		result := ArrayGetUInt16(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsInt32Array(array Array) []int32 {
	count := array.Count()
	results := make([]int32, count)
	for i := 0; i < count; i++ {
		result := ArrayGetInt32(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsUInt32Array(array Array) []uint32 {
	count := array.Count()
	results := make([]uint32, count)
	for i := 0; i < count; i++ {
		result := ArrayGetUInt32(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsInt64Array(array Array) []int64 {
	count := array.Count()
	results := make([]int64, count)
	for i := 0; i < count; i++ {
		result := ArrayGetInt64(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsUInt64Array(array Array) []uint64 {
	count := array.Count()
	results := make([]uint64, count)
	for i := 0; i < count; i++ {
		result := ArrayGetUInt64(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsIntArray(array Array) []int {
	count := array.Count()
	results := make([]int, count)
	for i := 0; i < count; i++ {
		result := ArrayGetInt(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsUIntArray(array Array) []uint {
	count := array.Count()
	results := make([]uint, count)
	for i := 0; i < count; i++ {
		result := ArrayGetUInt(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsFloat32Array(array Array) []float32 {
	count := array.Count()
	results := make([]float32, count)
	for i := 0; i < count; i++ {
		result := ArrayGetFloat32(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsFloat64Array(array Array) []float64 {
	count := array.Count()
	results := make([]float64, count)
	for i := 0; i < count; i++ {
		result := ArrayGetFloat64(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsStringArray(array Array) []string {
	count := array.Count()
	results := make([]string, count)
	for i := 0; i < count; i++ {
		result := ArrayGetString(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsTimeArray(array Array) []time.Time {
	count := array.Count()
	results := make([]time.Time, count)
	for i := 0; i < count; i++ {
		result := ArrayGetTime(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsDurationArray(array Array) []time.Duration {
	count := array.Count()
	results := make([]time.Duration, count)
	for i := 0; i < count; i++ {
		result := ArrayGetDuration(array, i)
		if result == nil {
			return nil
		}
		results[i] = *result
	}
	return results
}

func ArrayAsConfigArray(array Array) []Config {
	count := array.Count()
	results := make([]Config, count)
	for i := 0; i < count; i++ {
		result := ArrayGetConfig(array, i)
		if result == nil {
			return nil
		}
		results[i] = result
	}
	return results
}

func ArrayAsEnumArray[T any](array Array, enum func(string) *T) []T {
	count := array.Count()
	results := make([]T, count)
	for i := 0; i < count; i++ {
		result1 := ArrayGetString(array, i)
		if result1 == nil {
			return nil
		}
		result2 := enum(*result1)
		if result2 == nil {
			return nil
		}
		results[i] = *result2
	}
	return results
}

func ArrayAsTypeArray[T Configurable](array Array) []T {
	count := array.Count()
	results := make([]T, count)
	for i := 0; i < count; i++ {
		result1 := ArrayGetConfig(array, i)
		if result1 == nil {
			return nil
		}
		result2 := new(T)
		if !(*result2).LoadConfig(result1) {
			return nil
		}
		results[i] = *result2
	}
	return results
}

var emptyConfig empty

type empty struct {
}

func (this *empty) Empty() bool {
	return true
}

func (this *empty) Keys() interface{ Next() *string } {
	return this
}

func (this *empty) Next() *string {
	return nil
}

func (this *empty) GetBool(key string) *bool {
	return nil
}

func (this *empty) GetInt(key string) *int64 {
	return nil
}

func (this *empty) GetUInt(key string) *uint64 {
	return nil
}

func (this *empty) GetFloat(key string) *float64 {
	return nil
}

func (this *empty) GetString(key string) *string {
	return nil
}

func (this *empty) GetTime(key string) *time.Time {
	return nil
}

func (this *empty) GetDuration(key string) *time.Duration {
	return nil
}

func (this *empty) GetConfig(key string) Config {
	return nil
}

func (this *empty) GetArray(key string) Array {
	return nil
}

type skipConfig struct {
	config Config
	skips  []string
}

type skipkeys struct {
	keys  interface{ Next() *string }
	skips []string
}

func (this skipkeys) Next() *string {
	for {
		key := this.keys.Next()
		if key != nil {
			index := sort.SearchStrings(this.skips, *key)
			if index != len(this.skips) && this.skips[index] == *key {
				continue
			}
		}
		return key
	}
}

func (this skipConfig) Empty() bool {
	if this.config.Empty() {
		return true
	}
	keys := this.config.Keys()
	for {
		key := keys.Next()
		if key == nil {
			return true
		}
		index := sort.SearchStrings(this.skips, *key)
		if index == len(this.skips) {
			return false
		}
		if this.skips[index] != *key {
			return false
		}
	}
}

func (this skipConfig) Keys() interface{ Next() *string } {
	return skipkeys{
		keys:  this.config.Keys(),
		skips: this.skips,
	}
}

func (this skipConfig) GetBool(key string) *bool {
	return this.config.GetBool(key)
}

func (this skipConfig) GetInt(key string) *int64 {
	return this.config.GetInt(key)
}

func (this skipConfig) GetUInt(key string) *uint64 {
	return this.config.GetUInt(key)
}

func (this skipConfig) GetFloat(key string) *float64 {
	return this.config.GetFloat(key)
}

func (this skipConfig) GetString(key string) *string {
	return this.config.GetString(key)
}

func (this skipConfig) GetTime(key string) *time.Time {
	return this.config.GetTime(key)
}

func (this skipConfig) GetDuration(key string) *time.Duration {
	return this.config.GetDuration(key)
}

func (this skipConfig) GetConfig(key string) Config {
	return this.config.GetConfig(key)
}

func (this skipConfig) GetArray(key string) Array {
	return this.config.GetArray(key)
}

type combineConfig struct {
	configs []Config
}

type combinekeys struct {
	used    map[string]struct{}
	current interface{ Next() *string }
	configs []Config
	index   int
}

func (this *combinekeys) Next() *string {
	for {
		if this.current == nil {
			if this.index >= len(this.configs) {
				return nil
			}
			this.current = this.configs[this.index].Keys()
			this.index++
		}
		result := this.current.Next()
		if result != nil {
			_, exists := this.used[*result]
			if exists {
				continue
			}
			this.used[*result] = Void
			return result
		}
		this.current = nil
	}
}

func (this combineConfig) Empty() bool {
	for _, config := range this.configs {
		if !config.Empty() {
			return false
		}
	}
	return true
}

func (this combineConfig) Keys() interface{ Next() *string } {
	return &combinekeys{
		used:    make(map[string]struct{}),
		configs: this.configs,
		index:   0,
	}
}

func (this combineConfig) GetBool(key string) *bool {
	for _, config := range this.configs {
		result := config.GetBool(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetInt(key string) *int64 {
	for _, config := range this.configs {
		result := config.GetInt(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetUInt(key string) *uint64 {
	for _, config := range this.configs {
		result := config.GetUInt(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetFloat(key string) *float64 {
	for _, config := range this.configs {
		result := config.GetFloat(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetString(key string) *string {
	for _, config := range this.configs {
		result := config.GetString(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetTime(key string) *time.Time {
	for _, config := range this.configs {
		result := config.GetTime(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetDuration(key string) *time.Duration {
	for _, config := range this.configs {
		result := config.GetDuration(key)
		if result != nil {
			return result
		}
	}
	return nil
}

func (this combineConfig) GetConfig(key string) Config {
	var configs []Config
	for _, config := range this.configs {
		result := config.GetConfig(key)
		if result != nil {
			configs = append(configs, result)
		}
	}
	if len(configs) == 0 {
		return nil
	}
	return combineConfig{configs: configs}
}

func (this combineConfig) GetArray(key string) Array {
	var arrays []Array
	for _, config := range this.configs {
		result := config.GetArray(key)
		if result != nil {
			arrays = append(arrays, result)
		}
	}
	if len(arrays) == 0 {
		return nil
	}
	return combineArray{arrays: arrays}
}

type combineArray struct {
	arrays []Array
}

func (this combineArray) Count() int {
	sum := 0
	for _, array := range this.arrays {
		sum += array.Count()
	}
	return sum
}

func (this combineArray) GetBool(index int) *bool {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetBool(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetInt(index int) *int64 {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetInt(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetUInt(index int) *uint64 {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetUInt(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetFloat(index int) *float64 {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetFloat(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetString(index int) *string {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetString(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetTime(index int) *time.Time {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetTime(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetDuration(index int) *time.Duration {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetDuration(index)
		}
		index -= count
	}
	return nil
}

func (this combineArray) GetConfig(index int) Config {
	for _, array := range this.arrays {
		count := array.Count()
		if index < count {
			return array.GetConfig(index)
		}
		index -= count
	}
	return nil
}
