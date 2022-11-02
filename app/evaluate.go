package app

import (
	"errors"

	"github.com/opensourceways/xihe-inference-evaluate/domain"
	"github.com/opensourceways/xihe-inference-evaluate/domain/evaluate"
)

type EvaluateIndex = domain.EvaluateIndex

type CustomEvaluateCreateCmd domain.CustomEvaluate

func (cmd *CustomEvaluateCreateCmd) Validate() error {
	index := &cmd.EvaluateIndex

	b := index.Project.Id == "" ||
		index.Project.Owner == nil ||
		index.TrainingId == "" ||
		index.Id == "" ||
		cmd.AimPath == "" ||
		cmd.SurvivalTime <= 0

	if b {
		return errors.New("invalid cmd")
	}

	return nil
}

func (cmd *CustomEvaluateCreateCmd) toEvaluate() *domain.CustomEvaluate {
	return (*domain.CustomEvaluate)(cmd)
}

// standard
type StandardEvaluateCreateCmd domain.StandardEvaluate

func (cmd *StandardEvaluateCreateCmd) Validate() error {
	index := &cmd.EvaluateIndex

	b := index.Project.Id == "" ||
		index.Project.Owner == nil ||
		index.TrainingId == "" ||
		index.Id == "" ||
		cmd.LogPath == "" ||
		cmd.SurvivalTime <= 0

	if b {
		return errors.New("invalid cmd")
	}

	return nil
}

func (cmd *StandardEvaluateCreateCmd) toEvaluate() *domain.StandardEvaluate {
	return (*domain.StandardEvaluate)(cmd)
}

type EvaluateService interface {
	CreateCustom(*CustomEvaluateCreateCmd) error
	CreateStandard(*StandardEvaluateCreateCmd) error
}

func NewEvaluateService(
	manager evaluate.Evaluate,
) EvaluateService {
	return evaluateService{
		manager: manager,
	}
}

type evaluateService struct {
	manager evaluate.Evaluate
}

func (s evaluateService) CreateCustom(cmd *CustomEvaluateCreateCmd) error {
	return s.manager.CreateCustom(cmd.toEvaluate())
}

func (s evaluateService) CreateStandard(cmd *StandardEvaluateCreateCmd) error {
	return s.manager.CreateStandard(cmd.toEvaluate())
}
