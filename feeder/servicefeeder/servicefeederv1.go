/**
* (C) 2024 Ford Motor Company
*
* All files and artifacts are licensed under the provisions of the license provided by the LICENSE file in this repository.
*
**/

package main

import (
	"fmt"
	"database/sql"
	"encoding/json"
	"github.com/akamensky/argparse"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/covesa/vissr/utils"
	"github.com/go-redis/redis"
	_ "github.com/mattn/go-sqlite3"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"slices"
	"time"
)

type DomainData struct {
	Name  string  // service path
	Value string  // service input/output 
}

type ConfigData struct {
	Name    string `json:"name"`
	InfoType string `json:"infotype"`
	Scope []string `json:"scope"`
}

type ServiceConfigData struct {
	Name    string `json:"name"`
	ServiceType string `json:"servicetype"`
}

var redisClient *redis.Client
var memcacheClient *memcache.Client
var dbHandle *sql.DB
var stateDbType string

var notificationList []string

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readFeederConfig(configFilename string) ConfigData {
	var configData ConfigData
	data, err := os.ReadFile(configFilename)
	if err != nil {
		utils.Error.Printf("Could not open %s for reading feeder config data", configFilename)
		configData.InfoType = "error"
		return configData
	}
	err = json.Unmarshal([]byte(data), &configData)
	if err != nil {
		utils.Error.Printf("readFeederConfig:Error unmarshal json=%s", err)
		configData.InfoType = "error"
		return configData
	}
	return configData
}

func feederRegister(regSockFile string, configData ConfigData) string {
	conn, err := net.Dial("unix", regSockFile)
	if err != nil {
		utils.Error.Printf("feederRegister:Failed to UDS connect to the server. Err=%s", err)
		return ""
	}
	request := `{"action": "reg", "name": "` + configData.Name + `", "infotype": "` + configData.InfoType + `"}`
	_, err = conn.Write([]byte(request))
	if err != nil {
		utils.Error.Printf("feederRegister:Write failed, err = %s", err)
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		utils.Error.Printf("feederRegister:Read failed, err = %s", err)
		return ""
	}
	utils.Info.Printf("feederRegister:Reg response from server: %s", string(buf[:n]))
	var responseMap map[string]interface{}
	err = json.Unmarshal(buf[:n], &responseMap)
	if err != nil {
		utils.Error.Printf("feederRegister:Unmarshal error=%s", err)
		return ""
	}
	conn.Close()
	if responseMap["action"].(string) == "error" {
		utils.Error.Printf("feederRegister:Server responded with error")
		return ""
	}
	return responseMap["sockfile"].(string)
}

func initVSSInterfaceMgr(inputChan chan DomainData, outputChan chan DomainData, configData ConfigData) {
 	sockFile := feederRegister("/var/tmp/vissv2/feederReg.sock", configData)
	os.Remove(sockFile)
	listener, err := net.Listen("unix", sockFile)
	if err != nil {
		utils.Error.Printf("initVSSInterfaceMgr:UDS listen failed, err = %s", err)
		os.Exit(-1)
	}
	conn, err := listener.Accept()
	if err != nil {
		utils.Error.Printf("initVSSInterfaceMgr:UDS accept failed, err = %s", err)
		os.Exit(-1)
	}
	udsChan := make(chan string)
	go udsReader(conn, inputChan, udsChan, configData)
	go udsWriter(conn, udsChan)
	for {
		select {
		case outData := <-outputChan:
			if len(outData.Name) == 0 {
				continue
			}
			status := statestorageSet(outData.Name, outData.Value, utils.GetRfcTime())
			if status != 0 {
				utils.Error.Printf("initVSSInterfaceMgr():State storage write failed")
			} else {
				if onNotificationList(outData.Name) != -1 {
					message := `{"action": "subscription", "path":"` + outData.Name + `"}`
					udsChan <- message
					utils.Info.Printf("Server notified that data written to %s", outData.Name)
				}
			}
		}
	}
}

