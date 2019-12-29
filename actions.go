package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/helpers"
	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/sirupsen/logrus"
)

//HandleAction handles the event from gh-actions
func HandleAction(w http.ResponseWriter, r *http.Request) {
	ok, err := isTokenValid(r)
	if err != nil {
		logrus.Error(err)
		http.Error(w, fmt.Sprintf("failed to validate token. err: %v", err), http.StatusForbidden)
	}

	if !ok {
		http.Error(w, "failed to validate token", http.StatusForbidden)
	}

	//TODO: somehow get action data based on request body
	//then remaining code will be same
	actionData, err := realAction.GetActionData(nil)
	if err != nil {
		logrus.Error("failed to get actionData. error: ", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to get actionData."))
		return
	}

	if actionData.ReleaseInfo.GetPrerelease() {
		logrus.Infof("%s is a pre-release. not opening the PR", actionData.ReleaseInfo.GetTagName())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("its a prerelease."))
		return
	}

	tempdir, err := ioutil.TempDir("", "krew-index-")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create tempdir."))
		return
	}

	logrus.Infof("will operate in tempdir %s", tempdir)
	repo, err := helpers.CloneRepos(actionData, tempdir)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to clone the repo."))
		return
	}

	//https://raw.githubusercontent.com/rajatjindal/kubectl-modify-secret/master/.krew.yaml
	templateFileURI := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/master/.krew.yaml", actionData.RepoOwner, actionData.Repo)
	tempfile, err := ioutil.TempFile("", "krew-")
	if err != nil {
		logrus.Info("failed to create temp file")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create temp file."))
		return
	}
	defer os.Remove(tempfile.Name())

	err = krew.UpdatePluginManifest(templateFileURI, tempfile.Name(), actionData.ReleaseInfo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	actionData.Inputs.PluginName, err = krew.GetPluginName(tempfile.Name())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	logrus.Info("validating ownership")
	actualFile := filepath.Join(tempdir, "plugins", krew.PluginFileName(actionData.Inputs.PluginName))
	err = krew.ValidateOwnership(actualFile, actionData.RepoOwner)
	if err != nil {
		logrus.Errorf("failed when validating ownership with error: %s", err.Error())
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("could not verify ownership of plugin."))
		return
	}

	logrus.Info("update plugin manifest with latest release info")
	err = krew.ValidatePlugin(actionData.Inputs.PluginName, tempfile.Name())
	if err != nil {
		logrus.Errorf("failed when validating plugin spec with error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("validate spec failed"))
		return
	}

	_, err = copy(tempfile.Name(), actualFile)
	if err != nil {
		logrus.Errorf("failed when copying plugin spec with error: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("copy spec file failed"))
		return
	}

	logrus.Infof("pushing changes to branch %s", actionData.ReleaseInfo.GetTagName())
	commit := helpers.Commit{
		Msg:        fmt.Sprintf("new version %s of %s", actionData.ReleaseInfo.GetTagName(), actionData.Inputs.PluginName),
		RemoteName: helpers.OriginNameLocal,
	}
	err = helpers.AddCommitAndPush(repo, commit, actionData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	logrus.Info("submitting the pr")
	pr, err := helpers.SubmitPR(actionData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pr))
}

type injectAuth struct {
	token string
}

func (ij *injectAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", ij.token))
	return http.DefaultTransport.RoundTrip(req)
}

func isTokenValid(r *http.Request) (bool, error) {
	ok, err := isAppTokenValid(r)
	if err == nil && ok {
		return ok, err
	}

	logrus.Info("not a valid app token, falling back on checking if its user token")
	//TODO: there should be a better way to check if its a app-token or user-token

	return isUserTokenValid(r)
}

func isUserTokenValid(r *http.Request) (bool, error) {
	return false, fmt.Errorf("user token validation not supported yet")
}

func isAppTokenValid(r *http.Request) (bool, error) {
	token := r.Header.Get("x-github-token")
	if token == "" {
		return false, fmt.Errorf("token is empty")
	}

	mc := &http.Client{
		Transport: &injectAuth{token: token},
	}
	client := github.NewClient(mc)

	repos, _, err := client.Apps.ListRepos(context.TODO(), &github.ListOptions{})
	if err != nil {
		return false, err
	}

	if len(repos) == 0 {
		return false, fmt.Errorf("no repos can be accessed using this token")
	}

	//TODO: really compare the owner with the homepage of what we have received in request
	return true, nil
}
