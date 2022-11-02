package app

import (
	"errors"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/inference"
)

type InferenceIndex = domain.InferenceIndex

type InferenceCreateCmd domain.Inference

func (cmd *InferenceCreateCmd) Validate() error {
	index := &cmd.InferenceIndex

	b := index.Project.Id == "" ||
		index.Project.Owner == nil ||
		index.Id == "" ||
		cmd.LastCommit == "" ||
		cmd.ProjectName == nil ||
		cmd.UserToken == "" ||
		cmd.SurvivalTime <= 0

	if b {
		return errors.New("invalid cmd")
	}

	return nil
}

func (cmd *InferenceCreateCmd) toInference() *domain.Inference {
	return (*domain.Inference)(cmd)
}

type InferenceUpdateCmd struct {
	domain.InferenceIndex

	TimeToExtend int
}

func (cmd *InferenceUpdateCmd) Validate() error {
	index := &cmd.InferenceIndex

	b := index.Project.Id == "" ||
		index.Project.Owner == nil ||
		index.Id == "" ||
		cmd.TimeToExtend <= 0

	if b {
		return errors.New("invalid cmd")
	}

	return nil
}

type InferenceService interface {
	Create(*InferenceCreateCmd) error
	ExtendSurvivalTime(*InferenceUpdateCmd) error
}

func NewInferenceService(
	manager inference.Inference,
) InferenceService {
	return inferenceService{
		manager: manager,
	}
}

type inferenceService struct {
	manager inference.Inference
}

func (s inferenceService) Create(cmd *InferenceCreateCmd) error {
	return s.manager.Create(cmd.toInference())
}

func (s inferenceService) ExtendSurvivalTime(cmd *InferenceUpdateCmd) error {
	return s.manager.ExtendSurvivalTime(&cmd.InferenceIndex, cmd.TimeToExtend)
}
