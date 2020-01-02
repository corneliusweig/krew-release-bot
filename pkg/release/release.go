package release

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/rajatjindal/krew-release-bot/pkg/krew"
	"github.com/sirupsen/logrus"
)

//Request is the release request for new plugin
type Request struct {
	TagName            string `json:"tagName"`
	PluginName         string `json:"pluginName"`
	PluginRepo         string `json:"pluginRepo"`
	PluginOwner        string `json:"pluginOwner"`
	PluginReleaseActor string `json:"pluginReleaseActor"`
	TemplateFile       string `json:"templateFile"`
	ProcessedTemplate  []byte `json:"processedTemplate"`
}

//ProcessTemplate process the .krew.yaml template for the release request
func (r *Request) ProcessTemplate() error {
	t := template.New(".krew.yaml").Funcs(map[string]interface{}{
		"addURIAndSha": func(url, tag string) string {
			t := struct {
				TagName string
			}{
				TagName: tag,
			}
			buf := new(bytes.Buffer)
			temp, err := template.New("url").Parse(url)
			if err != nil {
				panic(err)
			}

			err = temp.Execute(buf, t)
			if err != nil {
				panic(err)
			}

			logrus.Infof("getting sha256 for %s", buf.String())
			sha256, err := getSha256ForAsset(buf.String())
			if err != nil {
				panic(err)
			}

			return fmt.Sprintf(`uri: %s
    sha256: %s`, buf.String(), sha256)
		},
	})

	templateObject, err := t.ParseFiles(r.TemplateFile)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = templateObject.Execute(buf, r)
	if err != nil {
		return err
	}

	r.ProcessedTemplate = buf.Bytes()
	r.PluginName, err = krew.GetPluginName(buf)
	if err != nil {
		return err
	}

	return nil
}
