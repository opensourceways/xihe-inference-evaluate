package sdk

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/opensourceways/community-robot-lib/utils"

	"github.com/opensourceways/xihe-inference-evaluate/controller"
)

type InferenceCreateOption = controller.InferenceCreateRequest
type InferenceUpdateOption = controller.InferenceUpdateRequest

type EvaluateUpdateOption = controller.EvaluateUpdateRequest
type CustomEvaluateCreateOption = controller.CustomEvaluateCreateRequest
type StandardEvaluateCreateOption = controller.StandardEvaluateCreateRequest

func NewInferenceEvaluate(endpoint string) InferenceEvaluate {
	return InferenceEvaluate{
		endpoint: strings.TrimSuffix(endpoint, "/"),
		cli:      utils.NewHttpClient(3),
	}
}

type InferenceEvaluate struct {
	endpoint string
	cli      utils.HttpClient
}

func (t InferenceEvaluate) inferenceURL() string {
	return fmt.Sprintf("%s/api/v1/inference/project", t.endpoint)
}

func (t InferenceEvaluate) CreateInference(opt *InferenceCreateOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, t.inferenceURL(), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return t.forwardTo(req, nil)
}

func (t InferenceEvaluate) ExtendExpiryOfInference(opt *InferenceUpdateOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, t.inferenceURL(), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return t.forwardTo(req, nil)
}

// evaluate
func (t InferenceEvaluate) evaluateURL() string {
	return fmt.Sprintf("%s/api/v1/evaluate/project", t.endpoint)
}

func (t InferenceEvaluate) CreateCustomEvaluate(opt *CustomEvaluateCreateOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, t.evaluateURL()+"/custom", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return t.forwardTo(req, nil)
}

func (t InferenceEvaluate) CreateStandardEvaluate(opt *StandardEvaluateCreateOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, t.evaluateURL()+"/standard", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return t.forwardTo(req, nil)
}

func (t InferenceEvaluate) ExtendExpiryOfEvaluate(opt *EvaluateUpdateOption) error {
	payload, err := utils.JsonMarshal(&opt)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, t.evaluateURL(), bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	return t.forwardTo(req, nil)
}

func (t InferenceEvaluate) forwardTo(req *http.Request, jsonResp interface{}) (err error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "xihe-inference-evaluate")

	if jsonResp != nil {
		v := struct {
			Data interface{} `json:"data"`
		}{jsonResp}

		_, err = t.cli.ForwardTo(req, &v)
	} else {
		_, err = t.cli.ForwardTo(req, jsonResp)
	}

	return
}
