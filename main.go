package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rajatjindal/krew-release-bot/pkg/actions"
	"github.com/sirupsen/logrus"
)

var realAction actions.RealAction

func main() {
	ghToken := os.Getenv("GH_TOKEN")
	webhookToken := os.Getenv("WEBHOOK_TOKEN")

	realAction = actions.RealAction{
		Token:           ghToken,
		WebhookSecret:   webhookToken,
		TokenEmail:      "krewpluginreleasebot@gmail.com",
		TokenUserHandle: "krew-release-bot",
		TokenUsername:   "Krew Release Bot",
	}

	logrus.Infof("user: %s, name: %q", realAction.TokenUserHandle, realAction.TokenUsername)
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", 8080),
		MaxHeaderBytes: 1 << 20,
	}

	http.HandleFunc("/", Handle)
	logrus.Fatal(s.ListenAndServe())
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