func statestorageSet(path string, val string, ts string) int {
	switch stateDbType {
	case "sqlite":
		stmt, err := dbHandle.Prepare("UPDATE VSS_MAP SET c_value=?, c_ts=? WHERE `path`=?")
		if err != nil {
			utils.Error.Printf("Could not prepare for statestorage updating, err = %s", err)
			return -1
		}
		defer stmt.Close()

		_, err = stmt.Exec(val, ts, path)
		if err != nil {
			utils.Error.Printf("Could not update statestorage, err = %s", err)
			return -1
		}
		return 0
	case "redis":
		dp := `{"value":"` + val + `", "ts":"` + ts + `"}`
		err := redisClient.Set(path, dp, time.Duration(0)).Err()
		if err != nil {
			utils.Error.Printf("Job failed. Err=%s", err)
			return -1
		}
		return 0
	case "memcache":
		dp := `{"value":"` + val + `", "ts":"` + ts + `"}`
		err := memcacheClient.Set(&memcache.Item{Key: path, Value: []byte(dp)})
		if err != nil {
			utils.Error.Printf("Job failed. Err=%s", err)
			return -1
		}
		return 0
	}
	return -1
}

func inFeederScope(testpath string, scope []string) bool {
	for i := 0; i < len(scope); i++ {
		scopepath := scope[i]
		if len(testpath) >= len(scopepath) && testpath[:len(scopepath)] == scopepath {
utils.Info.Printf("%s in scope=%s", testpath, scopepath)
			return true
		}
	}
utils.Info.Printf("%s not in scope", testpath)
	return false
}

func udsReader(conn net.Conn, inputChan chan DomainData, udsChan chan string, configData ConfigData) {
	defer conn.Close()
	buf := make([]byte, 8192)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			utils.Error.Printf("udsReader:Read failed, err = %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		utils.Info.Printf("udsReader:Message from server: %s", string(buf[:n]))
		if n > 8192 {
			utils.Error.Printf("udsReader:Max message size of 8192 chars exceeded. Message dropped")
			continue
		}
		var serverMessageMap map[string]interface{}
		err = json.Unmarshal(buf[:n], &serverMessageMap)
		if err != nil {
			utils.Error.Printf("udsReader:Unmarshal error=%s", err)
			continue
		}
		if serverMessageMap["action"] != nil {
			switch serverMessageMap["action"].(string) {
				case "set":
					domainData, _ := splitToDomainDataAndTs(serverMessageMap["data"].(map[string]interface{}))
					if inFeederScope(domainData.Name, configData.Scope) {
						inputChan <- domainData
					}
				case "subscribe":
					pathList := serverMessageMap["path"].([]interface{})
					for i := 0; i < len(pathList); i++ {
						if onNotificationList(pathList[i].(string)) == -1 {
							notificationList = append(notificationList, pathList[i].(string))
						}
					}
					response := `{"action": "subscribe", "status": "ok"}`
					udsChan <- response
				case "unsubscribe":
					pathList := serverMessageMap["path"].([]interface{})
					for i := 0; i < len(pathList); i++ {
						if onNotificationList(pathList[i].(string)) != -1 {
							notificationList = slices.Delete(notificationList, i, i+1)
						}
					}
				case "update":
					defaultArray := serverMessageMap["defaultList"].([]interface{})
					var domainData DomainData
					for i := 0; i < len(defaultArray); i++ {
						for k, v := range defaultArray[i].(map[string]interface{}) {
//							utils.Info.Printf("%d: key=%s, value=%s", i, k, v.(string))
							if k == "path" {
								domainData.Name = v.(string)
							} else if k == "default" {
								domainData.Value = v.(string)
							}
						}
						statestorageSet(domainData.Name, domainData.Value, utils.GetRfcTime())
					}
				default:
					utils.Error.Printf("udsReader:Message action unknown = %s", serverMessageMap["action"].(string))
			}
		}
	}
}

func udsWriter(conn net.Conn, udsChan chan string) {
	defer conn.Close()
	for {
		select {
		case message := <-udsChan:
		utils.Info.Printf("udsWriter:Message to server: %s", message)
			_, err := conn.Write([]byte(message))
			if err != nil {
				utils.Error.Printf("udsWriter:Write failed, err = %s", err)
			}
		}
	}
}

func onNotificationList(path string) int {
	for i := 0; i < len(notificationList); i++ {
		if notificationList[i] == path {
			return i
		}
	}
	return -1
}

func splitToDomainDataAndTs(serverMessageMap map[string]interface{}) (DomainData, string) { // server={"dp": {"ts": "Z","value": "Y"},"path": "X"}, redis={"value":"xxx", "ts":"zzz"}
	var domainData DomainData
	domainData.Name = serverMessageMap["path"].(string)
	dpMap := serverMessageMap["dp"].(map[string]interface{})
	domainData.Value = dpMap["value"].(string)
	return domainData, dpMap["ts"].(string)
}

