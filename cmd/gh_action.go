package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/go-resty/resty"
	"github.com/google/go-github/github"
	"github.com/rajatjindal/krew-release-bot/pkg/actions"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// actionCommand represents the base command when called without any subcommands
var actionCommand = &cobra.Command{
	Use:   "actionCommand",
	Short: "generates the .krew file and open PR",
	Run: func(cmd *cobra.Command, args []string) {
		processAction()
	},
}

func init() {
	rootCmd.AddCommand(actionCommand)
}

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
		Post("https://rajatjindal.o6s.io/echo-server")

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
	fmt.Println("here")
	return http.DefaultTransport.RoundTrip(req)
}
