package web

import (
	"fmt"
	"github.com/dtrunk90/switch-library-manager-web/pagination"
	"github.com/dtrunk90/switch-library-manager-web/process"
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"strings"
)

func (web *Web) HandleUpdates() {
	fsPatterns := []string {
		"resources/layout.html",
		"resources/partials/card.html",
		"resources/partials/filter.html",
		"resources/partials/pagination.html",
		"resources/pages/updates.html",
	}

	web.HandleFiltered("/updates.html", func(filter *TitleItemFilter) any {
		globalPageData.IsKeysFileAvailable = settings.IsKeysFileAvailable()
		globalPageData.Page = "updates"
		items, p := web.getMissingUpdates(filter)
		return TitleItemsPageData {
			GlobalPageData: globalPageData,
			TitleItems: items,
			Filter: filter,
			Pagination: p,
		}
	}, web.embedFS, fsPatterns...)
}

func (web *Web) getMissingUpdates(filter *TitleItemFilter) ([]TitleItem, pagination.Pagination) {
	items := []TitleItem{}

	if web.state.localDB == nil {
		return items, pagination.Calculate(filter.Page, filter.PerPage, 0)
	}

	missingUpdates := process.ScanForMissingUpdates(web.state.localDB.TitlesMap, web.state.switchDB.TitlesMap)

	for _, v := range missingUpdates {
		if filter.Keyword == "" || strings.Contains(strings.ToLower(v.Attributes.Id), strings.ToLower(filter.Keyword)) || strings.Contains(strings.ToLower(v.Attributes.Name), strings.ToLower(filter.Keyword)) {
			var imageUrl string
			if v.Attributes.IconUrl != "" {
				imageUrl = v.Attributes.IconUrl
			} else if v.Attributes.BannerUrl != "" {
				imageUrl = v.Attributes.BannerUrl
			}

			latest, err := strToTime("2006-01-02", v.LatestUpdateDate)
			if err != nil {
				web.sugarLogger.Error(fmt.Errorf("parsing time failed: %w", err))
			}

			release, err := intToTime(v.Attributes.ReleaseDate)
			if err != nil {
				web.sugarLogger.Error(fmt.Errorf("parsing time failed: %w", err))
			}

			items = append(items, TitleItem {
				ImageUrl:         imageUrl,
				Id:               strings.ToUpper(v.Attributes.Id),
				LatestUpdate:     v.LatestUpdate,
				LatestUpdateDate: latest,
				LocalUpdate:      v.LocalUpdate,
				Name:             v.Attributes.Name,
				Region:           v.Attributes.Region,
				ReleaseDate:      release,
				Type:             strings.ToUpper(v.Meta.Type),
			})
		}
	}

	p := pagination.Calculate(filter.Page, filter.PerPage, len(items))

	if err := sortItems(filter, items); err != nil {
		web.sugarLogger.Error(err)
	}

	return items[p.Start:p.End], p
}
