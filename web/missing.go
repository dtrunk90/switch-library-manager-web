package web

import (
	"fmt"
	"github.com/dtrunk90/switch-library-manager-web/pagination"
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"strings"
)

func (web *Web) HandleMissing() {
	fsPatterns := []string {
		"resources/layout.html",
		"resources/partials/card.html",
		"resources/partials/filter.html",
		"resources/partials/pagination.html",
		"resources/pages/missing.html",
	}

	web.HandleFiltered("/missing.html", func(filter *TitleItemFilter) any {
		globalPageData.IsKeysFileAvailable = settings.IsKeysFileAvailable()
		globalPageData.Page = "missing"
		items, p := web.getMissingGames(filter)
		return TitleItemsPageData {
			GlobalPageData: globalPageData,
			TitleItems: items,
			Filter: filter,
			Pagination: p,
		}
	}, web.embedFS, fsPatterns...)
}

func (web *Web) getMissingGames(filter *TitleItemFilter) ([]TitleItem, pagination.Pagination) {
	items := []TitleItem{}

	if web.state.localDB == nil {
		return items, pagination.Calculate(filter.Page, filter.PerPage, 0)
	}

	for k, v := range web.state.switchDB.TitlesMap {
		if _, ok := web.state.localDB.TitlesMap[k]; ok {
			continue
		}

		if v.Attributes.Name == "" || v.Attributes.Id == "" {
			continue
		}

		if filter.Keyword == "" || strings.Contains(strings.ToLower(v.Attributes.Id), strings.ToLower(filter.Keyword)) || strings.Contains(strings.ToLower(v.Attributes.Name), strings.ToLower(filter.Keyword)) {
			var imageUrl string
			if v.Attributes.IconUrl != "" {
				imageUrl = v.Attributes.IconUrl
			} else if v.Attributes.BannerUrl != "" {
				imageUrl = v.Attributes.BannerUrl
			}

			release, err := intToTime(v.Attributes.ReleaseDate)
			if err != nil {
				web.sugarLogger.Error(fmt.Errorf("parsing time failed: %w", err))
			}

			items = append(items, TitleItem {
				ImageUrl:    imageUrl,
				Id:          strings.ToUpper(v.Attributes.Id),
				Name:        v.Attributes.Name,
				Region:      v.Attributes.Region,
				ReleaseDate: release,
			})
		}
	}

	p := pagination.Calculate(filter.Page, filter.PerPage, len(items))

	if err := sortItems(filter, items); err != nil {
		web.sugarLogger.Error(err)
	}

	return items[p.Start:p.End], p
}
