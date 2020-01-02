package releaser

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/rajatjindal/krew-release-bot/pkg/release"
	"github.com/rajatjindal/krew-release-bot/pkg/release/source/actions"
	"github.com/rajatjindal/krew-release-bot/pkg/release/source/webhook"
	"github.com/sirupsen/logrus"
)

//Releaser is what opens PR
type Releaser struct {
	Token                         string
	TokenEmail                    string
	TokenUserHandle               string
	TokenUsername                 string
	UpstreamKrewIndexRepo         string
	UpstreamKrewIndexRepoOwner    string
	UpstreamKrewIndexRepoCloneURL string
	LocalKrewIndexRepo            string
	LocalKrewIndexRepoOwner       string
	LocalKrewIndexRepoCloneURL    string
}

//HandleGithubWebhook handles github webhook requests
func (releaser *Releaser) HandleGithubWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := webhook.NewGithubWebhook(os.Getenv("WEBHOOK_TOKEN"))
	if err != nil {
		http.Error(w, "failed to process", http.StatusInternalServerError)
		return
	}

	releaseRequest, err := hook.Parse(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting release request").Error(), http.StatusInternalServerError)
		return
	}

	pr, err := releaser.Release(releaseRequest)
	if err != nil {
		http.Error(w, errors.Wrap(err, "opening pr").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("PR %q submitted successfully", pr)))
}

//HandleActionWebhook handles requests from github actions
func (releaser *Releaser) HandleActionWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := actions.NewGithubActions()
	if err != nil {
		http.Error(w, errors.Wrap(err, "creating instance of action handler").Error(), http.StatusInternalServerError)
		return
	}

	releaseRequest, err := hook.Parse(r)
	if err != nil {
		http.Error(w, errors.Wrap(err, "getting release request").Error(), http.StatusInternalServerError)
		return
	}

	pr, err := releaser.Release(releaseRequest)
	if err != nil {
		http.Error(w, errors.Wrap(err, "opening pr").Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("PR %q submitted successfully", pr)))
}

//Release releases
func (releaser *Releaser) Release(request *release.Request) (string, error) {
	tempdir, err := ioutil.TempDir("", "krew-index-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tempdir)

	logrus.Infof("will operate in tempdir %s", tempdir)
	repo, err := releaser.cloneRepos(tempdir, request)
	if err != nil {
		return "", err
	}

	newIndexFile, err := ioutil.TempFile("", "krew-")
	if err != nil {
		return "", err
	}
	defer os.Remove(newIndexFile.Name())

	err = ioutil.WriteFile(newIndexFile.Name(), request.ProcessedTemplate, 0644)
	if err != nil {
		return "", err
	}

	logrus.Info("validating ownership")
	existingIndexFile := filepath.Join(tempdir, "plugins", krew.PluginFileName(request.PluginName))
	err = krew.ValidateOwnership(existingIndexFile, request.PluginOwner)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("plugin %q not found in existing repo. The first release of a new plugin has to be done manually", request.PluginName)
		}

		return "", fmt.Errorf("failed when validating ownership with error: %s", err.Error())
	}

	logrus.Info("update plugin manifest with latest release info")
	err = krew.ValidatePlugin(request.PluginName, newIndexFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed when validating plugin spec with error: %s", err.Error())
	}

	_, err = copyFile(newIndexFile.Name(), existingIndexFile)
	if err != nil {
		return "", fmt.Errorf("failed when copying plugin spec with error: %s", err.Error())
	}

	logrus.Infof("pushing changes to branch %s", *releaser.getBranchName(request))
	commit := commitConfig{
		Msg:        fmt.Sprintf("new version %s of %s", request.TagName, request.PluginName),
		RemoteName: OriginNameLocal,
	}

	err = releaser.addCommitAndPush(repo, commit, request)
	if err != nil {
		return "", err
	}

	logrus.Info("submitting the pr")
	pr, err := releaser.submitPR(request)
	if err != nil {
		return "", err
	}

	return pr, nil
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
