package main

import (
	"embed"
	"fmt"
	"github.com/dtrunk90/switch-library-manager-web/settings"
	"github.com/dtrunk90/switch-library-manager-web/web"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

//go:embed resources/static
//go:embed node_modules/bootstrap-icons/font/fonts
//go:embed node_modules/flag-icons/flags
//go:embed resources/layout.html
//go:embed resources/partials/*.html
//go:embed resources/pages/*.html
var embedFS embed.FS

func main() {

	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("failed to get executable directory, please ensure app has sufficient permissions. aborting")
		return
	}

	router := mux.NewRouter()

	dataFolder, ok := os.LookupEnv("SLM_DATA_DIR")
	if !ok {
		dataFolder = filepath.Dir(exePath)
	}

	appSettings := settings.ReadSettings(dataFolder)

	logger := createLogger(appSettings.Debug)

	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()

	destinationPath := filepath.Join(dataFolder, "img")
	if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
		if err := os.Mkdir(destinationPath, os.ModePerm); err != nil {
			sugar.Errorf("Failed to create folder img - %v\n", err)
		}
	}

	sugar.Info("[SLM starts]")
	sugar.Infof("[Executable: %v]", exePath)
	sugar.Infof("[Data folder: %v]", dataFolder)

	web.CreateWeb(router, embedFS, appSettings, dataFolder, sugar).Start()

}

func createLogger(debug bool) *zap.Logger {
	var config zap.Config
	if debug {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	logger, err := config.Build()
	if err != nil {
		fmt.Printf("failed to create logger - %v", err)
		panic(1)
	}
	zap.ReplaceGlobals(logger)
	return logger
}
