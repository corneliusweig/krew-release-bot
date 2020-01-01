package krew

import "os"

//GetKrewIndexRepoName returns the krew-index repo name
func GetKrewIndexRepoName() string {
	override := os.Getenv("upstream-krew-index-repo-name")
	if override != "" {
		return override
	}

	return "krew-index"
}

//GetKrewIndexRepoOwner returns the krew-index repo owner
func GetKrewIndexRepoOwner() string {
	override := os.Getenv("upstream-krew-index-repo-owner")
	if override != "" {
		return override
	}

	return "rajatjin"
}
