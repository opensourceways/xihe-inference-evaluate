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

	UserToken   string `json:"token"`
	LastCommit  string `json:"last_commit"`
	ProjectName string `json:"project_name"`
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

	Expiry int64 `json:"expiry"`
}

func (req *InferenceUpdateRequest) toCmd() (cmd app.InferenceUpdateCmd, err error) {
	if cmd.InferenceIndex, err = req.InferenceIndex.toIndex(); err != nil {
		return
	}

	cmd.Expiry = req.Expiry

	err = cmd.Validate()

	return
}
