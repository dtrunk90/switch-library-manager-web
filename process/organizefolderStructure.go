package process

import (
	"github.com/giwty/switch-library-manager/db"
	"github.com/giwty/switch-library-manager/settings"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"robpike.io/nihongo"
	"strconv"
	"strings"
)

var (
	folderIllegalCharsRegex = regexp.MustCompile(`[/\\?%*:;=|"<>]`)
	nonAscii                = regexp.MustCompile("[a-zA-Z0-9áéíóú@#%&',.\\s-\\[\\]\\(\\)\\+]")
)

func DeleteOldUpdates(localDB *db.LocalSwitchFilesDB, updateProgress db.ProgressUpdater) {
	i := 0
	for k, v := range localDB.TitlesMap {
		if !v.BaseExist {
			continue
		}
		i++
		if updateProgress != nil {
			updateProgress.UpdateProgress(i, len(localDB.TitlesMap), v.File.ExtendedInfo.Info.Name()+k)
		}
		if len(v.Updates) > 1 {

			for version, update := range v.Updates {
				if version < v.LatestUpdate && version != 0 {
					fileToRemove := filepath.Join(update.ExtendedInfo.BaseFolder, update.ExtendedInfo.Info.Name())
					zap.S().Infof("--> [Delete] Old update file: %v [latest update:%v]\n", fileToRemove, v.LatestUpdate)
					err := os.Remove(fileToRemove)
					if err != nil {
						zap.S().Errorf("Failed to delete file  %v  [%v]\n", fileToRemove, err)
					}
				}
			}
		}

	}
}

