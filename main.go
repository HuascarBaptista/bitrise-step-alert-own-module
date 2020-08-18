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
	/*var cfg Config = Config{
		Folders:           "tools\nbasket\nbase\n",
		PathConfiguration: "tools/responsible.json",
		Branch:            "fix/SHP-22/huascar",
		AllowedKeys:       "BAS|SHP|OT",
	}*/
	var cfg Config
	if err := stepconf.Parse(&cfg); err != nil {
		failf("Issue with input: %s", err)
	}
	fmt.Printf("The Path configuration is %s\n", cfg.PathConfiguration)

	file, _ := ioutil.ReadFile(cfg.PathConfiguration)

	var jsonDataArray []Responsible

	_ = json.Unmarshal(file, &jsonDataArray)

	branchKey := extraBranchKey(cfg.Branch, cfg.AllowedKeys)

	if branchKey == "" {
		failf("Key don't allowed")
	}

	var arrayOfFolders = strings.Fields(cfg.Folders)

	var indexOfKey = getIndexOfKeyProject(jsonDataArray, branchKey)

	if indexOfKey == -1 {
		failf("Not founded key: %s", branchKey)
	}

	foldersTouchedByProject := map[string][]string{}

	fillFoldersTouchedByProject(arrayOfFolders, jsonDataArray, indexOfKey, foldersTouchedByProject)
	if len(foldersTouchedByProject) > 0 {
		var message = "El PR " + cfg.PR + " está tocando algunas carpetas que no son de su modulo\n"
		for key, folders := range foldersTouchedByProject {
			var affectedIndex = getIndexOfKeyProject(jsonDataArray, key)
			responsible := jsonDataArray[affectedIndex].SlackResponsible
			message += "\n Guys " + strings.Join(responsible, ", ") + " estás carpetas [" + strings.Join(folders, ", ") + "] están siendo tocadas de su proyecto " + key + "\n"
			fmt.Printf("The Project affected was %s\n", key)
			fmt.Printf("The folders affected were %s\n", folders)
			fmt.Printf("The authors affected were %s\n", responsible)
		}
		fmt.Printf("Mensaje FINAL\n%s\n", message)

		if err := tools.ExportEnvironmentWithEnvman("ALERT_MESSAGE", message); err != nil {
			failf("error exporting variable", err)
		}
	} else {
		os.Exit(1)
	}

	// You can find more usage examples on envman's GitHub page
	//  at: https://github.com/bitrise-io/envman

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}

func fillFoldersTouchedByProject(arrayOfFolders []string, jsonDataArray []Responsible, indexOfKey int, foldersTouchedByProject map[string][]string) {
	for _, folder := range arrayOfFolders {
		indexOfKeyTouched := getIndexOfFolder(jsonDataArray, folder)
		if indexOfKeyTouched != indexOfKey && indexOfKeyTouched != -1 {
			key := jsonDataArray[indexOfKeyTouched].Key
			fmt.Printf("The Folder %s is property of the project %s\n", folder, key)
			if _, ok := foldersTouchedByProject[key]; ok {
				foldersTouchedByProject[key] = []string{folder}
			} else {
				foldersTouchedByProject[key] = append(foldersTouchedByProject[key], folder)
			}
			fmt.Printf("Folders of %s touched %s\n", key, foldersTouchedByProject[key])
		}
	}
}

func getIndexOfKeyProject(jsonDataArray []Responsible, branchKey string) int {
	for i := 0; i < len(jsonDataArray); i++ {
		fmt.Println("getIndexOfKeyProject Key: ", jsonDataArray[i].Key)
		if jsonDataArray[i].Key == branchKey {
			fmt.Println("Key founded: ", jsonDataArray[i].Key)
			return i
		}
	}
	fmt.Println("Key not founded: ", branchKey)
	return -1
}

func getIndexOfFolder(jsonDataArray []Responsible, folder string) int {
	for i := 0; i < len(jsonDataArray); i++ {
		fmt.Println("getIndexOfFolder folder: ", folder)
		fmt.Println("getIndexOfFolder loop Key to check: ", jsonDataArray[i].Key)
		projectKey := stringInArray(folder, jsonDataArray[i].Modules)
		if projectKey != "" {
			fmt.Println("folder founded in: ", jsonDataArray[i].Key)
			return i
		}
	}
	fmt.Println("Folder Not founded: ", folder)
	return -1
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
		if strings.Contains(branchPart, allowedKey) {
			return allowedKey
		}
	}
	return ""
}

func stringInArray(stringToCheck string, arrayContained []string) string {
	for _, value := range arrayContained {
		if stringToCheck == value {
			return value
		}
	}
	return ""
}
