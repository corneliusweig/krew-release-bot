package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/go-resty/resty"
	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/actions"
	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/sirupsen/logrus"
)

func processAction() error {
	tag, err := getTag()
	if err != nil {
		return err
	}

	mc := &http.Client{
		Transport: &injectAuth{token: os.Getenv("GITHUB_TOKEN")},
	}
	client := github.NewClient(mc)

	release, err := getReleaseForTag(client, tag)
	if err != nil {
		return err
	}

	if len(release.Assets) == 0 {
		return fmt.Errorf("no assets found for release with tag %q", tag)
	}

	ghaction := actions.GithubAction{}
	actionData, err := ghaction.GetActionData(release)
	if err != nil {
		return err
	}

	pluginManifest, err := processTemplate(actionData)
	if err != nil {
		return err
	}

	d := &OverTheWire{
		ActionData:     actionData,
		PluginManifest: string(pluginManifest),
	}

	err = submitForPR(d)
	if err != nil {
		return err
	}

	return nil
}

//OverTheWire sends request to open PR over the wire
type OverTheWire struct {
	ActionData     actions.ActionData `json:"action-data"`
	PluginManifest string             `json:"plugin-manifest"`
}

func submitForPR(data *OverTheWire) error {
	client := resty.New()
	resp, err := client.R().
		SetBody(data).
		SetHeader("x-github-token", os.Getenv("GITHUB_TOKEN")).
		Post("https://krew-bot-github-actions-3gfdrquxea-uc.a.run.app/gh-action")

	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusCreated {
		return fmt.Errorf("expected code %d, got %d", http.StatusCreated, resp.StatusCode())
	}

	logrus.Info(string(resp.Body()))
	return nil
}

func getTag() (string, error) {
	ref := os.Getenv("GITHUB_REF")
	if ref == "" {
		return "", fmt.Errorf("GITHUB_REF env variable not found")
	}

	//GITHUB_REF=refs/tags/v0.0.6
	if !strings.HasPrefix(ref, "refs/tags/") {
		return "", fmt.Errorf("GITHUB_REF expected to be of format refs/tags/<tag> but found %q", ref)
	}

	return strings.ReplaceAll(ref, "refs/tags/", ""), nil
}

func getReleaseForTag(client *github.Client, tag string) (*github.RepositoryRelease, error) {
	owner, repo := actions.GetOwnerAndRepo()
	release, _, err := client.Repositories.GetReleaseByTag(context.TODO(), owner, repo, tag)
	if err != nil {
		return nil, err
	}

	return release, nil
}

type injectAuth struct {
	token string
}

func (ij *injectAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", ij.token))
	x, _ := httputil.DumpRequestOut(req, true)
	fmt.Println(string(x))
	return http.DefaultTransport.RoundTrip(req)
}

//HandleActionWebhook is webhook to handle request from github actions
func HandleActionWebhook(w http.ResponseWriter, r *http.Request) {
	otw, err := validateTokenAndGetBody(r)
	if err != nil {
		http.Error(w, "invalid token", http.StatusForbidden)
		return
	}

	ghToken := os.Getenv("GH_TOKEN")
	pluginName := otw.ActionData.Inputs.PluginName
	otw.ActionData.Inputs = actions.Inputs{
		PluginName:                pluginName,
		Token:                     ghToken,
		TokenUserEmail:            "krewpluginreleasebot@gmail.com",
		TokenUserHandle:           "krew-release-bot",
		TokenUserName:             "Krew Release Bot",
		UpstreamKrewIndexOwner:    krew.GetKrewIndexRepoOwner(),
		UpstreamKrewIndexRepoName: krew.GetKrewIndexRepoName(),
	}

	otw.ActionData.Derived = actions.Derived{
		LocalCloneURL:    actions.GetRepoURL("krew-release-bot", "krew-index"),
		UpstreamCloneURL: actions.GetRepoURL(krew.GetKrewIndexRepoOwner(), krew.GetKrewIndexRepoName()),
	}

	pr, err := processMe(otw.ActionData, []byte(otw.PluginManifest))
	if err != nil {
		logrus.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed when running processMe"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(pr))
}

func validateTokenAndGetBody(r *http.Request) (*OverTheWire, error) {
	token := r.Header.Get("x-github-token")
	if token == "" {
		return nil, fmt.Errorf("token not found")
	}

	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	otw := &OverTheWire{}
	err = json.Unmarshal(b, otw)
	if err != nil {
		return nil, err
	}

	mc := &http.Client{
		Transport: &injectAuth{token: token},
	}
	client := github.NewClient(mc)

	repos, _, err := client.Apps.ListRepos(context.TODO(), &github.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, fmt.Errorf("no repos can be accessed using this token")
	}

	var found = false
	for _, repo := range repos {
		if repo.GetOwner().GetLogin() == otw.ActionData.RepoOwner && repo.GetName() == otw.ActionData.Repo {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("token not valid for this repo")
	}

	return otw, nil
}
