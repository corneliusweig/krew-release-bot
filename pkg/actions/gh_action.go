package actions

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

//GithubAction is the github action
type GithubAction struct{}

//GetActionData returns action data
func (r GithubAction) GetActionData(releaseInfo *github.RepositoryRelease) (ActionData, error) {
	owner, repo := GetOwnerAndRepo()

	return ActionData{
		Actor:       getActor(),
		Repo:        repo,
		RepoOwner:   owner,
		ReleaseInfo: releaseInfo,
	}, nil
}

//GetRepoURL returns clone url
func GetRepoURL(owner, repo string) string {
	return fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
}

//getInputForAction gets input to action
func (r GithubAction) getInputForAction(key string) string {
	return os.Getenv(fmt.Sprintf("INPUT_%s", strings.ToUpper(key)))
}

//GetPayload reads payload and returns it
func (r GithubAction) getPayload() ([]byte, error) {
	eventJSONPath := os.Getenv("GITHUB_EVENT_PATH")
	data, err := ioutil.ReadFile(eventJSONPath)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//GetOwnerAndRepo gets the owner and repo from the env
func GetOwnerAndRepo() (string, string) {
	s := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	return s[0], s[1]
}

func getActor() string {
	return os.Getenv("GITHUB_ACTOR")
}
