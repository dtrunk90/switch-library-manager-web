package web

import (
	"embed"
	"errors"
	"fmt"
	"github.com/dtrunk90/switch-library-manager-web/db"
	"github.com/dtrunk90/switch-library-manager-web/pagination"
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type WebState struct {
	sync.Mutex
	switchDB *db.SwitchTitlesDB
	localDB  *db.LocalSwitchFilesDB
}

type Web struct {
	state          WebState
	router         *mux.Router
	embedFS        embed.FS
	appSettings    *settings.AppSettings
	dataFolder     string
	localDbManager *db.LocalSwitchDBManager
	sugarLogger    *zap.SugaredLogger
}

type TitleItem struct {
	ImageUrl         string
	Id               string
	LatestUpdate     int
	LatestUpdateDate time.Time
	LocalUpdate      int
	MissingDLC       []string
	Name             string
	Region           string
	ReleaseDate      time.Time
	Type             string
	Version          string
}

type GlobalPageData struct {
	IsKeysFileAvailable bool
	Page                string
	SlmVersion          string	
	Version             string
}

type TitleItemsPageData struct {
	GlobalPageData
	TitleItems []TitleItem
	Filter     *TitleItemFilter
	Pagination pagination.Pagination
}

var funcMap = template.FuncMap {
	"add": func(a, b int) int {
		return a + b
	},
	"eq": func(a, b interface{}) bool {
		return a == b
	},
	"gt": func(a, b int) bool {
		return a > b
	},
	"lt": func(a, b int) bool {
		return a < b
	},
	"mkRange": func(min, max int) []int {
		a := make([]int, max - min + 1)
		for i := range a {
			a[i] = min + i
		}
		return a
	},
	"mkSlice": func(args ...interface{}) []interface{} {
		return args
	},
	"neq": func(a, b interface{}) bool {
		return a != b
	},
	"formatTime": func(value time.Time) string {
		if value.IsZero() {
			return ""
		}

		return value.Format("2006-01-02")
	},
	"subtract": func(a, b int) int {
		return a - b
	},
	"toLower": strings.ToLower,
}

var globalPageData = GlobalPageData {
	SlmVersion: settings.SLM_VERSION,
	Version: settings.SLM_WEB_VERSION,
}

func intToTime(value int) (time.Time, error) {
	if value <= 0 {
		return time.Time{}, nil
	}

	return strToTime("20060102", strconv.Itoa(value))
}

func strToTime(layout, value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}

	return time.Parse(layout, value)
}

func CreateWeb(router *mux.Router, embedFS embed.FS, appSettings *settings.AppSettings, dataFolder string, sugarLogger *zap.SugaredLogger) *Web {
	return &Web{state: WebState{}, router: router, embedFS: embedFS, appSettings: appSettings, dataFolder: dataFolder, sugarLogger: sugarLogger}
}

func (web *Web) Start() {
	web.updateDB()

	localDbManager, err := db.NewLocalSwitchDBManager(web.dataFolder)
	if err != nil {
		web.sugarLogger.Error("Failed to create local files db\n", err)
		return
	}

	_, err = settings.InitSwitchKeys(web.dataFolder)
	if err != nil {
		web.sugarLogger.Errorf("Failed to initialize switch keys: %s", err)
	}

	web.localDbManager = localDbManager
	defer localDbManager.Close()

	if _, err := web.buildLocalDB(web.localDbManager, true); err != nil {
		web.sugarLogger.Error(err)
	}

	// Run http server
	web.HandleResources()
	web.HandleImages()
	web.HandleIndex()
	web.HandleMissing()
	web.HandleUpdates()
	web.HandleDLC()
	web.HandleIssues()
	web.HandleSettings()
	web.HandleOrganize()
	web.HandleApi()

	web.router.Handle("/", http.RedirectHandler("/index.html", http.StatusMovedPermanently))

	http.Handle("/", web.router)

	web.sugarLogger.Info("[SLM started]")

	if err := http.ListenAndServe(fmt.Sprint(":", web.appSettings.Port), nil); err != nil {
		web.sugarLogger.Error(fmt.Errorf("running http server failed: %w", err))
		log.Fatal(err)
	}
}

func (web *Web) UpdateProgress(curr int, total int, message string) {
	web.sugarLogger.Debugf("%v (%v/%v)", message, curr, total)
}

func (web *Web) updateDB() {
	if web.state.switchDB == nil {
		switchDb, err := web.buildSwitchDb()
		if err != nil {
			web.sugarLogger.Error(err)
			return
		}
		web.state.switchDB = switchDb
	}
}

func (web *Web) buildSwitchDb() (*db.SwitchTitlesDB, error) {
	settingsObj := settings.ReadSettings(web.dataFolder)

	web.UpdateProgress(1, 4, "Downloading titles.json")
	filename := filepath.Join(web.dataFolder, settings.TITLE_JSON_FILENAME)
	titleFile, titlesEtag, err := db.LoadAndUpdateFile(settings.TITLES_JSON_URL, filename, settingsObj.TitlesEtag)

	if err != nil {
		return nil, errors.New("failed to download switch titles [reason:" + err.Error() + "]")
	}

	settingsObj.TitlesEtag = titlesEtag

	web.UpdateProgress(2, 4, "Downloading versions.json")
	filename = filepath.Join(web.dataFolder, settings.VERSIONS_JSON_FILENAME)
	versionsFile, versionsEtag, err := db.LoadAndUpdateFile(settings.VERSIONS_JSON_URL, filename, settingsObj.VersionsEtag)

	if err != nil {
		return nil, errors.New("failed to download switch updates [reason:" + err.Error() + "]")
	}

	settingsObj.VersionsEtag = versionsEtag

	settings.SaveSettings(settingsObj, web.dataFolder)

	web.UpdateProgress(3, 4, "Processing switch titles and updates ...")
	switchTitleDB, err := db.CreateSwitchTitleDB(titleFile, versionsFile)

	web.UpdateProgress(4, 4, "Finishing up...")

	return switchTitleDB, err
}

func (web *Web) buildLocalDB(localDbManager *db.LocalSwitchDBManager, ignoreCache bool) (*db.LocalSwitchFilesDB, error) {
	settingsObj := settings.ReadSettings(web.dataFolder)

	folderToScan := settingsObj.Folder

	scanFolders := settingsObj.ScanFolders
	scanFolders = append(scanFolders, folderToScan)

	localDB, err := localDbManager.CreateLocalSwitchFilesDB(web.state.switchDB, web.dataFolder, scanFolders, web, true, ignoreCache)
	web.state.localDB = localDB

	return localDB, err
}
