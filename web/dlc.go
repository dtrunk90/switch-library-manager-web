package web

import (
	"github.com/dtrunk90/switch-library-manager-web/pagination"
	"github.com/dtrunk90/switch-library-manager-web/process"
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"strings"
)

func (web *Web) HandleDLC() {
	fsPatterns := []string {
		"resources/layout.html",
		"resources/partials/filter.html",
		"resources/partials/pagination.html",
		"resources/pages/dlc.html",
	}

	web.HandleFiltered("/dlc.html", func(filter *TitleItemFilter) any {
		globalPageData.IsKeysFileAvailable = settings.IsKeysFileAvailable()
		globalPageData.Page = "dlc"
		items, p := web.getMissingDLC(filter)
		return TitleItemsPageData {
			GlobalPageData: globalPageData,
			TitleItems: items,
			Filter: filter,
			Pagination: p,
		}
	}, web.embedFS, fsPatterns...)
}

func (web *Web) getMissingDLC(filter *TitleItemFilter) ([]TitleItem, pagination.Pagination) {
	items := []TitleItem{}

	if web.state.localDB == nil {
		return items, pagination.Calculate(filter.Page, filter.PerPage, 0)
	}

	settingsObj := settings.ReadSettings(web.dataFolder)
	ignoreIds := map[string]struct{}{}

	for _, id := range settingsObj.IgnoreDLCTitleIds {
		ignoreIds[strings.ToLower(id)] = struct{}{}
	}

	missingDLC := process.ScanForMissingDLC(web.state.localDB.TitlesMap, web.state.switchDB.TitlesMap, ignoreIds)

	for _, v := range missingDLC {
		if filter.Keyword == "" || strings.Contains(strings.ToLower(v.Attributes.Id), strings.ToLower(filter.Keyword)) || strings.Contains(strings.ToLower(v.Attributes.Name), strings.ToLower(filter.Keyword)) {
			var imageUrl string
			if v.Attributes.IconUrl != "" {
				imageUrl = v.Attributes.IconUrl
			} else if v.Attributes.BannerUrl != "" {
				imageUrl = v.Attributes.BannerUrl
			}

			items = append(items, TitleItem {
				ImageUrl:         imageUrl,
				Id:               strings.ToUpper(v.Attributes.Id),
				MissingDLC:       v.MissingDLC,
				Name:             v.Attributes.Name,
				Region:           v.Attributes.Region,
			})
		}
	}

	p := pagination.Calculate(filter.Page, filter.PerPage, len(items))

	if err := sortItems(filter, items); err != nil {
		web.sugarLogger.Error(err)
	}

	return items[p.Start:p.End], p
}
