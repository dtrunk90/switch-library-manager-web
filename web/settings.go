package web

import (
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"os"
	"regexp"
	"strings"
)

type SettingsPageData struct {
	GlobalPageData
	Settings *settings.AppSettings
}

type SettingsForm struct {
	Prodkeys          string `in:"form=prod_keys"`
	ScanFolders       string `in:"form=scan_folders"`
	IgnoreDLCTitleIds string `in:"form=ignore_dlc_title_ids"`
}

func SplitAndTrimSpaceArray(s string, sep string) []string {
	arr := []string{}

	for _, v := range strings.Split(s, sep) {
		if value := strings.TrimSpace(v); value != "" {
			arr = append(arr, value)
		}
	}

	return arr
}

func (web *Web) HandleSettings() {
	fsPatterns := []string {
		"resources/layout.html",
		"resources/pages/settings.html",
	}

	web.HandleValidated("/settings.html", SettingsForm{}, func() any {
		globalPageData.IsKeysFileAvailable = settings.IsKeysFileAvailable()
		globalPageData.Page = "settings"
		return SettingsPageData {
			GlobalPageData: globalPageData,
			Settings: web.appSettings,
		}
	}, func(value any) ErrorResponse {
		settingsForm := value.(*SettingsForm)
		errorResponse := ErrorResponse{
			FieldErrors: []FieldError{},
		}

		if strings.TrimSpace(settingsForm.Prodkeys) != "" {
			keys, err := settings.GetSwitchKeys(settingsForm.Prodkeys)
			if err != nil {
				errorResponse.FieldErrors = append(errorResponse.FieldErrors, FieldError {
					Field: "prod_keys",
					Message: "Error trying to read Product Keys (" + err.Error() + ")",
				})
			} else if keys["header_key"] == "" {
				errorResponse.FieldErrors = append(errorResponse.FieldErrors, FieldError {
					Field: "prod_keys",
					Message: "Please provide a valid Product Keys Path",
				})
			}
		}

		scanFolders := SplitAndTrimSpaceArray(settingsForm.ScanFolders, "\n")

		if len(scanFolders) == 0 {
			errorResponse.FieldErrors = append(errorResponse.FieldErrors, FieldError {
				Field: "scan_folders",
				Message: "Please provide at least one Folder to scan",
			})
		} else {
			for _, value := range scanFolders {
				if _, err := os.Stat(value); os.IsNotExist(err) || os.IsPermission(err) {
					errorResponse.FieldErrors = append(errorResponse.FieldErrors, FieldError {
						Field: "scan_folders",
						Message: "Folder not found: " + value,
					})

					break
				}
			}
		}

		r, _ := regexp.Compile("^[0-9A-Fa-f]+$")

		for _, value := range SplitAndTrimSpaceArray(settingsForm.IgnoreDLCTitleIds, "\n") {
			if !r.MatchString(value) {
				errorResponse.FieldErrors = append(errorResponse.FieldErrors, FieldError {
					Field: "ignore_dlc_title_ids",
					Message: "Invalid Title ID: " + value,
				})

				break
			}
		}

		return errorResponse
	}, func(value any) SuccessResponse {
		settingsForm := value.(*SettingsForm)
		scanFolders := SplitAndTrimSpaceArray(settingsForm.ScanFolders, "\n")

		appSettings := settings.ReadSettings(web.dataFolder)
		appSettings.Prodkeys = settingsForm.Prodkeys
		appSettings.IgnoreDLCTitleIds = SplitAndTrimSpaceArray(settingsForm.IgnoreDLCTitleIds, "\n")
		appSettings.Folder = scanFolders[0]
		if len(scanFolders) > 1 {
			appSettings.ScanFolders = scanFolders[1:]
		} else {
			appSettings.ScanFolders = []string{}
		}

		settings.SaveSettings(appSettings, web.dataFolder)
		web.appSettings = appSettings

		settings.InitSwitchKeys(web.dataFolder)
		web.buildLocalDB(web.localDbManager, true)

		return SuccessResponse {
			StrongMessage: "Success!",
			Message: "Settings changed successfully.",
		}
	}, web.embedFS, fsPatterns...)
}
