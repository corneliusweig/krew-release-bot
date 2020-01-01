package actions

import (
	"net/http"

	"github.com/rajatjindal/krew-release-bot/pkg/release"
)

//GithubActions is github webhook handler
type GithubActions struct{}

//NewGithubActions gets new git webhook instance
func NewGithubActions() (*GithubActions, error) {
	return &GithubActions{}, nil
}

//Parse validates the request
func (w *GithubActions) Parse(r *http.Request) (*release.Request, error) {
	return nil, nil
}
