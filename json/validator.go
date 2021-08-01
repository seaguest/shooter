package json

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator"
)

const (
	tokenInt       = "{int}"
	tokenFloat     = "{float}"
	tokenString    = "{string}"
	tokenList      = "{list}"
	tokenRe        = "{re}" // regular expression
	tokenInput     = "{input}"
	tokenOutput    = "{output}"
	tokenSeparator = "|"

	typeErrMsg = "tag [%s] expected [%s], but received [%+v]"
)

type Validator struct {
	vd    *validator.Validate
	cache map[string]interface{}
}

func NewValidator() *Validator {
	v := &Validator{}
	v.vd = validator.New()
	v.cache = make(map[string]interface{})
	return v
}

func (v *Validator) GetCache() map[string]interface{} {
	return v.cache
}

func (v *Validator) Validate(val, tpl interface{}, tag string) (err error) {
	if tpl == nil {
		return
	}

	switch inputData := val.(type) {
	case []interface{}:
		switch tplData := tpl.(type) {
		case []interface{}:
			if len(inputData) != len(tplData) {
				err = fmt.Errorf("tag [%s] not match", tag)
				return
			}

			for idx, iv := range inputData {
				err = v.Validate(iv, tplData[idx], fmt.Sprintf("%s[%d]", tag, idx))
				if err != nil {
					return
				}
			}
		case string:
			return v.validateVar(inputData, tplData, tag)
		default:
			return fmt.Errorf("tag [%s] type inconsistent", tag)
		}
	case map[string]interface{}:
		switch tplData := tpl.(type) {
		case map[string]interface{}:
			for k, iv := range inputData {
				err = v.Validate(iv, tplData[k], k)
				if err != nil {
					return
				}
			}
		default:
			return fmt.Errorf("tag [%s] type inconsistent", tag)
		}
	default:
		switch tplData := tpl.(type) {
		case string:
			return v.validateVar(inputData, tplData, tag)
		default:
			return fmt.Errorf("tag [%s] type inconsistent", tag)
		}
	}
	return
}

func (v *Validator) validateVar(val interface{}, tpl, tag string) (err error) {
	ss := strings.Split(tpl, tokenSeparator)
	if len(ss) != 2 {
		return
	}
	token := ss[0]
	fieldTag := ss[1]

	defer func() {
		if err != nil {
			err = fmt.Errorf(strings.Replace(err.Error(), "''", fmt.Sprintf("'%s'", tag), -1))
		}
	}()

	err = checkKind(token, val, tag)
	if err != nil {
		return
	}

	switch token {
	case tokenInt, tokenFloat, tokenString, tokenList:
		err = v.vd.Var(val, fieldTag)
	case tokenRe:
		exp := regexp.MustCompile(fieldTag)
		match := exp.Match([]byte(fmt.Sprint(val)))
		if !match {
			err = fmt.Errorf(typeErrMsg, tag, fieldTag, fmt.Sprint(val))
		}
	case tokenOutput:
		v.cache[fieldTag] = val
	}
	return
}

func checkKind(token string, val interface{}, tag string) error {
	invalid := false
	switch token {
	case tokenInt, tokenFloat:
		if reflect.TypeOf(val).Kind() != reflect.Float64 {
			invalid = true
		}
	case tokenString:
		if reflect.TypeOf(val).Kind() != reflect.String {
			invalid = true
		}
	}
	if invalid {
		return fmt.Errorf(typeErrMsg, tag, token, reflect.TypeOf(val))
	}
	return nil
}
