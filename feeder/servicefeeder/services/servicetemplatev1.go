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
	"github.com/covesa/vissr/utils"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ConfigData struct {
	ServicePath    string `json:"servicepath"` // array of path(s)
	InfoType string `json:"infotype"`
}

func feederRegistration(action string, configData ConfigData) string {
	conn, err := net.Dial("unix", "/var/tmp/vissv2/serviceFeederReg.sock")
	if err != nil {
		utils.Error.Printf("feederRegistration:Failed to UDS connect to the server. Err=%s", err)
		return ""
	}
	request := `{"action": "` + action + `", "service": ` + configData.ServicePath + `, "infotype": "` + configData.InfoType + `"}`
	_, err = conn.Write([]byte(request))
	if err != nil {
		utils.Error.Printf("feederRegistration:Write failed, err = %s", err)
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		utils.Error.Printf("feederRegistration:Read failed, err = %s", err)
		return ""
	}
	utils.Info.Printf("feederRegistration:Reg response from server: %s", string(buf[:n]))
	var responseMap map[string]interface{}
	err = json.Unmarshal(buf[:n], &responseMap)
	if err != nil {
		utils.Error.Printf("feederRegistration:Unmarshal error=%s", err)
		return ""
	}
	conn.Close()
	if responseMap["action"].(string) == "error" {
		utils.Error.Printf("feederRegistration:Server responded with error")
		return ""
	}
	if action == "dereg" {
		return ""
	}
	return responseMap["sockfile"].(string)
}

func main() {
	// Create new parser object
	parser := argparse.NewParser("print", "Service template version 1")
	logFile := parser.Flag("", "logfile", &argparse.Options{Required: false, Help: "outputs to logfile in ./logs folder"})
	logLevel := parser.Selector("", "loglevel", []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}, &argparse.Options{
		Required: false,
		Help:     "changes log output level",
		Default:  "info"})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	utils.InitLog("service-log.txt", "./logs", *logFile, *logLevel)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	feederConfig := ConfigData{ ServicePath: `["VehicleService.Seating.Row1.DriverSide.MoveSeat", "VehicleService.Seating.Row1.PassengerSide.MoveSeat",
					["VehicleService.Seating.Row1.DriverSide.MoveSeat", "VehicleService.Seating.Row1.PassengerSide.MoveSeat"]`,
					InfoType: "Service" }
	ioSocket := feederRegistration("reg", feederConfig)
	if len(ioSocket) == 0 {
		utils.Error.Printf("Registration failed for service=%s",feederConfig.ServicePath)
		os.Exit(-1)
	}

	for {
		select {
		case sig := <- sigChan:
			if sig == syscall.SIGUSR1 {
				feederRegistration("dereg", feederConfig)
				time.Sleep(1 * time.Second)
				os.Exit(1)
			} else {
				utils.Info.Printf("Received unknown signal=%d",sig)
			}
		}
	}
}
