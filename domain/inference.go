package domain

type Inference struct {
	InferenceIndex
	ProjectName  ProjectName
	LastCommit   string
	UserToken    string
	SurvivalTime int // seconds
	User         string
}

type InferenceIndex struct {
	Project ResourceIndex
	Id      string
}

type ResourceIndex struct {
	Owner Account
	Id    string
}
