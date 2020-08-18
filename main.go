package main

import (
	"encoding/json"
	"fmt"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/bitrise-tools/go-steputils/tools"
	"io/ioutil"
	"os"
	"strings"
)

type Responsible struct {
	Key              string   `json:"key"`
	Modules          []string `json:"modules"`
	SlackResponsible []string `json:"slack_responsible"`
}

type Config struct {
	AllowedKeys       string `env:"jira_keys"`
	PR                string `env:"pr"`
	Branch            string `env:"branch"`
	Folders           string `env:"folders"`
	PathConfiguration string `env:"path_configuration"`
}

func main() {
	/*
		var cfg Config = Config{
			Folders:           "tools\nbasket\nbase\n",
			PathConfiguration: "tools/responsible.json",
			Branch:            "fix/SHP-22/huascar",
			AllowedKeys:       "BAS|SHP|OT",
		}
	*/
	var cfg Config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Issue with input: %s", err)
	}

	file, _ := ioutil.ReadFile(cfg.PathConfiguration)

	var jsonDataArray []Responsible

	_ = json.Unmarshal(file, &jsonDataArray)

	branchKey := extraBranchKey(cfg.Branch, cfg.AllowedKeys)

	if branchKey == "" {
		failf("Key don't allowed")
	}

	var arrayOfFolders = removeDuplicateValues(strings.Split(cfg.Folders, "|"))

	var indexOfKey = getIndexOfKeyProject(jsonDataArray, branchKey)

	foldersTouchedByProject := map[string][]string{}

	fillFoldersTouchedByProject(arrayOfFolders, jsonDataArray, indexOfKey, foldersTouchedByProject)
	if len(foldersTouchedByProject) > 0 {
		var message = ""
		if indexOfKey != -1 {
			message = "<https://bitbucket.org/rappinc/rappi/pull-requests/" + cfg.PR + "|El PR de " + jsonDataArray[indexOfKey].Key + "> está tocando algunos modulos que no son de su propiedad:\n"

		} else {
			message = "<https://bitbucket.org/rappinc/rappi/pull-requests/" + cfg.PR + "|El PR " + cfg.PR + "> está tocando algunos modulos que no son de su propiedad:\n"
		}
		for key, folders := range foldersTouchedByProject {
			var affectedIndex = getIndexOfKeyProject(jsonDataArray, key)
			responsible := jsonDataArray[affectedIndex].SlackResponsible
			if len(folders) > 1 {
				message += "*" + key + "*: modulos afectados: [*" + strings.Join(folders, ", ") + "*] cc:" + strings.Join(responsible, ", ") + "\n"
			} else {
				message += "*" + key + "*: modulo afectado: *" + strings.Join(folders, ", ") + "* cc:" + strings.Join(responsible, ", ") + "\n"
			}
		}
		if err := tools.ExportEnvironmentWithEnvman("ALERT_MESSAGE", message); err != nil {
			failf("error exporting variable", err)
		}
	} else {
		fmt.Printf("No folders touched\n")
		os.Exit(-1)
	}

	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}

func fillFoldersTouchedByProject(arrayOfFolders []string, jsonDataArray []Responsible, indexOfKey int, foldersTouchedByProjectResult map[string][]string) {
	for _, folder := range arrayOfFolders {
		indexOfKeyTouched := getIndexOfFolder(jsonDataArray, folder)
		if indexOfKeyTouched != -1 && indexOfKeyTouched != indexOfKey {
			key := jsonDataArray[indexOfKeyTouched].Key
			if _, ok := foldersTouchedByProjectResult[key]; ok {
				foldersTouchedByProjectResult[key] = []string{folder}
			} else {
				foldersTouchedByProjectResult[key] = append(foldersTouchedByProjectResult[key], folder)
			}
		}
	}
}

func getIndexOfKeyProject(jsonDataArray []Responsible, branchKey string) int {
	for i := 0; i < len(jsonDataArray); i++ {
		if jsonDataArray[i].Key == branchKey {
			return i
		}
	}
	return -1
}

func getIndexOfFolder(jsonDataArray []Responsible, folder string) int {
	for i := 0; i < len(jsonDataArray); i++ {
		projectKey := stringInArray(folder, jsonDataArray[i].Modules)
		if projectKey != "" {
			return i
		}
	}
	fmt.Println("Folder Not founded: ", folder)
	return -1
}

func removeDuplicateValues(stringSlice []string) []string {
	keys := make(map[string]bool)
	var list []string

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func extraBranchKey(branch string, allowedKeys string) string {
	dividerBySlashPath := strings.Split(branch, "/")
	allowedKeysSeparated := strings.Split(allowedKeys, "|")
	if len(dividerBySlashPath) > 2 {
		var key = stringContainsInArray(dividerBySlashPath[1], allowedKeysSeparated)
		if key != "" {
			return key
		} else {
			failf("Key %s in branch %s don't founded in allowed keys: %s", dividerBySlashPath[1], branch, allowedKeys)
		}
	}
	return ""
}

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func stringContainsInArray(branchPart string, allowedKeysSeparated []string) string {
	for _, allowedKey := range allowedKeysSeparated {
		if strings.Contains(normalize(branchPart), normalize(allowedKey)) {
			return allowedKey
		}
	}
	return ""
}

func stringInArray(stringToCheck string, arrayContained []string) string {
	for _, value := range arrayContained {
		if normalize(stringToCheck) == normalize(value) {
			return value
		}
	}
	return ""
}

func normalize(stringToCheck string) string {
	return strings.TrimSpace(strings.ToLower(stringToCheck))
}
