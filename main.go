package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func GetCurrentDirectory() string {
	//返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	//将\替换成/
	return strings.Replace(dir, "\\", "/", -1)
}

func getBranchDesc(showBranchResult [][]string, index int, branch string) {
	defer wg.Done()

	var descText string
	cleanDesc := strings.TrimSpace(strings.Trim(branch, "*"))
	branchDesc := "branch." + cleanDesc + ".description"
	cmd := exec.Command("git", "config", branchDesc)
	// pwd, _ := os.Getwd()
	cmd.Dir = GetCurrentDirectory()
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

	showBranchResult := make([][]string, len(allBranchList))

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
