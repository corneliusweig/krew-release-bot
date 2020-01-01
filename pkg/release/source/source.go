package source

import (
	"net/http"

	"github.com/rajatjindal/krew-release-bot/release"
)

//Source is a release source interface
type Source interface {
	Parse(r *http.Request) (*release.Request, error)
}
