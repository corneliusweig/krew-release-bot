package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/release/source/actions"
	"github.com/rajatjindal/krew-release-bot/pkg/releaser"
	"github.com/sirupsen/logrus"
)

func main() {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		err := actions.RunAction()
		if err != nil {
			logrus.Fatal(err)
		}

		os.Exit(0)
	}

	ghToken := os.Getenv("GH_TOKEN")
	releaser := &releaser.Releaser{
		Token:           ghToken,
		TokenEmail:      "krewpluginreleasebot@gmail.com",
		TokenUserHandle: "krew-release-bot",
		TokenUsername:   "Krew Release Bot",
	}

	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8080),
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/github-webhook", releaser.HandleGithubWebhook)
	http.HandleFunc("/github-action-webhook", releaser.HandleActionWebhook)

	logrus.Fatal(s.ListenAndServe())
}
