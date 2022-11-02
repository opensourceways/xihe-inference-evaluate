package controller

import (
	"github.com/opensourceways/xihe-inference-evaluate/app"
	"github.com/opensourceways/xihe-inference-evaluate/domain"
)

type InferenceIndex struct {
	User        string `json:"user"`
	ProjectId   string `json:"project_id"`
	InferenceId string `json:"inference_id"`
}

func (req *InferenceIndex) toIndex() (index app.InferenceIndex, err error) {
	if index.Project.Owner, err = domain.NewAccount(req.User); err != nil {
		return
	}

	index.Project.Id = req.ProjectId
	index.Id = req.InferenceId

	return
}

type InferenceCreateRequest struct {
	InferenceIndex

	UserToken    string `json:"token"`
	LastCommit   string `json:"last_commit"`
	ProjectName  string `json:"project_name"`
	SurvivalTime int    `json:"survival_time"`
}

func (req *InferenceCreateRequest) toCmd() (
	cmd app.InferenceCreateCmd, err error,
) {
	if cmd.InferenceIndex, err = req.InferenceIndex.toIndex(); err != nil {
		return
	}

	if cmd.ProjectName, err = domain.NewProjectName(req.ProjectName); err != nil {
		return
	}

	cmd.UserToken = req.UserToken
	cmd.LastCommit = req.LastCommit

	err = cmd.Validate()

	return
}

type InferenceUpdateRequest struct {
	InferenceIndex

	// TimeToExtend stands for the time in seconds to
	// extend the survival time of instance.
	TimeToExtend int `json:"time_to_extend"`
}

func (req *InferenceUpdateRequest) toCmd() (cmd app.InferenceUpdateCmd, err error) {
	if cmd.InferenceIndex, err = req.InferenceIndex.toIndex(); err != nil {
		return
	}

	cmd.TimeToExtend = req.TimeToExtend

	err = cmd.Validate()

	return
}
