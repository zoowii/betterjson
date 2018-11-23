package betterjson

import (
	"github.com/bitly/go-simplejson"
	"github.com/pkg/errors"
	"log"
	"encoding/json"
	"bytes"
	"sort"
)

// Json is immutable type when it's empty
type Json struct {
	value *simplejson.Json
}

type jsonWithItemKeyValue struct {
	json *Json
	key string
	value *Json
}

func FromSimpleJson(jsonVal *simplejson.Json) (*Json, error) {
	if jsonVal == nil {
		return nil, errors.New("can't parse null to json")
	}
	json := new(Json)
	json.value = jsonVal
	return json, nil
}

func FromNotEmptySimpleJson(jsonVal *simplejson.Json) *Json {
	json := new(Json)
	json.value = jsonVal
	return json
}

func NewEmpty() *Json {
	json := new(Json)
	json.value = nil
	return json
}

func NewJSONObject() *Json {
	json := new(Json)
	json.value = simplejson.New()
	return json
}

func NewJSONArray() *Json {
	json := new(Json)
	json.value = simplejson.New()
	json.value.SetPath([]string{}, make([]interface{}, 0))
	return json
}

func (val *Json) ToSimpleJson() *simplejson.Json {
	return val.value
}

func (val *Json) IsEmpty() bool {
	return val.value == nil
}

func (val *Json) IsNullJson() bool {
	encoded, err := val.value.Encode()
	if err != nil {
		return false
	}
	return len(encoded) == 4 && string(encoded) == "null"
}

func (j *Json) IsEmptyOrNull() bool {
	return j.IsEmpty() || j.IsNullJson()
}

func (val *Json) Select(key string) *Json {
	if val.IsEmpty() {
		return val
	}
 	item, ok := val.value.CheckGet(key)
 	if !ok {
 		return val
	}
	return FromNotEmptySimpleJson(item)
}

type JsonValueProcessor func (json *simplejson.Json) *simplejson.Json

func (val *Json) Apply(processor JsonValueProcessor) *Json {
	if val.IsEmpty() {
		return val
	}
	result := processor(val.value)
	if result == nil {
		return NewEmpty()
	}
	return FromNotEmptySimpleJson(result)
}

// return new json with all key-values if all the keys exists
func (val *Json) GetKeyValuesIfAllContains(keys []string) *Json {
	if val.IsEmpty() {
		return val
	}
	result := simplejson.New()
	emptyInnerJson := simplejson.New()
	for _, key := range keys {
		itemVal, ok := val.value.CheckGet(key)
		if !ok {
			return FromNotEmptySimpleJson(emptyInnerJson)
		}
		result.Set(key, itemVal)
	}
	return FromNotEmptySimpleJson(result)
}

// WithKey({a: b, ...remaining}, key) => ({a: b, ...remaining}, a, b)
func (j *Json) WithKey(key string) *jsonWithItemKeyValue {
	result := new(jsonWithItemKeyValue)
	result.json = j
	result.key = key
	if j.IsEmpty() {
		result.value = j
		return result
	}
	itemVal := j.Get(key)
	if itemVal.IsEmptyOrNull() {
		result.value = NewEmpty()
		return result
	}
	result.value = itemVal
	return result
}

type JsonKeyValueProcessor = func(*Json, string, *Json)*Json

func (j *jsonWithItemKeyValue) Apply(processor JsonKeyValueProcessor) *Json {
	if j.json.IsEmpty() {
		return j.json
	}
	return processor(j.json, j.key, j.value)
}

func (j *Json) TrampolineKeys(keys []string, processors []JsonKeyValueProcessor, initJson *Json) (*Json, error) {
	if j.IsEmpty() {
		return initJson, nil
	}
	resultJson := initJson
	if len(keys) > len(processors) {
		return initJson, errors.New("keys count great than processor funcs count")
	}
	for i, key := range keys {
		processor := processors[i]
		item := j.CheckGet(key)
		resultJson = processor(resultJson, key, item)
	}
	return resultJson, nil
}

// CheckGet returns a pointer to a new `Json` object and
// a `bool` identifying success or failure
//
// useful for chained operations when success is important:
//    if data, ok := js.Get("top_level").CheckGet("inner"); ok {
//        log.Println(data)
//    }
func (j *Json)CheckGet(key string) *Json {
	if j.IsEmpty() {
		return j
	}
	item, ok := j.value.CheckGet(key)
	if !ok {
		return NewEmpty()
	}
	return FromNotEmptySimpleJson(item)
}

// Interface returns the underlying data
func (j *Json) Interface() interface{} {
	if j.IsEmpty() {
		return nil
	}
	return j.value.Interface()
}

