/**
* (C) 2026 Ford Motor Company
*
* All files and artifacts are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package main

import (
	"fmt"
	"encoding/json"
	"github.com/akamensky/argparse"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"strconv"
//	"os/signal"
	"syscall"
	"time"
)

type FeederElem struct {
	Name    string `json:"name"`
	SearchPath string `json:"searchpath"`
	CliParams string `json:"cliparams"`
	Pid string `json:"pid"`
	FeederType string `json:"feedertype"`
	Status string `json:"feederstatus"`
}
var feederList []FeederElem

var feederListFile string //searchpath to feeder list file

func readFeederList(feederListFile string) []FeederElem {
	var feederList []FeederElem
	data, err := os.ReadFile(feederListFile)
	if err != nil {
		fmt.Printf("Could not open %s for reading feeder list\n", feederListFile)
		return nil
	}
	err = json.Unmarshal([]byte(data), &feederList)
	if err != nil {
		fmt.Printf("readFeederList:Error unmarshal json=%s\n", err)
		return nil
	}
	return feederList
}

func saveFeederList() {
	jsonFeederList, err := json.Marshal(feederList)
	if err != nil {
		fmt.Printf("saveFeederList: Feeder list syntax error: %s\n", err)
		return
	}
	err = os.WriteFile(feederListFile, jsonFeederList, 0644)
	if err != nil {
		fmt.Printf("saveFeederList: write error: %s\n", err)
		return
	}
}

func startFeeders(feederIndex int) {
	var firstIndex, lastIndex int
	if feederIndex == -1 {
		firstIndex = 0
		lastIndex = len(feederList)
	} else if feederIndex >= 0 && feederIndex < len(feederList) {
		firstIndex = feederIndex
		lastIndex = feederIndex + 1
	} else {
		fmt.Printf("Feeder list index=%d is not on list\n", feederIndex)
		return
	}
	for i := firstIndex; i < lastIndex; i++ {
		if  feederList[i].Pid != "0" {
			dispatchTerminateMessage(i)
			feederList[i].Pid = "0"
			time.Sleep(100*time.Millisecond)
		}
		if feederList[i].Status == "active" {
			pid, syntaxError, executeError := launchFeeder(feederList[i].Name, feederList[i].SearchPath, feederList[i].CliParams)
			fmt.Printf("PID = %s\n", pid)
			if len(syntaxError) > 0 {
				fmt.Printf("syntaxError = %s\n", syntaxError)
			}
			if executeError != nil {
				fmt.Printf("executeError = %s\n", executeError)
			} else {
				feederList[i].Pid = pid
			}
		}
	}
	saveFeederList()
}

func launchFeeder(feederName string, searchPath string, cliParams string) (string, string, error) {
	var stdout strings.Builder
	var stderr strings.Builder
	command := "./" + filepath.Base(searchPath) + " " + cliParams + " > ./logs/" + feederName + "-log.txt &"
fmt.Printf("startFeeder: command: %s\n", command)
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = filepath.Dir(searchPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Could not execute the command: ./%s %s\n", searchPath, cliParams)
		return "", "", err	
	}
	time.Sleep(5*time.Second) // wait to see if feeder succeeds in registration
	pid := getPid(filepath.Base(searchPath))
	if len(pid) == 0 || isDuplicatePid(pid) {
		pid = "0"
	}
	return pid, stderr.String(), err
}

func isDuplicatePid(pid string) bool {
	for i := 0; i < len(feederList); i++ {
		if feederList[i].Pid == pid {
			return true
		}
	}
	return false
}

func getPid(processname string) string {
	var stdout strings.Builder
	command := "pgrep " + processname
//fmt.Printf("getPid: command: %s\n", command)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Run()
	pidString := stdout.String()
	if len(pidString) == 0 {
		return ""
	}
	pids := strings.Split(pidString, "\n")  //"pid1\npid2\n"
//fmt.Printf("getPid: pid: %s\n", pids[len(pids)-2])
	return pids[len(pids)-2]
}

func uiEngine() {
	var commandNumber string
	for {
		displayUiOptions()
		fmt.Scanf("%s\n", &commandNumber)
		switch commandNumber {
			case "0": return
			case "1": displayFeederListSummary()
			case "2": addFeeder()
			case "3": removeFeeder()
			case "4": startFeeder()
			case "5": terminateFeeder()
			case "6": changeFeederStatus()
			default: 
				fmt.Printf("Selected option not supported\n")
		}
	}
}

func displayUiOptions() {
	fmt.Printf("\n\nSelect one of the following numbers:\n")
	fmt.Printf("1: Display feederlist summary\n")
	fmt.Printf("2: Add feeder to list\n")
	fmt.Printf("3: Remove feeder from list\n")
	fmt.Printf("4: Start feeder on list\n")
	fmt.Printf("5: Terminate running feeder on list\n")
	fmt.Printf("6: Change feeder status\n")
	fmt.Printf("0: Exit feederManager\n")
	fmt.Printf("\nOption number selected: ")
}

func displayFeederListSummary() {
	fmt.Printf("\n  Name, PID, Status, Type, Search path, CLI params\n")
	for i := 0; i < len(feederList); i++ {
		fmt.Printf("%d: %s, %s, %s, %s, %s, %s\n", i+1, feederList[i].Name, feederList[i].Pid, feederList[i].Status, feederList[i].FeederType, feederList[i].SearchPath, feederList[i].CliParams)
	}
}

func removeFeeder() {
	var feederNo int
	displayFeederListSummary()
	fmt.Printf("\nSelect feeder index from the summary: ")
	fmt.Scanf("%d\n", &feederNo)
	if feederNo < 1 || feederNo > len(feederList)  {
		fmt.Printf("Selected feeder index not on the list\n")
		return
	}
	feederNo--
	if feederList[feederNo].Pid != "0" {
		fmt.Printf("Selected feeder must be terminated before being removed\n")
		return
	}
	if feederList[feederNo].FeederType == "static" {
		fmt.Printf("The feederManager does not have permission to remove the selected feeder\n")
		return
	}
	feederList = append(feederList[:feederNo], feederList[feederNo+1:]...)
	displayFeederListSummary()
}

func addFeeder() {
	var newFeeder FeederElem
	fmt.Printf("\nName of the new feeder: ")
	fmt.Scanf("%s\n", &newFeeder.Name)
	if isNameOnList(newFeeder.Name)  {
		fmt.Printf("Provided feeder name is already on the list\n")
		return
	}
	fmt.Printf("\nName of the new feeder binary (incl. path relative to feederManager directory): ")
	fmt.Scanf("%s\n", &newFeeder.SearchPath)
	fmt.Printf("\nCLI params for the command line start of the feeder (Return if none): ")
	fmt.Scanf("%s\n", &newFeeder.CliParams)
	var binaryOption rune
	fmt.Printf("\nShall the feeder status be 'active'/'passive' (a/p): ")
	fmt.Scanf("%c\n", &binaryOption)
	if binaryOption == 'a' {
		newFeeder.Status = "active"
	} else {
		newFeeder.Status = "passive"
	}
	fmt.Printf("\nShall the feeder type be 'dynamic'/'static' (d/s): ")
	fmt.Scanf("%c\n", &binaryOption)
	if binaryOption == 'd' {
		newFeeder.FeederType = "dynamic"
	} else {
		newFeeder.FeederType = "static"
	}
	newFeeder.Pid = "0"
	feederList = append(feederList, newFeeder)
	saveFeederList()
	displayFeederListSummary()
}

func isNameOnList(feederName string) bool {
	for i := 0; i < len(feederList); i++ {
		if feederList[i].Name == feederName {
			return true
		}
	}
	return false
}

func selectFeeder() int {
	var feederNo int
	displayFeederListSummary()
	fmt.Printf("\nSelect feeder index from the summary: ")
	fmt.Scanf("%d\n", &feederNo)
	if feederNo < 1 || feederNo > len(feederList)  {
		fmt.Printf("Selected feeder index not on the list\n")
		feederNo = -1
	}
	feederNo--
	return feederNo
}

func startFeeder() {
	feederNo := selectFeeder()
	if feederNo < 0 {
		return
	}
	if feederList[feederNo].Pid != "0" {
		fmt.Printf("Selected feeder is already running\n")
		return
	}
	startFeeders(feederNo)
	saveFeederList()
}

func terminateFeeder() {
	feederNo := selectFeeder()
	if feederNo < 0 {
		return
	}
	if feederList[feederNo].Pid == "0" {
		fmt.Printf("Selected feeder is not running\n")
		return
	}
	if feederList[feederNo].FeederType == "static" {
		fmt.Printf("The feederManager does not have permission to terminate the selected feeder\n")
		return
	}
	dispatchTerminateMessage(feederNo)
	saveFeederList()
}

func changeFeederStatus() {
	feederNo := selectFeeder()
	if feederNo < 0 {
		return
	}
	if feederList[feederNo].Status == "active" {
		fmt.Printf("Selected feeder is set to 'passive'\n")
		feederList[feederNo].Status = "passive"
	} else {
		fmt.Printf("Selected feeder is set to 'active'\n")
		feederList[feederNo].Status = "active"
	}
	saveFeederList()
}

func buildFeeders() {
	for i := 0; i < len(feederList); i++ {
		if feederList[i].Status == "active" {
			fmt.Printf("Building %s...\n", feederList[i].SearchPath)
			err := buildFeeder(feederList[i].SearchPath)
			if err != nil {
				fmt.Printf("Build of %s failed with error=s\n", feederList[i].SearchPath, err)
			} else {
				fmt.Printf("%s built successfully\n", feederList[i].SearchPath)
			}
		}
	}
}

func buildFeeder(searchPath string) error {
	var stdout strings.Builder
	command := "go build -o " + filepath.Base(searchPath)
fmt.Printf("buildFeeder: command: %s\n", command)
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Dir = filepath.Dir(searchPath)
	err := cmd.Run()
	return err
}

func terminateFeeders() {
	for i := 0; i < len(feederList); i++ {
		if feederList[i].Status == "active" && feederList[i].Pid != "0" {
			dispatchTerminateMessage(i)
		}
	}
	saveFeederList()
}

func dispatchTerminateMessage(index int) {
	pid, _ := strconv.Atoi(feederList[index].Pid)
	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		fmt.Printf("dispatchTerminateMessage: sending message to PID=%s failed with error=%s\n", feederList[index].Pid, err)
	} else {
		fmt.Printf("%s with PID=%s is terminated\n", feederList[index].Name, feederList[index].Pid)
	}
	feederList[index].Pid = "0"
}

func main() {
	// Create new parser object
	parser := argparse.NewParser("print", "Feeder manager")
	startflag := parser.Flag("s", "start", &argparse.Options{Required: false, Help: "start all feeders"})
	buildflag := parser.Flag("b", "build", &argparse.Options{Required: false, Help: "build active feeders"})
	terminateflag := parser.Flag("t", "terminate", &argparse.Options{Required: false, Help: "terminate running feeders"})
	feederLF := parser.String("f", "feederfile", &argparse.Options{
		Required: false,
		Help:     "Feeder-list filename",
		Default:  "feederList.json"})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	feederListFile = *feederLF
	feederList = readFeederList(feederListFile)

	if *startflag {
		startFeeders(-1)
	} else if *buildflag {
		buildFeeders()
	} else if *terminateflag {
		terminateFeeders()
	} else {
		uiEngine()
	}
}
