package actions

import (
	"github.com/google/go-github/github"
)

//Inputs is action inputs
type Inputs struct {
	PluginName                string
	Token                     string
	TokenUserHandle           string
	TokenUserEmail            string
	TokenUserName             string
	UpstreamKrewIndexRepoName string
	UpstreamKrewIndexOwner    string
}

//Derived is derived data
type Derived struct {
	UpstreamCloneURL string
	LocalCloneURL    string
}

//ActionData is action data
type ActionData struct {
	Workspace   string                    `json:"workspace"`
	Actor       string                    `json:"actor"`
	Repo        string                    `json:"repo"`
	RepoOwner   string                    `json:"repo-owner"`
	Inputs      Inputs                    `json:"inputs"`
	ReleaseInfo *github.RepositoryRelease `json:"release"`
	Derived     Derived                   `json:"derived"`
}
