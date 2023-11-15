package web

import (
	"encoding/json"
	"github.com/dtrunk90/switch-library-manager-web/db"
	"github.com/gorilla/mux"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type ApiFileInfo struct {
	DownloadUrl string `json:"downloadUrl,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Type        string `json:"type,omitempty"`
}

type ApiSystemInfo struct {
	RequiredSystemVersion int `json:"requiredSystemVersion,omitempty"`
}

type ApiExtendedFileInfo struct {
	ApiFileInfo
	DisplayVersion string `json:"displayVersion,omitempty"`
	Version        int    `json:"version"`
}

type ApiUpdateItem struct {
	ApiExtendedFileInfo
	ApiSystemInfo
}

type ApiDlcItem struct {
	ApiExtendedFileInfo
	Name                       string `json:"name"`
	RequiredApplicationVersion int    `json:"requiredApplicationVersion"`
}

type ApiTitleItem struct {
	ApiFileInfo
	ApiSystemInfo
	BannerUrl     string                `json:"bannerUrl,omitempty"`
	IconUrl       string                `json:"iconUrl,omitempty"`
	ThumbnailUrl  string                `json:"thumbnailUrl,omitempty"`
	LatestUpdate  ApiUpdateItem         `json:"latestUpdate"`
	Name          map[string]string     `json:"name"`
	Region        string                `json:"region,omitempty"`
	Dlc           map[string]ApiDlcItem `json:"dlc,omitempty"`
}

func (web *Web) HandleApi() {
	web.router.HandleFunc("/api/titles", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		items := map[string]ApiTitleItem{}

		if web.state.localDB != nil {
			for k, v := range web.state.localDB.TitlesMap {
				if v.BaseExist {
					titleId := strings.ToUpper(v.File.Metadata.TitleId)
					latestUpdate := ApiUpdateItem{ ApiExtendedFileInfo: ApiExtendedFileInfo { Version: v.LatestUpdate } }
					name := map[string]string{}

					if v.File.Metadata.Ncap != nil {
						latestUpdate.DisplayVersion = v.File.Metadata.Ncap.DisplayVersion
					}

					if v.Updates != nil && len(v.Updates) != 0 {
						latestUpdate.DownloadUrl = "/api/titles/" + titleId + "/updates/" + strconv.Itoa(v.LatestUpdate)
						latestUpdate.Size = v.Updates[v.LatestUpdate].ExtendedInfo.Size
						latestUpdate.Type = strings.ToUpper(filepath.Ext(v.Updates[v.LatestUpdate].ExtendedInfo.FileName)[1:])

						if v.Updates[v.LatestUpdate].Metadata.Ncap != nil {
							latestUpdate.DisplayVersion = v.Updates[v.LatestUpdate].Metadata.Ncap.DisplayVersion
						}
					}

					if v.File.Metadata.Ncap != nil {
						for _, langV := range v.File.Metadata.Ncap.TitleName {
							if langV.Title != "" {
								name[langV.Language.ToLanguageTag()] = langV.Title
							}
						}
					}

					if len(name) == 0 {
						name["unknown"] = db.ParseTitleNameFromFileName(v.File.ExtendedInfo.FileName)
					}

					items[titleId] = ApiTitleItem {
						ApiFileInfo:   ApiFileInfo {
							DownloadUrl:  "/api/titles/" + titleId,
							Size:         v.File.ExtendedInfo.Size,
							Type:         strings.ToUpper(filepath.Ext(v.File.ExtendedInfo.FileName)[1:]),
						},
						ApiSystemInfo: ApiSystemInfo {
							RequiredSystemVersion:        v.File.Metadata.RequiredTitleVersion,
						},
						LatestUpdate:  latestUpdate,
						Name:          name,
					}

					if title, ok1 := web.state.switchDB.TitlesMap[k]; ok1 {
						if item, ok2 := items[titleId]; ok2 {
							item.Region = title.Attributes.Region
							items[titleId] = item
						}
					}

					if item, ok1 := items[titleId]; ok1 {
						if v.Banner != "" {
							item.BannerUrl = "/i/" + v.Banner
						}

						if v.Icon != "" {
							item.IconUrl = "/i/" + v.Icon
						}

						if item.IconUrl != "" {
							item.ThumbnailUrl = item.IconUrl + "?width=90"
						} else if item.BannerUrl != "" {
							item.ThumbnailUrl = item.BannerUrl + "?width=90"
						}

						item.Dlc = map[string]ApiDlcItem{}

						for id, dlc := range v.Dlc {
							dlcTitleId := strings.ToUpper(id)

							item.Dlc[dlcTitleId] = ApiDlcItem {
								ApiExtendedFileInfo:        ApiExtendedFileInfo {
									ApiFileInfo:    ApiFileInfo {
										DownloadUrl: "/api/titles/" + titleId + "/dlc/" + dlcTitleId,
										Size:        dlc.ExtendedInfo.Size,
										Type:        strings.ToUpper(filepath.Ext(dlc.ExtendedInfo.FileName)[1:]),
									},
									Version:        dlc.Metadata.Version,
								},
								RequiredApplicationVersion: dlc.Metadata.RequiredTitleVersion,
							}

							if entry, ok2 := web.state.switchDB.TitlesMap[k].Dlc[id]; ok2 {
								if dlcItem, ok3 := item.Dlc[dlcTitleId]; ok3 {
									dlcItem.Name = entry.Name
									item.Dlc[dlcTitleId] = dlcItem
								}
							}

							if entry, ok2 := item.Dlc[dlcTitleId]; ok2 {
								if dlc.Metadata.Ncap != nil {
									entry.DisplayVersion = dlc.Metadata.Ncap.DisplayVersion
								}

								item.Dlc[dlcTitleId] = entry
							}
						}

						items[titleId] = item
					}
				}
			}
		}

		json.NewEncoder(w).Encode(items)
	})

	web.router.HandleFunc("/api/titles/{titleId}", func(w http.ResponseWriter, r *http.Request) {
		if web.state.localDB != nil {
			vars := mux.Vars(r)

			for _, v := range web.state.localDB.TitlesMap {
				if v.BaseExist && strings.ToUpper(v.File.Metadata.TitleId) == strings.ToUpper(vars["titleId"]) {
					w.Header().Set("Content-Type", "application/octet-stream")
					w.Header().Set("Content-Disposition", "attachment; filename=" + strconv.Quote(v.File.ExtendedInfo.FileName))
					http.ServeFile(w, r, filepath.Join(v.File.ExtendedInfo.BaseFolder, v.File.ExtendedInfo.FileName))
					return
				}
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	web.router.HandleFunc("/api/titles/{titleId}/updates/{version}", func(w http.ResponseWriter, r *http.Request) {
		if web.state.localDB != nil {
			vars := mux.Vars(r)

			if version, err := strconv.Atoi(vars["version"]); err == nil {
				for _, v := range web.state.localDB.TitlesMap {
					if v.BaseExist && strings.ToUpper(v.File.Metadata.TitleId) == strings.ToUpper(vars["titleId"]) {
						if v.Updates != nil && len(v.Updates) != 0 {
							if update, ok := v.Updates[version]; ok {
								w.Header().Set("Content-Type", "application/octet-stream")
								w.Header().Set("Content-Disposition", "attachment; filename=" + strconv.Quote(update.ExtendedInfo.FileName))
								http.ServeFile(w, r, filepath.Join(update.ExtendedInfo.BaseFolder, update.ExtendedInfo.FileName))
								return
							}
						}

						break
					}
				}
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})

	web.router.HandleFunc("/api/titles/{titleId}/dlc/{dlcTitleId}", func(w http.ResponseWriter, r *http.Request) {
		if web.state.localDB != nil {
			vars := mux.Vars(r)

			for _, v := range web.state.localDB.TitlesMap {
				if v.BaseExist && strings.ToUpper(v.File.Metadata.TitleId) == strings.ToUpper(vars["titleId"]) {
					if v.Dlc != nil && len(v.Dlc) != 0 {
						if dlc, ok := v.Dlc[strings.ToUpper(vars["dlcTitleId"])]; ok {
							w.Header().Set("Content-Type", "application/octet-stream")
							w.Header().Set("Content-Disposition", "attachment; filename=" + strconv.Quote(dlc.ExtendedInfo.FileName))
							http.ServeFile(w, r, filepath.Join(dlc.ExtendedInfo.BaseFolder, dlc.ExtendedInfo.FileName))
							return
						}
					}

					break
				}
			}
		}

		w.WriteHeader(http.StatusNotFound)
	})
}
