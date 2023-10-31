package web

import (
	"fmt"
	"github.com/dtrunk90/switch-library-manager-web/db"
	"github.com/dtrunk90/switch-library-manager-web/pagination"
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"path/filepath"
	"strings"
)

func getType(gameFile *db.SwitchGameFiles) string {
	if gameFile.IsSplit {
		return "split"
	}

	if gameFile.MultiContent {
		return "multi-content"
	}

	ext := filepath.Ext(gameFile.File.ExtendedInfo.FileName)
	if len(ext) > 1 {
		return ext[1:]
	}

	return ""
}

func (web *Web) HandleIndex() {
	fsPatterns := []string {
		"resources/layout.html",
		"resources/partials/card.html",
		"resources/partials/filter.html",
		"resources/partials/pagination.html",
		"resources/pages/index.html",
	}

	web.HandleFiltered("/index.html", func(filter *TitleItemFilter) any {
		globalPageData.IsKeysFileAvailable = settings.IsKeysFileAvailable()
		globalPageData.Page = "index"
		items, p := web.getLibrary(filter)
		return TitleItemsPageData {
			GlobalPageData: globalPageData,
			TitleItems: items,
			Filter: filter,
			Pagination: p,
		}
	}, web.embedFS, fsPatterns...)
}

func (web *Web) getLibrary(filter *TitleItemFilter) ([]TitleItem, pagination.Pagination) {
	items := []TitleItem{}

	if web.state.localDB == nil {
		return items, pagination.Calculate(filter.Page, filter.PerPage, 0)
	}

	for k, v := range web.state.localDB.TitlesMap {
		if v.BaseExist {
			version := ""
			name := ""
			if v.File.Metadata.Ncap != nil {
				version = v.File.Metadata.Ncap.DisplayVersion
				name = v.File.Metadata.Ncap.TitleName["AmericanEnglish"].Title
			}

			if v.Updates != nil && len(v.Updates) != 0 {
				if v.Updates[v.LatestUpdate].Metadata.Ncap != nil {
					version = v.Updates[v.LatestUpdate].Metadata.Ncap.DisplayVersion
				} else {
					version = ""
				}
			}

			if title, ok := web.state.switchDB.TitlesMap[k]; ok {
				if title.Attributes.Name != "" {
					name = title.Attributes.Name
				}

				if filter.Keyword == "" || strings.Contains(strings.ToLower(v.File.Metadata.TitleId), strings.ToLower(filter.Keyword)) || strings.Contains(strings.ToLower(name), strings.ToLower(filter.Keyword)) {
					var imageUrl string
					if v.Icon != "" {
						imageUrl = "/i/" + v.Icon
					} else if v.Banner != "" {
						imageUrl = "/i/" + v.Banner
					}

					release, err := intToTime(title.Attributes.ReleaseDate)
					if err != nil {
						web.sugarLogger.Error(fmt.Errorf("parsing time failed: %w", err))
					}

					items = append(items, TitleItem {
						ImageUrl:    imageUrl,
						Id:          strings.ToUpper(v.File.Metadata.TitleId),
						LocalUpdate: v.LatestUpdate,
						Name:        name,
						Region:      title.Attributes.Region,
						ReleaseDate: release,
						Type:        strings.ToUpper(getType(v)),
						Version:     version,
					})
				}
			} else {
				if name == "" {
					name = db.ParseTitleNameFromFileName(v.File.ExtendedInfo.FileName)
				}

				if filter.Keyword == "" || strings.Contains(strings.ToLower(v.File.Metadata.TitleId), strings.ToLower(filter.Keyword)) || strings.Contains(strings.ToLower(name), strings.ToLower(filter.Keyword)) {
					items = append(items, TitleItem {
						Id:          strings.ToUpper(v.File.Metadata.TitleId),
						LocalUpdate: v.LatestUpdate,
						Name:        name,
						Type:        strings.ToUpper(getType(v)),
						Version:     version,
					})
				}
			}
		}
	}

	p := pagination.Calculate(filter.Page, filter.PerPage, len(items))

	if err := sortItems(filter, items); err != nil {
		web.sugarLogger.Error(err)
	}

	return items[p.Start:p.End], p
}
