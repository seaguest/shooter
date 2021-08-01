package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"shooter/constant"
	"shooter/json"
	"shooter/service"
	"strconv"
	"strings"
	"time"

	"github.com/seaguest/log"
)

func main() {
	scenarioPath := flag.String("dir", "../scenario", "scenario path")
	server := flag.String("server", "http://127.0.0.1:10001", "server")
	flag.Parse()

	s, err := os.Stat(*scenarioPath)
	if err != nil {
		log.Error(err)
		return
	}

	if !s.IsDir() {
		log.Error("scenario dir is not a folder")
		return
	}

	// 遍历根目录
	fileList, err := ioutil.ReadDir(*scenarioPath)
	if err != nil {
		log.Error(err)
		return
	}

	for _, file := range fileList {
		if !file.IsDir() {
			continue
		}

		// test case
		go func(dir string) {
			scenario := &service.Scenario{}
			scenario.Server = *server
			var inputTmp []string
			outputTmp := make(map[string]bool)

			// 遍历根目录
			scenariosSteps, err := ioutil.ReadDir(dir)
			if err != nil {
				log.Error(err)
				return
			}
			for _, step := range scenariosSteps {
				if step.Name() == constant.CommonInputFile {
					scenario.Common = filepath.Join(dir, step.Name())
				} else if strings.HasSuffix(step.Name(), constant.SuffixInput) {
					inputTmp = append(inputTmp, filepath.Join(dir, step.Name()))
				} else if strings.HasSuffix(step.Name(), constant.SuffixOutput) {
					outputTmp[filepath.Join(dir, step.Name())] = true
				}
			}

			for _, n := range inputTmp {
				out := strings.Replace(n, constant.SuffixInput, constant.SuffixOutput, -1)
				_, ok := outputTmp[out]
				if !ok {
					log.Error(fmt.Sprintf("[%s] output template missing", n))
				}
			}

			scenario.Input = make([]string, len(inputTmp))
			scenario.Output = make([]string, len(inputTmp))

			for _, n := range inputTmp {
				ss := strings.Split(filepath.Base(n), "_")
				if len(ss) < 2 {
					log.Error(fmt.Sprintf("[%s] name invalid", n))
				}
				idx, _ := strconv.Atoi(ss[0])

				if idx-1 >= len(inputTmp) {
					log.Error(fmt.Sprintf("[%s] step number invalid", filepath.Base(n)))
					return
				}

				scenario.Input[idx-1] = n
				out := strings.Replace(n, constant.SuffixInput, constant.SuffixOutput, -1)
				scenario.Output[idx-1] = out
			}

			jv := json.NewValidator()
			for {
				service.ValidateScenario(scenario, jv)
				time.Sleep(time.Second * 5)
			}
		}(filepath.Join(*scenarioPath, file.Name()))
	}

	for {
		time.Sleep(time.Minute)
	}
}
