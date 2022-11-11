package domain

import (
	"errors"
	"regexp"
	"strings"
)

var (
	reName = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
)

// Account
type Account interface {
	Account() string
}

func NewAccount(v string) (Account, error) {
	if v == "" || strings.ToLower(v) == "root" || !reName.MatchString(v) {
		return nil, errors.New("invalid user name")
	}

	return dpAccount(v), nil
}

type dpAccount string

func (r dpAccount) Account() string {
	return string(r)
}

// ProjectName
type ProjectName interface {
	ProjectName() string
}

func NewProjectName(v string) (ProjectName, error) {
	if v == "" {
		return nil, errors.New("invalid project name")
	}

	return projectName(v), nil
}

type projectName string

func (r projectName) ProjectName() string {
	return string(r)
}
