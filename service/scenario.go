package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"shooter/constant"
	jsn "shooter/json"
	"strings"

	"github.com/imroc/req"
	"github.com/seaguest/log"
)

// contains common input data
type Scenario struct {
	Server string
	Common string
	Input  []string
	Output []string
}

type CommonInput struct {
	Headers map[string]interface{} `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

type Input struct {
	Url     string                 `json:"url"`
	Method  string                 `json:"method"`
	Headers map[string]interface{} `json:"headers"`
	Body    map[string]interface{} `json:"body"`
}

func ValidateScenario(scenario *Scenario, v *jsn.Validator) {
	cb, err := ioutil.ReadFile(scenario.Common)
	if err != nil {
		log.Error(err)
		return
	}

	var commonInput CommonInput
	if err = json.Unmarshal(cb, &commonInput); err != nil {
		log.Error(err)
		return
	}

	for idx, f := range scenario.Input {
		ib, err := ioutil.ReadFile(f)
		if err != nil {
			log.Error(err)
			return
		}

		var input Input
		if err = json.Unmarshal(ib, &input); err != nil {
			log.Error(err)
			return
		}

		ob, err := ioutil.ReadFile(scenario.Output[idx])
		if err != nil {
			log.Error(err)
			return
		}

		var output interface{}
		if err = json.Unmarshal(ob, &output); err != nil {
			log.Error(err)
			return
		}

		// override common by input
		jsn.Override(input, commonInput)

		input.Url = scenario.Server + input.Url

		filledHeaders := jsn.Fill(input.Headers, v.GetCache())
		filledBody := jsn.Fill(input.Body, v.GetCache())
		input.Headers = filledHeaders.(map[string]interface{})
		input.Body = filledBody.(map[string]interface{})

		// create log directory
		baseDir := filepath.Join(filepath.Dir(f), constant.LogDir)
		os.Mkdir(baseDir, 0755)

		ib, _ = json.MarshalIndent(input, "", "    ")
		_ = ioutil.WriteFile(filepath.Join(baseDir, filepath.Base(f)), ib, 0644)

		rsp, err := request(input)
		var respBody interface{}
		err = rsp.ToJSON(&respBody)
		if err != nil {
			log.Error(err)
			return
		}

		ib, _ = json.MarshalIndent(respBody, "", "    ")
		_ = ioutil.WriteFile(filepath.Join(baseDir, filepath.Base(scenario.Output[idx])), ib, 0644)

		log.Error("-------response--------")
		log.Error(rsp.ToString())

		err = v.Validate(respBody, output, "")
		if err != nil {
			log.Error(err)
			return
		}
	}
}

func convertHeaders(headers map[string]interface{}) map[string]string {
	nh := make(map[string]string)
	for k, v := range headers {
		nh[k] = fmt.Sprint(v)
	}
	return nh
}

func request(input Input) (rsp *req.Resp, err error) {
	switch strings.ToUpper(input.Method) {
	case http.MethodPost:
		rsp, err = req.Post(input.Url, req.Header(convertHeaders(input.Headers)), req.BodyJSON(input.Body))
		if err != nil {
			log.Error(err)
			return
		}
	case http.MethodGet:
		rsp, err = req.Get(input.Url, req.Header(convertHeaders(input.Headers)), req.Param(input.Body))
		if err != nil {
			log.Error(err)
			return
		}
	default:
		return
	}

	statusCode := rsp.Response().StatusCode
	if statusCode != http.StatusOK {
		err = fmt.Errorf("bad status code [%d]", statusCode)
		log.Error(err)
		return
	}
	return
}
