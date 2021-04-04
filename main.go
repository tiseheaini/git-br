package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// func main() {
// 	cmd := exec.Command("git", "branch", "--list")
// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	if err := cmd.Start(); err != nil {
// 		log.Fatal(err)
// 	}

// 	if err := cmd.Wait(); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("%v\n", stdout)
// 	//fmt.Println(reflect.TypeOf(stdout))
// }
var wg sync.WaitGroup

type Job struct {
	index  int
	branch string
}

func getBranchDesc(showBranchResult [][]string, index int, branch string) {
	defer wg.Done()

	var descText string
	cleanDesc := strings.TrimSpace(strings.Trim(branch, "*"))
	branchDesc := "branch." + cleanDesc + ".description"
	cmd := exec.Command("git", "config", branchDesc)
	pwd, _ := os.Getwd()
	cmd.Dir = pwd
	out, err := cmd.CombinedOutput()
	if err != nil {
		descText = ""
	} else {
		descText = string(out)
	}
	showBranchResult[index] = []string{branch, descText}
}

func main() {
	cpuNum := runtime.NumCPU()
	// 开启5倍核心数量的 channel
	jobs := make(chan Job, cpuNum*5)
	cmd := exec.Command("git", "branch", "--list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with2 %s\n", err)
	}
	allBranchList := strings.Split(string(out), "\n")

	var showBranchResult [][]string
	for i := 0; i < len(allBranchList); i++ {
		showBranchResult = append(showBranchResult, []string{})
	}

	go func() {
		for index, branch := range allBranchList {
			wg.Add(1)
			jobs <- Job{index, branch}
		}
		close(jobs)
	}()

	for job := range jobs {
		go getBranchDesc(showBranchResult, job.index, job.branch)
	}

	wg.Wait()

	for _, result := range showBranchResult {
		if result[0] == "" {
			continue
		}

		branchStr := result[0]
		if strings.Contains(result[0], "*") {
			branchStr = fmt.Sprintf("%c[%d;%d;%dm%s%c[0m", 0x1B, 0, 0, 32, result[0], 0x1B)
		}

		fmt.Printf(" %s %s \n", branchStr, strings.Replace(result[1], "\n", "", -1))
	}

}
