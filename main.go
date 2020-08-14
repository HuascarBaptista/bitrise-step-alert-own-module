package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
)

type AutoGenerated []struct {
	Key              string   `json:"key"`
	Modules          []string `json:"modules"`
	SlackResponsible []string `json:"slack_responsible"`
}

func convertByteArrayToStringArray(b []byte) []string {
	s := make([]string, len(b))
	for i := range b {
		s[i] = strconv.Itoa(int(b[i]))
	}
	return s
}

func main() {
	command := []string{
		"git",
		"diff",
		"--dirstat=files,0",
		"HEAD~1",
		"|",
		"sed",
		"-E",
		"'s/^[ 0-9.]+% //g'",
		"|",
		"sed",
		"-E",
		"'s/\\/.*$//g'",
	}
	fmt.Printf("The commands %s\n", command)

	cmd := exec.Command(command[0], command[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The Git diff is %s\n", out)
	var arregloDeCarpetas = convertByteArrayToStringArray(out)

	for i := 0; i < len(out); i++ {
		fmt.Println("Folder: ", out[i])
	}

	file, _ := ioutil.ReadFile("tools/responsible.json")

	data := AutoGenerated{}

	_ = json.Unmarshal(file, &data)

	for i := 0; i < len(data); i++ {
		fmt.Println("Key: ", data[i].Key)
	}

	for i := 0; i < len(arregloDeCarpetas); i++ {
		fmt.Println("Folder two: ", arregloDeCarpetas[i])
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
