package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	out, err := exec.Command("ls", "-la").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The list folder is %s\n", out)

	out2, err2 := exec.Command("pwd").Output()
	if err2 != nil {
		log.Fatal(err2)
	}
	fmt.Printf("The path folder is %s\n", out2)

	// You can find more usage examples on envman's GitHub page
	//  at: https://github.com/bitrise-io/envman

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