func (j *Json)Set(key string, val interface{}) *Json {
	if j.IsEmpty() {
		return j
	}
	if val == nil {
		j.value.Set(key, val)
		return j
	}
	valJson, valIsJson := val.(*Json)
	if valIsJson {
		if valJson.IsEmpty() {
			j.value.Set(key, nil)
		} else {
			j.value.Set(key, valJson.value.Interface())
		}
	} else {
		valSimpleJson, valIsSimpleJson := val.(*simplejson.Json)
		if valIsSimpleJson {
			j.value.Set(key, valSimpleJson.Interface())
		} else {
			j.value.Set(key, val)
		}
	}
	return j
}

// SetPath modifies `Json`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value
func (j *Json) SetPath(branch []string, val interface{}) *Json {
	valJson, valIsJSON := val.(*Json)
	if j.IsEmpty() {
		if valIsJSON {
			*j = *valJson
		} else {
			j.value = simplejson.New()
			j.value.SetPath(branch, val)
		}
		return j
	}
	if len(branch) == 0 {
		if valIsJSON {
			*j = *valJson
		} else {
			j.value.SetPath(branch, val)
		}
		return j
	}
	if valIsJSON {
		j.value.SetPath(branch, valJson.value)
	} else {
		j.value.SetPath(branch, val)
	}
	return j
}

// Del modifies `Json` map by deleting `key` if it is present.
func (j *Json) Del(key string) *Json {
	if j.IsEmpty() {
		return j
	}
	j.value.Del(key)
	return j
}

// Get returns a pointer to a new `Json` object
// for `key` in its `map` representation
//
// useful for chaining operations (to traverse a nested JSON):
//    js.Get("top_level").Get("dict").Get("value").Int()
func (j *Json) Get(key string) *Json {
	return FromNotEmptySimpleJson(j.value.Get(key))
}

// GetPath searches for the item as specified by the branch
// without the need to deep dive using Get()'s.
//
//   js.GetPath("top_level", "dict")
func (j *Json) GetPath(branch ...string) *Json {
	jin := j
	for _, p := range branch {
		jin = jin.Get(p)
	}
	return jin
}

// GetIndex returns a pointer to a new `Json` object
// for `index` in its `array` representation
//
// this is the analog to Get when accessing elements of
// a json array instead of a json object:
//    js.Get("top_level").Get("array").GetIndex(1).Get("key").Int()
func (j *Json) GetIndex(index int) *Json {
	return FromNotEmptySimpleJson(j.value.GetIndex(index))
}


// Map type asserts to `map`
func (j *Json) Map() (map[string]interface{}, error) {
	if j.IsEmpty() {
		return nil, errors.New("empty json parse to map[string]interface{} failed")
	}
	return j.value.Map()
}

// Array type asserts to an `array`
func (j *Json) Array() ([]interface{}, error) {
	if j.IsEmpty() {
		return nil, errors.New("empty json parse to []interface{} failed")
	}
	return j.value.Array()
}

// Bool type asserts to `bool`
func (j *Json) Bool() (bool, error) {
	if j.IsEmpty() {
		return false, errors.New("empty json parse to bool failed")
	}
	return j.value.Bool()
}

// String type asserts to `string`
func (j *Json) String() (string, error) {
	if j.IsEmpty() {
		return "", errors.New("empty json parse to string failed")
	}
	return j.value.String()
}

// Bytes type asserts to `[]byte`
func (j *Json) Bytes() ([]byte, error) {
	if j.IsEmpty() {
		return nil, errors.New("empty json parse to []byte failed")
	}
	return j.value.Bytes()
}

// StringArray type asserts to an `array` of `string`
func (j *Json) StringArray() ([]string, error) {
	if j.IsEmpty() {
		return nil, errors.New("empty json parse to []string failed")
	}
	return j.value.StringArray()
}

// MustArray guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, v := range js.Get("results").MustArray() {
//			fmt.Println(i, v)
//		}
func (j *Json) MustArray(args ...[]interface{}) []interface{} {
	if j.IsEmpty() {
		log.Panicf("empty json MustArray failed")
		return nil
	}
	return j.value.MustArray(args...)
}

// MustMap guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
//		for k, v := range js.Get("dictionary").MustMap() {
//			fmt.Println(k, v)
//		}
func (j *Json) MustMap(args ...map[string]interface{}) map[string]interface{} {
	if j.IsEmpty() {
		log.Panicf("empty json MustMap failed")
		return nil
	}
	return j.value.MustMap(args...)
}

// MustString guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//     myFunc(js.Get("param1").MustString(), js.Get("optional_param").MustString("my_default"))
func (j *Json) MustString(args ...string) string {
	if j.IsEmpty() {
		log.Panicf("empty json MustString failed")
		return ""
	}
	return j.value.MustString(args...)
}

// MustStringArray guarantees the return of a `[]string` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
//		for i, s := range js.Get("results").MustStringArray() {
//			fmt.Println(i, s)
//		}
func (j *Json) MustStringArray(args ...[]string) []string {
	if j.IsEmpty() {
		log.Panicf("empty json MustStringArray failed")
		return nil
	}
	return j.value.MustStringArray(args...)
}

