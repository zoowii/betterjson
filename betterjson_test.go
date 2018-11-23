package betterjson

import (
	"testing"
	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestFromNotEmptySimpleJson(t *testing.T) {
	a := simplejson.New()
	a.Set("hello", "world")
	b := FromNotEmptySimpleJson(a)
	bStr, err := b.EncodeToString()
	fmt.Println(bStr)
	assert.True(t, err == nil)
	assert.True(t, bStr == "{\"hello\":\"world\"}")
}

func TestFromSimpleJson(t *testing.T) {
	a := simplejson.New()
	emptyItem := a.Get("name")
	b, err := FromSimpleJson(emptyItem)
	assert.True(t, err == nil)
	assert.True(t, b.IsNullJson())
	bStr, err := b.EncodeToString()
	fmt.Println(bStr)
	assert.True(t, err == nil)
	assert.True(t, bStr == "null")
}

func TestJson_Apply(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world")
	a.Apply(func (val *simplejson.Json) *simplejson.Json {
		val.Set("hello", 123)
		return val
	})
	bStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	fmt.Println(bStr)
	assert.True(t, bStr == "{\"hello\":123}")
}
func TestJson_Get(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world")
	c, err := a.Get("hello").EncodeToString()
	assert.True(t, err == nil)
	fmt.Println(c)
	assert.True(t, c == "\"world\"")
}

func TestJson_CheckGet(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world")
	c := a.CheckGet("hello")
	assert.True(t, !c.IsEmptyOrNull())
	d, err := c.EncodeToString()
	assert.True(t, err == nil)
	println(d)
	assert.True(t, d == "\"world\"")
}

func TestJson_DigestJSONForEqual(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world").Set("hi", NewJSONObject().Set("age", 18).Set("items", NewJSONArray().TryAdd(1).TryAdd(nil).TryAdd("China"))).Set("times", 123).Set("a", "head")
	aStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	println(aStr)
	aDigest := a.DigestJSONForEqual()
	println(aDigest)
	assert.True(t, aDigest == "{\"a\":\"head\",\"hello\":\"world\",\"hi\":{\"age\":18,\"items\":[1,null,\"China\"]},\"times\":123}")
}

func TestJson_WithKey(t *testing.T) {
	a := NewJSONObject()
	hiRawJSON := NewJSONObject().Set("age", 18).Set("items", NewJSONArray().TryAdd(1).TryAdd(nil).TryAdd("China"))
	a.Set("hello", "world").Set("hi", hiRawJSON).Set("times", 123).Set("a", "head")
	aStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	println(aStr)
	hiJson := a.Get("hi")
	hiMap, err := hiJson.Map()
	if err != nil {
		println(err.Error())
	}
	fmt.Println(hiMap)
	b := hiJson.WithKey("age").Apply(func (j *Json, key string, value *Json) *Json {
		return NewEmpty().SetValue(value.MustInt() * 100)
	})
	bStr, err := b.EncodeToString()
	assert.True(t, err == nil)
	println(bStr)
	assert.True(t, bStr == "1800")
}

func TestJson_GetKeyValuesIfAllContains(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world").Set("hi", NewJSONObject().Set("age", 18).Set("items", NewJSONArray().TryAdd(1).TryAdd(nil).TryAdd("China"))).Set("times", 123)
	aStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	println(aStr)
	b := a.GetKeyValuesIfAllContains([]string{"times", "hello"})
	bStr, err := b.EncodeToString()
	assert.True(t, err == nil)
	println(bStr)
	assert.True(t, b.ContainsKey("hello") && b.ContainsKey("times"))
}

func TestJson_IsSameJSONWith(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world").Set("hi", NewJSONObject().Set("age", 18).Set("items", NewJSONArray().TryAdd(1).TryAdd(nil).TryAdd("China"))).Set("times", 123)
	aStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	println(aStr)
	bj1, err := simplejson.NewJson([]byte(aStr))
	assert.True(t, err == nil)
	bj2, err := simplejson.NewJson([]byte(aStr))
	assert.True(t, err == nil)
	b1, b2 := FromNotEmptySimpleJson(bj1), FromNotEmptySimpleJson(bj2)
	b12Same := b1.IsSameJSONWith(b2)
	assert.True(t, b12Same)
}

func TestJson_IsNullJson(t *testing.T) {
	a := NewJSONObject()
	b := a.Get("hello")
	assert.True(t, b.IsNullJson())
}

func TestJson_ToSimpleJson(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world").Set("hi", NewJSONObject().Set("age", 18).Set("items", NewJSONArray().TryAdd(1).TryAdd(nil).TryAdd("China"))).Set("times", 123)
	aStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	println(aStr)
	b := a.ToSimpleJson()
	bBytes, err := b.Encode()
	assert.True(t, err == nil)
	bStr := string(bBytes)
	println(bStr)
	assert.True(t, len(bStr)==len(aStr))
}

func TestJson_TrampolineKeys(t *testing.T) {
	a := NewJSONObject()
	a.Set("hello", "world").Set("hi", NewJSONObject().Set("age", 18).Set("items", NewJSONArray().TryAdd(1).TryAdd(nil).TryAdd("China"))).Set("times", 123)
	aStr, err := a.EncodeToString()
	assert.True(t, err == nil)
	println(aStr)
	var countFunc = func (resultJSON *Json, key string, value *Json) *Json {
		resultJSON.SetValue(resultJSON.MustInt()+1)
		return resultJSON
	}
	resultJSON, err := a.TrampolineKeys([]string{"age", "hello"}, []JsonKeyValueProcessor{countFunc, countFunc}, NewEmpty().SetValue(0))
	assert.True(t, err == nil)
	println("result count: ", resultJSON.MustInt())
	assert.True(t, resultJSON.MustInt() == 2)
}