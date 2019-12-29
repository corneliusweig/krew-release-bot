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
	UpstreamCloneURL string `json:"upstream-clone-url,omitempty"`
	LocalCloneURL    string `json:"local-clone-url,omitempty"`
}

//ActionData is action data
type ActionData struct {
	Workspace   string                    `json:"workspace,omitempty"`
	Actor       string                    `json:"actor,omitempty"`
	Repo        string                    `json:"repo,omitempty"`
	RepoOwner   string                    `json:"repo-owner,omitempty"`
	Inputs      Inputs                    `json:"inputs,omitempty"`
	ReleaseInfo *github.RepositoryRelease `json:"release,omitempty"`
	Derived     Derived                   `json:"derived,omitempty"`
}
