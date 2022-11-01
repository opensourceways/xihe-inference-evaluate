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
		cmd.UserToken == ""

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

	Expiry int64
}

func (cmd *InferenceUpdateCmd) Validate() error {
	index := &cmd.InferenceIndex

	b := index.Project.Id == "" ||
		index.Project.Owner == nil ||
		index.Id == "" ||
		cmd.Expiry <= 0

	if b {
		return errors.New("invalid cmd")
	}

	return nil
}

type InferenceService interface {
	Create(*InferenceCreateCmd) error
	ExtendExpiry(*InferenceUpdateCmd) error
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

func (s inferenceService) ExtendExpiry(cmd *InferenceUpdateCmd) error {
	return s.manager.ExtendExpiry(&cmd.InferenceIndex, cmd.Expiry)
}
