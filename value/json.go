package value

import (
	"encoding/json"
	"io"
)

func decodeJson(a interface{}) Value {
	var val Value = Zero{}

	switch v := a.(type) {
	case string:
		val = String{Value: v}
	case float64:
		if v == float64(int64(v)) {
			val = Int{Value: int64(v)}
			break
		}
		val = Float{Value: v}
	case bool:
		val = Bool{Value: v}
	case []interface{}:
		arr := &Array{
			set:   make(map[uint32]struct{}),
			Items: make([]Value, 0, len(v)),
		}

		for _, a := range v {
			arr.Items = append(arr.Items, decodeJson(a))
		}

		arr.hashItems()

		val = arr
	case map[string]interface{}:
		obj := &Object{
			Order: make([]string, 0, len(v)),
			Pairs: make(map[string]Value),
		}

		for k, a := range v {
			obj.Order = append(obj.Order, k)
			obj.Pairs[k] = decodeJson(a)
		}
		val = obj
	}
	return val
}

// DecodeJSON attempts to decode all of the data in the given reader to JSON.
// The returned value will either be of type String, Int, Bool, Array, Object,
// or Zero depending on the JSON string being decoded.
func DecodeJSON(r io.Reader) (Value, error) {
	var p interface{}

	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, err
	}
	return decodeJson(p), nil
}