func OrganizeByFolders(baseFolder string,
	localDB *db.LocalSwitchFilesDB,
	titlesDB *db.SwitchTitlesDB,
	updateProgress db.ProgressUpdater) {

	//validate template rules

	options := settings.ReadSettings(baseFolder).OrganizeOptions
	if !isOptionsValid(options) {
		zap.S().Error("the organize options in settings.json are not valid, please check that the template contains file/folder name")
		return
	}
	i := 0
	tasksSize := len(localDB.TitlesMap) + 2
	for k, v := range localDB.TitlesMap {
		i++
		if v.BaseExist == false {
			continue
		}
		if updateProgress != nil {
			updateProgress.UpdateProgress(i, tasksSize, v.File.ExtendedInfo.Info.Name())
		}

		titleName := getTitleName(titlesDB.TitlesMap[k], v)

		templateData := map[string]string{}

		templateData[settings.TEMPLATE_TITLE_ID] = v.File.Metadata.TitleId
		//templateData[settings.TEMPLATE_TYPE] = "BASE"
		templateData[settings.TEMPLATE_TITLE_NAME] = titleName
		templateData[settings.TEMPLATE_VERSION_TXT] = ""
		templateData[settings.TEMPLATE_VERSION] = "0"

		if v.File.Metadata.Ncap != nil {
			templateData[settings.TEMPLATE_VERSION_TXT] = v.File.Metadata.Ncap.DisplayVersion
		}

		var destinationPath = v.File.ExtendedInfo.BaseFolder

		//create folder if needed
		if options.CreateFolderPerGame {
			folderToCreate := getFolderName(options, templateData)
			destinationPath = filepath.Join(baseFolder, folderToCreate)
			if _, err := os.Stat(destinationPath); os.IsNotExist(err) {
				err = os.Mkdir(destinationPath, os.ModePerm)
				if err != nil {
					zap.S().Errorf("Failed to create folder %v - %v\n", folderToCreate, err)
					continue
				}
			}
		}

		//process base title
		from := filepath.Join(v.File.ExtendedInfo.BaseFolder, v.File.ExtendedInfo.Info.Name())
		to := filepath.Join(destinationPath, getFileName(options, v.File.ExtendedInfo.Info.Name(), templateData))
		err := moveFile(from, to)
		if err != nil {
			zap.S().Errorf("Failed to move file [%v]\n", err)
			continue
		}

		//process updates
		for update, updateInfo := range v.Updates {
			if updateInfo.Metadata != nil {
				templateData[settings.TEMPLATE_TITLE_ID] = updateInfo.Metadata.TitleId
			}
			templateData[settings.TEMPLATE_VERSION] = strconv.Itoa(update)
			templateData[settings.TEMPLATE_TYPE] = "UPD"
			if updateInfo.Metadata.Ncap != nil {
				templateData[settings.TEMPLATE_VERSION_TXT] = updateInfo.Metadata.Ncap.DisplayVersion
			} else {
				templateData[settings.TEMPLATE_VERSION_TXT] = ""
			}

			from = filepath.Join(updateInfo.ExtendedInfo.BaseFolder, updateInfo.ExtendedInfo.Info.Name())
			if options.CreateFolderPerGame {
				to = filepath.Join(destinationPath, getFileName(options, updateInfo.ExtendedInfo.Info.Name(), templateData))
			} else {
				to = filepath.Join(updateInfo.ExtendedInfo.BaseFolder, getFileName(options, updateInfo.ExtendedInfo.Info.Name(), templateData))
			}
			err := moveFile(from, to)
			if err != nil {
				zap.S().Errorf("Failed to move file [%v]\n", err)
				continue
			}
		}

		//process DLC
		for id, dlc := range v.Dlc {
			if dlc.Metadata != nil {
				templateData[settings.TEMPLATE_VERSION] = strconv.Itoa(dlc.Metadata.Version)
			}
			templateData[settings.TEMPLATE_TYPE] = "DLC"
			templateData[settings.TEMPLATE_TITLE_ID] = id
			templateData[settings.TEMPLATE_DLC_NAME] = getDlcName(titlesDB.TitlesMap[k], dlc)
			from = filepath.Join(dlc.ExtendedInfo.BaseFolder, dlc.ExtendedInfo.Info.Name())
			if options.CreateFolderPerGame {
				to = filepath.Join(destinationPath, getFileName(options, dlc.ExtendedInfo.Info.Name(), templateData))
			} else {
				to = filepath.Join(dlc.ExtendedInfo.BaseFolder, getFileName(options, dlc.ExtendedInfo.Info.Name(), templateData))
			}
			err = moveFile(from, to)
			if err != nil {
				zap.S().Errorf("Failed to move file [%v]\n", err)
				continue
			}
		}
	}

	if options.DeleteEmptyFolders {
		if updateProgress != nil {
			i += 1
			updateProgress.UpdateProgress(i, tasksSize, "deleting empty folders... (can take 1-2min)")
		}
		err := deleteEmptyFolders(baseFolder)
		if err != nil {
			zap.S().Errorf("Failed to delete empty folders [%v]\n", err)
		}
		if updateProgress != nil {
			i += 1
			updateProgress.UpdateProgress(i, tasksSize, "done")
		}
	} else {
		if updateProgress != nil {
			i += 2
			updateProgress.UpdateProgress(i, tasksSize, "done")
		}
	}
}

func isOptionsValid(options settings.OrganizeOptions) bool {
	if options.RenameFiles {
		if options.FileNameTemplate == "" {
			zap.S().Error("file name template cannot be empty")
			return false
		}
		if !strings.Contains(options.FileNameTemplate, settings.TEMPLATE_TITLE_NAME) &&
			!strings.Contains(options.FileNameTemplate, settings.TEMPLATE_TITLE_ID) {
			zap.S().Error("file name template needs to contain one of the following - titleId or title name")
			return false
		}

	}

	if options.CreateFolderPerGame {
		if options.FolderNameTemplate == "" {
			zap.S().Error("folder name template cannot be empty")
			return false
		}
		if !strings.Contains(options.FolderNameTemplate, settings.TEMPLATE_TITLE_NAME) &&
			!strings.Contains(options.FolderNameTemplate, settings.TEMPLATE_TITLE_ID) {
			zap.S().Error("folder name template needs to contain one of the following - titleId or title name")
			return false
		}
	}
	return true
}

