package domain

import "strings"

type CustomEvaluate struct {
	EvaluateIndex
	AimPath string
}

func (e *CustomEvaluate) Type() string {
	return "custom"
}

type StandardEvaluate struct {
	EvaluateIndex

	LogPath           string
	MomentumScope     EvaluateScope
	BatchSizeScope    EvaluateScope
	LearningRateScope EvaluateScope
}

func (e *StandardEvaluate) Type() string {
	return "standard"
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
