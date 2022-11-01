package controller

import (
	"github.com/opensourceways/xihe-inference-evaluate/app"
	"github.com/opensourceways/xihe-inference-evaluate/domain"
)

type EvaluateIndex struct {
	User       string `json:"user"`
	ProjectId  string `json:"project_id"`
	EvaluateId string `json:"evaluate_id"`
	TrainingId string `json:"training_id"`
}

func (req *EvaluateIndex) toIndex() (index app.EvaluateIndex, err error) {
	if index.Project.Owner, err = domain.NewAccount(req.User); err != nil {
		return
	}

	index.Id = req.EvaluateId
	index.Project.Id = req.ProjectId
	index.TrainingId = req.TrainingId

	return
}

type CustomEvaluateCreateRequest struct {
	EvaluateIndex

	AimPath string `json:"aim_path"`
}

func (req *CustomEvaluateCreateRequest) toCmd() (
	cmd app.CustomEvaluateCreateCmd, err error,
) {
	if cmd.EvaluateIndex, err = req.EvaluateIndex.toIndex(); err != nil {
		return
	}

	cmd.AimPath = req.AimPath

	err = cmd.Validate()

	return
}

type StandardEvaluateCreateRequest struct {
	EvaluateIndex

	LogPath string `json:"log_path"`

	MomentumScope     domain.EvaluateScope `json:"momentum_scope"`
	BatchSizeScope    domain.EvaluateScope `json:"batch_size_scope"`
	LearningRateScope domain.EvaluateScope `json:"learning_rate_scope"`
}

func (req *StandardEvaluateCreateRequest) toCmd() (
	cmd app.StandardEvaluateCreateCmd, err error,
) {
	if cmd.EvaluateIndex, err = req.EvaluateIndex.toIndex(); err != nil {
		return
	}

	cmd.LogPath = req.LogPath
	cmd.MomentumScope = req.MomentumScope
	cmd.BatchSizeScope = req.BatchSizeScope
	cmd.LearningRateScope = req.LearningRateScope

	err = cmd.Validate()

	return
}

type EvaluateUpdateRequest struct {
	EvaluateIndex

	Expiry int64 `json:"expiry"`
}

func (req *EvaluateUpdateRequest) toCmd() (cmd app.EvaluateUpdateCmd, err error) {
	if cmd.EvaluateIndex, err = req.EvaluateIndex.toIndex(); err != nil {
		return
	}

	cmd.Expiry = req.Expiry

	err = cmd.Validate()

	return
}
