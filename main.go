package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/release"
	"github.com/rajatjindal/krew-release-bot/pkg/release/source/actions"
	"github.com/rajatjindal/krew-release-bot/pkg/release/source/webhook"
	"github.com/sirupsen/logrus"
)

//Releaser is what opens PR
type Releaser struct {
	Token           string
	TokenEmail      string
	TokenUserHandle string
	TokenUsername   string
}

func main() {
	ghToken := os.Getenv("GH_TOKEN")

	if os.Getenv("GITHUB_ACTIONS") == "true" {
		//runaction
	}

	releaser := &Releaser{
		Token:           ghToken,
		TokenEmail:      "krewpluginreleasebot@gmail.com",
		TokenUserHandle: "krew-release-bot",
		TokenUsername:   "Krew Release Bot",
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8080),
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/webhook", releaser.HandleWebhook)
	http.HandleFunc("/gh-action-webhook", releaser.HandleActionWebhook)

	logrus.Fatal(s.ListenAndServe())
}

//HandleWebhook handles github webhook requests
func (releaser *Releaser) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := webhook.NewGithubWebhook(os.Getenv("WEBHOOK_TOKEN"))
	if err != nil {
		http.Error(w, "failed to process", http.StatusInternalServerError)
		return
	}

	releaseRequest, err := hook.Parse(r)
	if err != nil {
		http.Error(w, "failed to get release request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

//HandleActionWebhook handles requests from github actions
func (releaser *Releaser) HandleActionWebhook(w http.ResponseWriter, r *http.Request) {
	hook, err := actions.NewGithubActions()
	if err != nil {
		http.Error(w, "failed to process", http.StatusInternalServerError)
		return
	}

	releaseRequest, err := hook.Parse(r)
	if err != nil {
		http.Error(w, "failed to get release request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

//Release releases
func (releaser *Releaser) Release(request *release.Request) error {

}

func copy(src, dst string) (int64, error) {
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
