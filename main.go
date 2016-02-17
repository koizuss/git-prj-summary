package main

import (
	"fmt"
	"os/exec"
	"os"
	"bufio"
	"regexp"
	"flag"
	"sort"
	"strings"
)

type Pair struct {
	Key string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p)  }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i]  }
func (p PairList) Less(i, j int) bool { return p[i].Value > p[j].Value  }

func main() {
	debug := flag.Int("d", 0, "debug mode [0|1|2]")
	workDir := flag.String("t", ".", "working dir")
	excludes := flag.String("x", "", "exclude path")
	authors := flag.String("a", "", "exclude path")

	flag.Parse()

	excludeList := []string{}
	for _, e := range strings.Split(*excludes, ",") {
		if len(e) > 0 {
			excludeList = append(excludeList, strings.TrimSpace(e))
		}
	}
	if *debug > 1 { fmt.Println(excludeList) }

	authMap := make(map[string]string)
	for _, a := range strings.Split(*authors, ",") {
		if len(a) > 0 {
			author := strings.Split(strings.TrimSpace(a), ":")
			authMap[author[0]] = author[1]
		}
	}

	gitGrep := exec.Command("git", "grep", "-Il", "\\'\\'")
	rgxFindAuth := regexp.MustCompile(`^[^\(]*\((.*) \d{4,4}-\d{2,2}-\d{2,2}[^a-zA-Z]+\).*$`)
	summay := make(map[string]int)
	all := 0

	gitGrep.Dir = *workDir
	if *debug > 1 { fmt.Println(gitGrep.Dir) }

	stdout, err := gitGrep.StdoutPipe()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	gitGrep.Start()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		file := scanner.Text()
		if *debug > 0 { fmt.Println(file) }

		exclude := false
		for _, e := range excludeList {
			if strings.Contains(file, e) {
				if *debug > 0 { fmt.Printf("skip %s %s\n", file, e) }
				exclude = true
				break
			}
		}
		if exclude { continue }

		gitBlame := exec.Command("git", "blame", file)

		gitBlame.Dir = *workDir
		if *debug > 1 { fmt.Println(gitBlame.Dir) }

		stdout2, err2 := gitBlame.StdoutPipe()

		if err2 != nil {
			fmt.Println(err2)
			os.Exit(1)
		}

		gitBlame.Start()

		scanner2 := bufio.NewScanner(stdout2)
		for scanner2.Scan() {
			line := scanner2.Text()
			if *debug > 1 { fmt.Println(line) }
			auth := strings.TrimSpace(rgxFindAuth.ReplaceAllString(line, "$1"))
			_, ok1 := authMap[auth]
			if ok1 {
				auth = authMap[auth]
			}

			if *debug > 1 { fmt.Println(auth) }
			_, ok2 := summay[auth]
			if ok2 {
				summay[auth] = summay[auth] + 1
			} else {
				summay[auth] = 1
			}
			all = all + 1
		}

		gitBlame.Wait()
	}

	gitGrep.Wait()

	p := make(PairList, len(summay))
	i := 0
	for k, v := range summay {
		p[i] = Pair{k, v}
		i++
	}
	sort.Sort(p)

	for _, k := range p {
		avg := float64(k.Value) / float64(all) * 100
		fmt.Printf("%s\t: %d (%6.2f%%)", k.Key, k.Value, avg)
		fmt.Println()
	}
}