func getDlcName(switchTitle *db.SwitchTitle, file db.SwitchFileInfo) string {
	if switchTitle == nil {
		return ""
	}
	if dlcAttributes, ok := switchTitle.Dlc[file.Metadata.TitleId]; ok {
		name := dlcAttributes.Name
		name = strings.ReplaceAll(name, "\n", "")
		return name
	}
	return ""
}

func getTitleName(switchTitle *db.SwitchTitle, v *db.SwitchGameFiles) string {
	if switchTitle != nil && switchTitle.Attributes.Name != "" {
		return switchTitle.Attributes.Name
	} else {

		if v.File.Metadata.Ncap != nil {
			name := v.File.Metadata.Ncap.TitleName["AmericanEnglish"].Title
			if name != "" {
				return name
			}
		}
		//for non eshop games (cartridge only), grab the name from the file
		return db.ParseTitleNameFromFileName(v.File.ExtendedInfo.Info.Name())
	}
}

func getFolderName(options settings.OrganizeOptions, templateData map[string]string) string {

	return applyTemplate(templateData, options.SwitchSafeFileNames, options.FolderNameTemplate)
}

func getFileName(options settings.OrganizeOptions, originalName string, templateData map[string]string) string {
	if !options.RenameFiles {
		return originalName
	}
	ext := path.Ext(originalName)
	result := applyTemplate(templateData, options.SwitchSafeFileNames, options.FileNameTemplate)
	return result + ext
}

func moveFile(from string, to string) error {
	if from == to {
		return nil
	}
	err := os.Rename(from, to)
	return err
}

func applyTemplate(templateData map[string]string, useSafeNames bool, template string) string {
	result := strings.Replace(template, "{"+settings.TEMPLATE_TITLE_NAME+"}", templateData[settings.TEMPLATE_TITLE_NAME], 1)
	result = strings.Replace(result, "{"+settings.TEMPLATE_TITLE_ID+"}", strings.ToUpper(templateData[settings.TEMPLATE_TITLE_ID]), 1)
	result = strings.Replace(result, "{"+settings.TEMPLATE_VERSION+"}", templateData[settings.TEMPLATE_VERSION], 1)
	result = strings.Replace(result, "{"+settings.TEMPLATE_TYPE+"}", templateData[settings.TEMPLATE_TYPE], 1)
	result = strings.Replace(result, "{"+settings.TEMPLATE_VERSION_TXT+"}", templateData[settings.TEMPLATE_VERSION_TXT], 1)
	//remove title name from dlc name
	dlcName := strings.Replace(templateData[settings.TEMPLATE_DLC_NAME], templateData[settings.TEMPLATE_TITLE_NAME], "", 1)
	dlcName = strings.TrimSpace(dlcName)
	dlcName = strings.TrimPrefix(dlcName, "-")
	dlcName = strings.TrimSpace(dlcName)
	result = strings.Replace(result, "{"+settings.TEMPLATE_DLC_NAME+"}", dlcName, 1)
	result = strings.ReplaceAll(result, "[]", "")
	result = strings.ReplaceAll(result, "()", "")
	result = strings.ReplaceAll(result, "<>", "")
	if strings.HasSuffix(result, ".") {
		result = result[:len(result)-1]
	}

	if useSafeNames {
		result = nihongo.RomajiString(result)
		safe := nonAscii.FindAllString(result, -1)
		result = strings.Join(safe, "")
	}
	result = strings.ReplaceAll(result, "  ", " ")
	result = strings.TrimSpace(result)
	return folderIllegalCharsRegex.ReplaceAllString(result, "")
}

func deleteEmptyFolders(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			zap.S().Error("Error while deleting empty folders", err)
		}
		if info != nil && info.IsDir() {
			err = deleteEmptyFolder(path)
			if err != nil {
				zap.S().Error("Error while deleting empty folders", err)
			}
		}

		return nil
	})
	return err
}

func deleteEmptyFolder(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return err
	}

	if len(files) != 0 {
		return nil
	}

	zap.S().Infof("\nDeleting empty folder [%v]", path)
	_ = os.Remove(path)

	return nil
}
