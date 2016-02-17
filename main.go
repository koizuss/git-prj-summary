package main

import (
	"fmt"
	"os/exec"
	"os"
	"bufio"
)

func main() {
	cmd := exec.Command("git", "grep", "-Il", "''")

	// TODO: setup working dir
	cmd.Dir = "/Users/kiyoshiro/src/bitbucket.org/koizuss/cost3"
	fmt.Println(cmd.Dir)

	stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cmd.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	cmd.Wait()
}