// MustInt guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//     myFunc(js.Get("param1").MustInt(), js.Get("optional_param").MustInt(5150))
func (j *Json) MustInt(args ...int) int {
	if j.IsEmpty() {
		log.Panicf("empty json MustInt failed")
		return 0
	}
	return j.value.MustInt(args...)
}

// MustFloat64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     myFunc(js.Get("param1").MustFloat64(), js.Get("optional_param").MustFloat64(5.150))
func (j *Json) MustFloat64(args ...float64) float64 {
	if j.IsEmpty() {
		log.Panicf("empty json MustFloat64 failed")
		return 0
	}
	return j.value.MustFloat64(args...)
}

// MustBool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//     myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (j *Json) MustBool(args ...bool) bool {
	if j.IsEmpty() {
		log.Panicf("empty json MustBool failed")
		return false
	}
	return j.value.MustBool(args...)
}

// MustInt64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//     myFunc(js.Get("param1").MustInt64(), js.Get("optional_param").MustInt64(5150))
func (j *Json) MustInt64(args ...int64) int64 {
	if j.IsEmpty() {
		log.Panicf("empty json MustInt64 failed")
		return 0
	}
	return j.value.MustInt64(args...)
}

// MustUInt64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     myFunc(js.Get("param1").MustUint64(), js.Get("optional_param").MustUint64(5150))
func (j *Json) MustUint64(args ...uint64) uint64 {
	if j.IsEmpty() {
		log.Panicf("empty json MustUint64 failed")
		return 0
	}
	return j.value.MustUint64(args...)
}

func (j *Json)Encode() ([]byte, error) {
	if j.IsEmpty() {
		return []byte{}, errors.New("empty json can't be encoded")
	}
	return j.value.Encode()
}

func (j *Json)EncodeToString() (string, error) {
	bs, err := j.Encode()
	if err != nil {
		return "", err
	}
	return string(bs), err
}

func (j *Json)EncodeToStringOrDefault(defaultVal string) string {
	bs, err := j.Encode()
	if err != nil {
		return defaultVal
	}
	return string(bs)
}

func (j *Json) DigestJSONForEqual() string {
	if j.IsEmpty() {
		return "nil"
	}
	jsonVal := j
	jsonArray, err := jsonVal.Array()
	if err == nil {
		var digestBuffer bytes.Buffer
		digestBuffer.WriteString("[")
		for idx, _ := range jsonArray {
			if idx > 0 {
				digestBuffer.WriteString(",")
			}
			itemJson := jsonVal.GetIndex(idx)
			digestBuffer.WriteString(itemJson.DigestJSONForEqual())
		}
		digestBuffer.WriteString("]")
		return digestBuffer.String()
	}
	jsonMap, err := jsonVal.Map()
	if err == nil {
		var digestBuffer bytes.Buffer
		digestBuffer.WriteString("{")
		keys := make([]string, 0)
		for k, _ := range jsonMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for idx, key := range keys {
			if idx > 0 {
				digestBuffer.WriteString(",")
			}
			keyEncode, err := json.Marshal(key)
			if err != nil {
				digestBuffer.WriteString("\"error\":\"error\"")
				continue
			}
			item := jsonVal.Get(key)
			digestBuffer.WriteString(string(keyEncode))
			digestBuffer.WriteString(":")
			digestBuffer.WriteString(item.DigestJSONForEqual())
		}
		digestBuffer.WriteString("}")
		return digestBuffer.String()
	}
	encoded, err := jsonVal.Encode()
	if err != nil {
		return "error"
	}
	encodedStr := string(encoded)
	return encodedStr
}

// whether json a and json b have the same value
func (j *Json) IsSameJSONWith(other *Json) bool {
	if other==nil || other.IsEmpty() {
		return j.IsEmpty()
	}

	return j.DigestJSONForEqual() == other.DigestJSONForEqual()
}

// try add item when is array
func (j *Json) TryAdd(val interface{}) *Json {
	jsonArray, err := j.Array()
	if err != nil {
		return j
	}
	valJson, valIsJson := val.(*Json)
	if valIsJson {
		if valJson.IsEmpty() {
			jsonArray = append(jsonArray, nil)
		} else {
			jsonArray = append(jsonArray, valJson.value)
		}
	} else {
		jsonArray = append(jsonArray, val)
	}
	j.SetPath([]string{}, jsonArray)
	return j
}

func (j *Json) ArrayLength() int {
	jsonArray, err := j.Array()
	if err != nil {
		return 0
	}
	return len(jsonArray)
}

func (j *Json) ContainsKey(key string) bool {
	if j.IsEmpty() {
		return false
	}
	val := j.CheckGet(key)
	return !val.IsEmpty()
}

func (j *Json) SetValue(val interface{}) *Json {
	j.SetPath([]string{}, val)
	return j
}