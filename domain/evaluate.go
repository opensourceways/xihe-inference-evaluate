package domain

import "strings"

const (
	EvaluateTypeCustom   = "custom"
	EvaluateTypeStandard = "standard"
)

type CustomEvaluate struct {
	EvaluateIndex
	AimPath string
}

func (e *CustomEvaluate) Type() string {
	return EvaluateTypeCustom
}

type StandardEvaluate struct {
	EvaluateIndex

	LogPath           string
	MomentumScope     EvaluateScope
	BatchSizeScope    EvaluateScope
	LearningRateScope EvaluateScope
}

func (e *StandardEvaluate) Type() string {
	return EvaluateTypeStandard
}

type EvaluateScope []string

func (s EvaluateScope) String() string {
	if len(s) == 0 {
		return ""
	}

	return "[" + strings.Join(([]string)(s), ", ") + "]"
}

type EvaluateIndex struct {
	Project    ResourceIndex
	TrainingId string
	Id         string
}
