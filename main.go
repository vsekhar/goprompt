package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

var debug = flag.Bool("debug", false, "debug mode")

func DebugLogf(s string, v ...interface{}) {
	if *debug {
		log.Printf(s, v...)
	}
}

func Fatal(e error) {
	fmt.Printf("goprompt-err(%s)", e)
	os.Exit(1)
}

var collapsablePrefixes = map[string]string{
	"~/Code/go/src/": "",
}

func main() {
	flag.Parse()
	var prompt string

	cwd, err := os.Getwd()
	if err != nil {
		Fatal(err)
	}
	displayPath := cwd

	// home prefix
	home, err := homedir.Dir()
	if err != nil {
		Fatal(err)
	}
	if strings.HasPrefix(displayPath, home) {
		displayPath = "~" + strings.TrimPrefix(displayPath, home)
	}

	// other collapsable prefixes
	for k, v := range collapsablePrefixes {
		if strings.HasPrefix(displayPath, k) {
			displayPath = v + strings.TrimPrefix(displayPath, k)
		}
	}
	prompt += displayPath

	// git
	ccwd := cwd
	isGit := false
	for {
		gitPath := path.Join(ccwd, ".git")
		DebugLogf("Checking for %s", gitPath)
		s, err := os.Stat(gitPath)
		if err == nil && s.IsDir() {
			isGit = true
			DebugLogf("Found git dir: %s", gitPath)
			break
		}
		if ccwd == "/" {
			break
		}
		ccwd = path.Dir(ccwd)
	}
	if isGit {
		cmd := exec.Command("git", "status", "--porcelain", "--branch")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			Fatal(err)
		}
		reader := bufio.NewReader(stdout)
		branch, err := reader.ReadString('\n')
		if err != nil {
			Fatal(err)
		}
		branch = strings.TrimPrefix(branch, "## ")
		branch = strings.SplitN(branch, "...", 2)[0]

		scanner := bufio.NewScanner(reader)
		untracked := false
		isDirty := false   // work tree != index
		isPending := false // index != HEAD
		for scanner.Scan() {
			line := string(scanner.Text())
			if strings.HasPrefix(line, "?? ") {
				untracked = true
				continue
			}
			if line[0] != ' ' {
				isPending = true
			}
			if line[1] != ' ' {
				isDirty = true
			}
		}
		prompt += ":"
		colorCode := ""
		switch {
		case isPending:
			colorCode = "\x1b[33;1m" // yellow
		case isDirty:
			colorCode = "\x1b[31;1m" // red
		default:
			colorCode = "\x1b[32;1m" // green
		}
		prompt += colorCode
		prompt += branch
		if untracked {
			prompt += "+?"
		}
		prompt += "\x1b[0m"
	}

	// clear formatting for sure so we don't mess up the terminal
	prompt += "\x1b[0m"

	fmt.Printf("%s", prompt)
}
