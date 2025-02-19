package json

import (
	"encoding/json"
	"fmt"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestSimple(t *testing.T) {
	const str = `
	local json = require("json")
	assert(type(json) == "table")
	assert(type(json.decode) == "function")
	assert(type(json.encode) == "function")

	assert(json.encode(true) == "true")
	assert(json.encode(1) == "1")
	assert(json.encode(-10) == "-10")
	assert(json.encode(nil) == "null")
	assert(json.encode({}) == "[]")
	assert(json.encode({1, 2, 3}) == "[1,2,3]")

	local _, err = json.encode({1, 2, 3, name = "Tim"})
	assert(string.find(err, "mixed or invalid key types"))

	local _, err = json.encode({name = "Tim", [false] = 123})
	assert(string.find(err, "mixed or invalid key types"))

	local obj = {"a",1,"b",2,"c",3}
	local jsonStr = json.encode(obj)
	local jsonObj = json.decode(jsonStr)
	for i = 1, #obj do
		assert(obj[i] == jsonObj[i])
	end

	local obj = {name="Tim",number=12345}
	local jsonStr = json.encode(obj)
	local jsonObj = json.decode(jsonStr)
	assert(obj.name == jsonObj.name)
	assert(obj.number == jsonObj.number)

	assert(json.decode("null") == nil)

	assert(json.decode(json.encode({person={name = "tim",}})).person.name == "tim")

	local obj = {
		abc = 123,
		def = nil,
	}
	local obj2 = {
		obj = obj,
	}
	obj.obj2 = obj2
	assert(json.encode(obj) == nil)

	local a = {}
	for i=1, 5 do
		a[i] = i
	end
	assert(json.encode(a) == "[1,2,3,4,5]")

	-- UserData removal
	local t = setmetatable({10}, {
		__call = function(t, value)
			return value
		end
	})

	assert(t(37) == 37)
	assert(json.encode(t) == "[10]")
	`
	s := lua.NewState()
	defer s.Close()

	Preload(s)
	if err := s.DoString(str); err != nil {
		t.Error(err)
	}
}

func TestCustomRequire(t *testing.T) {
	const str = `
	local j = require("JSON")
	assert(type(j) == "table")
	assert(type(j.decode) == "function")
	assert(type(j.encode) == "function")
	`
	s := lua.NewState()
	defer s.Close()

	s.PreloadModule("JSON", Loader)
	if err := s.DoString(str); err != nil {
		t.Error(err)
	}
}

func TestDecodeValue_jsonNumber(t *testing.T) {
	s := lua.NewState()
	defer s.Close()

	v := DecodeValue(s, json.Number("124.11"))
	if v.Type() != lua.LTString || v.String() != "124.11" {
		t.Fatalf("expecting LString, got %T", v)
	}
}

func TestEncode_SparseArray(t *testing.T) {
	tests := []struct {
		table    string
		expected string
	}{
		{
			table: `{
				1,
				2,
				[10] = 3
			}`,
			expected: `[1,2,"[10] = 3"]`,
		},
		{
			table: `{
				nested = {
					[37] = "index 37"
				}
			}`,
			expected: `{"nested":["[37] = index 37"]}`,
		},
		{
			table: `{
				nested = {
					"index 1",
					[37] = "index 37"
				}
			}`,
			expected: `{"nested":["index 1","[37] = index 37"]}`,
		},
		{
			table: `{
				nested = {
					"index 1",
					[37] = "index 37"
				}
			}`,
			expected: `{"nested":["index 1","[37] = index 37"]}`,
		},
	}

	for _, test := range tests {
		s := lua.NewState()
		defer s.Close()
		Preload(s)

		luaScript := fmt.Sprintf(`
			local json = require("json")
			local t = %s
			assert(json.encode(t) == '%s')`, test.table, test.expected)
		if err := s.DoString(luaScript); err != nil {
			t.Error(err)
		}
	}
}
