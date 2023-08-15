package web

import (
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"path/filepath"
)

type Issue struct {
	File   string
	Reason string
}

type IssuesPageData struct {
	GlobalPageData
	Issues []Issue
}

func (web *Web) HandleIssues() {
	fsPatterns := []string {
		"resources/layout.html",
		"resources/pages/issues.html",
	}

	web.Handle("/issues.html", func() any {
		globalPageData.IsKeysFileAvailable = settings.IsKeysFileAvailable()
		globalPageData.Page = "issues"
		issues := web.getIssues()
		return IssuesPageData {
			GlobalPageData: globalPageData,
			Issues: issues,
		}
	}, web.embedFS, fsPatterns...)
}

func (web *Web) getIssues() []Issue {
	issues := []Issue{}

	if web.state.localDB == nil {
		return issues
	}

	for _, v := range web.state.localDB.TitlesMap {
		if !v.BaseExist {
			for _, update := range v.Updates {
				issues = append(issues, Issue{File: filepath.Join(update.ExtendedInfo.BaseFolder, update.ExtendedInfo.FileName), Reason: "base file is missing"})
			}

			for _, dlc := range v.Dlc {
				issues = append(issues, Issue{File: filepath.Join(dlc.ExtendedInfo.BaseFolder, dlc.ExtendedInfo.FileName), Reason: "base file is missing"})
			}
		}
	}

	for k, v := range web.state.localDB.Skipped {
		issues = append(issues, Issue{File: filepath.Join(k.BaseFolder, k.FileName), Reason: v.ReasonText})
	}

	return issues
}