func initServiceHub(inputChan chan DomainData, outputChan chan DomainData, sigChan chan os.Signal) {
	serviceChan := make(chan DomainData)
	go serviceMgr(serviceChan, sigChan)
	for {
		select {
		case input := <-outputChan:
			utils.Info.Printf("Data for calling a service: Name=%s, Input=%s", input.Name, input.Value)
		case output := <- serviceChan:
			utils.Info.Printf("Data from a called  service: Name=%s, Output=%s", output.Name, output.Value)
			inputChan <- output
		}
	}
}

func serviceMgr(serviceChan chan DomainData, sigChan chan os.Signal) {
	// läs in servicelistan, 
	for {
		select {
		case sig := <- sigChan:
			if sig == syscall.SIGUSR1 {
				utils.Info.Printf("SIGUSR1 received")
				// updateServicesList()
			} else {
				utils.Info.Printf("Unknown signal=%d received", sig)
			}
		}
	}
}

func main() {
	// Create new parser object
	parser := argparse.NewParser("print", "Service feeder template version 1")
	configFile := parser.String("c", "configfile", &argparse.Options{
		Required: false,
		Help:     "Feeder configuration filename",
		Default:  "feederConfig.json"})
	logFile := parser.Flag("", "logfile", &argparse.Options{Required: false, Help: "outputs to logfile in ./logs folder"})
	logLevel := parser.Selector("", "loglevel", []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}, &argparse.Options{
		Required: false,
		Help:     "changes log output level",
		Default:  "info"})
	stateDB := parser.Selector("d", "statestorage", []string{"sqlite", "redis", "memcache", "none"}, &argparse.Options{Required: false,
		Help: "Statestorage must be either sqlite, redis, memcache, or none", Default: "redis"})
	dbFile := parser.String("f", "dbfile", &argparse.Options{
		Required: false,
		Help:     "statestorage database filename",
		Default:  "../../server/vissv2server/serviceMgr/statestorage.db"})
	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
	}

	utils.InitLog("feeder-log.txt", "./logs", *logFile, *logLevel)

	stateDbType = *stateDB

	switch stateDbType {
	case "sqlite":
		var dbErr error
		if utils.FileExists(*dbFile) {
			dbHandle, dbErr = sql.Open("sqlite3", *dbFile)
			if dbErr != nil {
				utils.Error.Printf("Could not open state storage file = %s, err = %s", *dbFile, dbErr)
				os.Exit(1)
			} else {
				utils.Info.Printf("SQLite state storage initialised.")
			}
		} else {
			utils.Error.Printf("Could not find state storage file = %s", *dbFile)
		}
	case "redis":
		redisClient = redis.NewClient(&redis.Options{
			Network:  "unix",
			Addr:     "/var/tmp/vissv2/redisDB.sock",
			Password: "",
			DB:       1,
		})
		err := redisClient.Ping().Err()
		if err != nil {
			utils.Error.Printf("Could not initialise redis DB, err = %s", err)
			os.Exit(1)
		} else {
			utils.Info.Printf("Redis state storage initialised.")
		}
	case "memcache":
		memcacheClient = memcache.New("/var/tmp/vissv2/memcacheDB.sock")
		err := memcacheClient.Ping()
		if err != nil {
			utils.Info.Printf("Memcache daemon not alive. Trying to start it")
			cmd := exec.Command("/usr/bin/bash", "memcacheNativeInit.sh")
			err := cmd.Run()
			if err != nil {
				utils.Error.Printf("Memcache daemon startup failed, err=%s", err)
				os.Exit(1)
			}
		}
		utils.Info.Printf("Memcache daemon alive.")
	default:
		utils.Error.Printf("Unknown state storage type = %s", stateDbType)
		os.Exit(1)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGUSR1)

	vssInputChan := make(chan DomainData, 3)
	vssOutputChan := make(chan DomainData, 3)
	vehicleInputChan := make(chan DomainData, 3)
	vehicleOutputChan := make(chan DomainData, 3)

	feederConfig := readFeederConfig(*configFile)
	go initVSSInterfaceMgr(vssInputChan, vssOutputChan, feederConfig)
	go initServiceHub(vehicleInputChan, vehicleOutputChan, sigChan)

	for {
		select {
		case vssInData := <-vssInputChan:
			vehicleOutputChan <- vssInData
		case vehicleInData := <-vehicleInputChan:
			vssOutputChan <- vehicleInData
		}
	}
}
