package json

import (
	"fmt"
	"reflect"
	"strings"
)

// override common json by input json
func Override(input, common interface{}) {
	switch inputData := input.(type) {
	case []interface{}:
		switch commonData := common.(type) {
		case []interface{}:
			for idx, v := range inputData {
				Override(v, commonData[idx])
			}
		}
	case map[string]interface{}:
		switch commonData := common.(type) {
		case map[string]interface{}:
			for k, v := range commonData {
				switch reflect.TypeOf(v).Kind() {
				case reflect.Slice, reflect.Map:
					Override(inputData[k], v)
				default:
					// do simply replacement for primitive type
					_, ok := inputData[k]
					if !ok {
						inputData[k] = v
					}
				}
			}
		}
	}
	return
}

// replace predefined value with value from cache
func Fill(input interface{}, cache map[string]interface{}) {
	switch inputData := input.(type) {
	case []interface{}:
		for _, v := range inputData {
			Fill(v, cache)
		}
	case map[string]interface{}:
		for k, v := range inputData {
			switch reflect.TypeOf(v).Kind() {
			case reflect.Slice, reflect.Map:
				Fill(v, cache)
			default:
				// do simply fill for primitive type
				ss := strings.Split(fmt.Sprint(v), tokenSeparator)
				if len(ss) != 2 {
					continue
				}
				typ := ss[0]
				if typ == tokenInput {
					inputData[k], _ = cache[ss[1]]
				}
			}
		}
	}
	return
}
