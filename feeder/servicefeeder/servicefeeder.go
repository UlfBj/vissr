/**
* (C) 2026 Ford Motor Company
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
	"strconv"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"slices"
	"time"
)
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

type ServiceChannelElem struct {
	Busy bool
	Channel chan string
}
var serviceChannelList []ServiceChannelElem
const MAXSERVICES = 25

type ServiceRegElem struct {
	Name string
	InfoType string
	SockFile string
	Conn net.Conn
	ChannelIndex int
}
var serviceList []ServiceRegElem
const SERVICE_REG_DIR = "/var/tmp/vissv2/"
const SERVICE_REG_SOCKFILE = SERVICE_REG_DIR + "serviceReg.sock"

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readServiceConfig(configFilename string) ConfigData {
	var configData ConfigData
	data, err := os.ReadFile(configFilename)
	if err != nil {
		utils.Error.Printf("Could not open %s for reading service config data", configFilename)
		configData.InfoType = "error"
		return configData
	}
	err = json.Unmarshal([]byte(data), &configData)
	if err != nil {
		utils.Error.Printf("readServiceConfig:Error unmarshal json=%s", err)
		configData.InfoType = "error"
		return configData
	}
	return configData
}

func serviceRegistration(action string, configData ConfigData) string {
	conn, err := net.Dial("unix", "/var/tmp/vissv2/serviceReg.sock")
	if err != nil {
		utils.Error.Printf("serviceRegistration:Failed to UDS connect to the server. Err=%s", err)
		return ""
	}
	request := `{"action": "` + action + `", "name": "` + configData.Name + `", "infotype": "` + configData.InfoType + `"}`
	_, err = conn.Write([]byte(request))
	if err != nil {
		utils.Error.Printf("serviceRegistration:Write failed, err = %s", err)
	}
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		utils.Error.Printf("serviceRegistration:Read failed, err = %s", err)
		return ""
	}
	utils.Info.Printf("serviceRegistration:Reg response from server: %s", string(buf[:n]))
	var responseMap map[string]interface{}
	err = json.Unmarshal(buf[:n], &responseMap)
	if err != nil {
		utils.Error.Printf("serviceRegistration:Unmarshal error=%s", err)
		return ""
	}
	conn.Close()
	if responseMap["action"].(string) == "error" {
		utils.Error.Printf("serviceRegistration:Server responded with error")
		return ""
	}
	if action == "dereg" {
		return ""
	}
	return responseMap["sockfile"].(string)
}

func initVSSInterfaceMgr(fromServerChan chan map[string]interface{}, toStorageChan chan map[string]interface{}, configData ConfigData) {
 	sockFile := serviceRegistration("reg", configData)
 	if len(sockFile) == 0 {
		utils.Error.Printf("initVSSInterfaceMgr:Registration failed, the service terminates")
		os.Exit(-1)
 	}
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
	toServerChan := make(chan string)
	go serverReader(conn, fromServerChan, toServerChan, configData)
	go serverWriter(conn, toServerChan)
	for {
		select {
		case output := <-toStorageChan:
			path := output["path"].(string)
			if len(path) == 0 {
				continue
			}
			value, _ := json.Marshal(output["value"])
			status := statestorageSet(path, string(value), utils.GetRfcTime())
			if status != 0 {
				utils.Error.Printf("initVSSInterfaceMgr():State storage write failed")
			} else {
				if onNotificationList(path) != -1 {
					message := `{"action": "subscription", "path":"` + path + `"}`
					toServerChan <- message
					utils.Info.Printf("Server notified that data written to %s", path)
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

func inserviceScope(testpath string, scope []string) bool {
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

func serverReader(conn net.Conn, fromServerChan chan map[string]interface{}, toServerChan chan string, configData ConfigData) {
	defer conn.Close()
	buf := make([]byte, 8192)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			utils.Error.Printf("serverReader:Read failed, err = %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		utils.Info.Printf("serverReader:Message from server: %s", string(buf[:n]))
		if n > 8192 {
			utils.Error.Printf("serverReader:Max message size of 8192 chars exceeded. Message dropped")
			continue
		}
		var serverMessageMap map[string]interface{}
		err = json.Unmarshal(buf[:n], &serverMessageMap)
		if err != nil {
			utils.Error.Printf("serverReader:Unmarshal error=%s", err)
			continue
		}
		if serverMessageMap["action"] != nil {
			switch serverMessageMap["action"].(string) {
				case "invoke":
					path := serverMessageMap["path"].(string)
					if inserviceScope(path, configData.Scope) {
						addToNotificationList(path)
						fromServerChan <- serverMessageMap
					}
/*				case "monitor":
					pathList := serverMessageMap["path"].([]interface{})
					for i := 0; i < len(pathList); i++ {
						addToNotificationList(pathList[i].(string))
					}
					response := `{"action": "subscribe", "status": "ok"}`
					toServerChan <- response*/
				case "cancel":
					pathList := serverMessageMap["path"].([]interface{})
					for i := 0; i < len(pathList); i++ {
						removeFromNotificationList(pathList[i].(string))
					}
					fromServerChan <- serverMessageMap
				default:
					utils.Error.Printf("serverReader:Message action unknown = %s", serverMessageMap["action"].(string))
			}
		}
	}
}

func addToNotificationList(path string) {
	if onNotificationList(path) == -1 {
		notificationList = append(notificationList, path)
	}
}

func removeFromNotificationList(path string) {
	index := onNotificationList(path)
	if index != -1 {
		notificationList = slices.Delete(notificationList, index, index+1)
	}
}

func serverWriter(conn net.Conn, toServerChan chan string) {
	defer conn.Close()
	for {
		select {
		case message := <-toServerChan:
		utils.Info.Printf("serverWriter:Message to server: %s", message)
			_, err := conn.Write([]byte(message))
			if err != nil {
				utils.Error.Printf("serverWriter:Write failed, err = %s", err)
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

func serviceInterfaceMgr(fromServiceChan chan map[string]interface{}, toServiceChan chan map[string]interface{}) {
	serviceOutputChan := make(chan string, MAXSERVICES) // {"path": "x.y.z", "output": {xxx}}
	serviceInputChan := make([]chan string, MAXSERVICES) // {"path": "x.y.z", "input": {xxx}}
	for i := 0; i < MAXSERVICES; i++ {
		serviceInputChan[i] = make(chan string)
	}
	for {
		select {
		case input := <-toServiceChan:
			serviceName := input["path"].(string)
			message, _ := json.Marshal(input)
			utils.Info.Printf("Data for calling a service: %s", message)
			channelIndex := getchannelIndex(serviceName)
			if channelIndex != -1 {
				serviceChannelList[channelIndex].Channel <- string(message)
			} else {
				utils.Error.Printf("serviceInterfaceMgr:Service %s not found", serviceName)
			}
		case output := <- serviceOutputChan:
			utils.Info.Printf("Data from a called  service: %s", output)
			var outputMap map[string]interface{}
			json.Unmarshal([]byte(output), &outputMap)
			fromServiceChan <- outputMap
		}
	}
}

func decodeServiceRegRequest(request []byte, regIndex string) ServiceRegElem {  // {"action": "reg"/"dereg", "name": "xxxx", "infotype": "service"}
	var serviceRegElem ServiceRegElem
	var reqMap map[string]interface{}
	err := json.Unmarshal(request, &reqMap)
	if err != nil {
		utils.Error.Printf("decodeserviceRegRequest:Service reg request corrupt, err = %s\nmessage=%s", err, request)
	}
	if (reqMap["action"] == nil || reqMap["name"] == nil || reqMap["infotype"] == nil) ||
	   (reqMap["action"] != "reg" && reqMap["action"] != "dereg") || (reqMap["infotype"] != "Data" && reqMap["infotype"] != "Service") {
		serviceRegElem.InfoType = "error"
	} else {
		if reqMap["action"] == "dereg" {
			serviceRegElem.Name = reqMap["name"].(string)
			serviceRegElem.InfoType = "dereg"
		} else {
			serviceRegElem.Name = reqMap["name"].(string)
			serviceRegElem.InfoType = reqMap["infotype"].(string)
			serviceRegElem.SockFile = SERVICE_REG_DIR + "serviceReg" + regIndex + ".sock"
		}
	}
	return serviceRegElem
}

func serviceRegServer(serviceRegChan chan ServiceRegElem) {
	os.Remove(SERVICE_REG_SOCKFILE)
	listener, err := net.Listen("unix", SERVICE_REG_SOCKFILE)
	if err != nil {
		utils.Error.Printf("serviceRegServer:UDS listen failed, err = %s", err)
		os.Exit(-1)
	}
	regIndex := 1
	var serviceRegElem ServiceRegElem
	var serviceNameList []string
	buf := make([]byte, 512)
	serviceRegElem.ChannelIndex = -1
	for {
		conn, err := listener.Accept()
		if err != nil {
			utils.Error.Printf("serviceRegServer:UDS accept failed, err = %s", err)
			serviceRegElem.InfoType = "error"
		} else {
			n, err := conn.Read(buf)
			if err != nil {
				utils.Error.Printf("serviceRegServer:Read failed, err = %s", err)
				time.Sleep(1 * time.Second)
				serviceRegElem.InfoType = "error"
			} else {
				utils.Info.Printf("serviceRegServer:Request from service: %s", string(buf[:n]))
				if n > 512 {
					utils.Error.Printf("serviceRegServer:Max message size of 512 chars exceeded. Request dropped")
					serviceRegElem.InfoType = "error"
				} else {
					serviceRegElem = decodeServiceRegRequest(buf[:n], strconv.Itoa(regIndex))
					regIndex++
				}
//utils.Info.Printf("serviceRegServer:serviceRegElem.InfoType=%s, serviceRegElem.Name=%s", serviceRegElem.InfoType, serviceRegElem.Name)
				var response string
				if serviceRegElem.InfoType != "error" && (serviceRegElem.InfoType == "dereg" || !serviceNameClash(serviceNameList, serviceRegElem.Name)) {
					if serviceRegElem.InfoType == "dereg" {
						response = `{"action": "dereg"` + `, "name": "` + serviceRegElem.Name + `"}`
					} else {
						response = `{"action": "reg"` + `, "name": "` + serviceRegElem.Name + `", "sockfile": "` + serviceRegElem.SockFile + `"}`
					}
				} else {
					response = `{"action": "error"}`
					serviceRegElem.InfoType = "error"
				}
				_, err := conn.Write([]byte(response))
				if err != nil {
					utils.Error.Printf("serviceRegServer:Write failed, err = %s", err)
					serviceRegElem.InfoType = "error"
				}
			}
		}
		time.Sleep(3 * time.Second)  //wait some time for the service to be ready for a connect request
		serviceRegChan <- serviceRegElem
		serviceRegElem = <- serviceRegChan  //updated list of service names on Name element
		err = json.Unmarshal([]byte(serviceRegElem.Name), &serviceNameList)
		if err != nil {
			utils.Error.Printf("serviceRegServer:Unmarshal failed, err = %s", err)
		}
		conn.Close()
	}
}

func serviceNameClash(serviceNameList []string, serviceName string) bool {
	for i := 0; i < len(serviceNameList); i++ {
		if serviceNameList[i] == serviceName {
			return true
		}
	}
	return false
}

func updateServiceList(serviceRegElem ServiceRegElem) { //TODO: can this clash with reading on other threads?
	if serviceRegElem.InfoType != "dereg" {
		serviceList = append(serviceList, serviceRegElem)
	} else {
		for i := 0; i < len(serviceList); i++ {
			if serviceList[i].Name == serviceRegElem.Name {
				serviceList[i].Conn = nil
				// serviceReader terminates when the service closes the UDS connection?
				freeServiceChannel(serviceList[i].ChannelIndex)
				serviceList = append(serviceList[:i], serviceList[i+1:]...)
			}
		}
	}
}

func getchannelIndex(serviceName string) int {
	for i := 0; i < len(serviceList); i++ {
		if serviceList[i].Name == serviceName {
			return i
		}
	}
	return -1
}

func createServiceNameList() ServiceRegElem {
	serviceNameList := "["
	for i := 0; i < len(serviceList); i++ {
		serviceNameList += `"` + serviceList[i].Name + `", `
	}
	if len(serviceList) == 0 {
		serviceNameList += "]"
	} else {
		serviceNameList = serviceNameList[:len(serviceNameList)-2] + "]"
	}
	var serviceRegElem ServiceRegElem
	serviceRegElem.Name = serviceNameList
	return serviceRegElem
}

func getServiceChannelIndex(channelIndex int) int {
	if channelIndex != -1 {
		return channelIndex
	}
	for i := 0; i < len(serviceChannelList); i++ {
		if !serviceChannelList[i].Busy {
			serviceChannelList[i].Busy = true
			return i
		}
	}
	return -1
}

func freeServiceChannel(channelIndex int) {
	serviceChannelList[channelIndex].Busy = false
}

func connectToService(serviceReq *ServiceRegElem, fromServiceChan chan map[string]interface{}) {
	utils.Info.Printf("connectToService:Trying to connect to service...")
	serviceReq.Conn, _ = net.Dial("unix", serviceReq.SockFile)
	if serviceReq.Conn == nil {
		utils.Error.Printf("connectToService:Failed to UDS connect to the service %s", serviceReq.Name)
		return
	}
	serviceReq.ChannelIndex = getServiceChannelIndex(serviceReq.ChannelIndex)
	if serviceReq.ChannelIndex == -1 {
		serviceReq.Conn.Close()
		serviceReq.Conn = nil
		utils.Error.Printf("connectToService:No available channel for service %s", serviceReq.Name)
		return
	}
	go serviceReader(serviceReq.Conn, fromServiceChan)
	utils.Info.Printf("connectToService:Connected to the service %s", serviceReq.Name)
	return
}

func serviceConnectRetry(fromServiceChan chan map[string]interface{}) { //TODO:should probably be used..
	for i := 0; i < len(serviceList); i++ {
		if serviceList[i].ChannelIndex != -1 && serviceChannelList[serviceList[i].ChannelIndex].Busy && 
		   serviceChannelList[serviceList[i].ChannelIndex].Channel == nil {
			connectToService(&serviceList[i], fromServiceChan)
			break
		}
	}
}

func serviceReader(udsConn net.Conn, fromServiceChan chan map[string]interface{}) {
	buf := make([]byte, 512)
	for {
		nr, err := udsConn.Read(buf)
		if err != nil {
			utils.Error.Printf("serviceReader:Read from failed, err = %s", err)
			break
		} else {
			var messageMap map[string]interface{}
			err := json.Unmarshal(buf[:nr], &messageMap)
			if err != nil {
				utils.Error.Printf("serviceReader:Unmarshal failed, err = %s", err)
			} else {
				fromServiceChan <- messageMap
			}
		}
	}
}

func main() {
	// Create new parser object
	parser := argparse.NewParser("print", "Service service")
	configFile := parser.String("c", "configfile", &argparse.Options{
		Required: false,
		Help:     "Service service configuration filename",
		Default:  "serviceConfig.json"})
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

	utils.InitLog("service-log.txt", "./logs", *logFile, *logLevel)

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

	fromServerChan := make(chan map[string]interface{}, 3)
	toStorageChan := make(chan map[string]interface{}, 3)
	fromServiceChan := make(chan map[string]interface{}, 3)
	toServiceChan := make(chan map[string]interface{}, 3)

	serviceConfig := readServiceConfig(*configFile)

	go initVSSInterfaceMgr(fromServerChan, toStorageChan, serviceConfig)
	serviceRegChan := make(chan ServiceRegElem)
	go serviceRegServer(serviceRegChan)
	go serviceInterfaceMgr(fromServiceChan, toServiceChan)

	for {
		select {
		case serviceInput := <-fromServerChan:
			toServiceChan <- serviceInput
		case serviceOutput := <-fromServiceChan:
			toStorageChan <- serviceOutput
		case sig := <- sigChan:
			if sig == syscall.SIGUSR1 {
				serviceRegistration("dereg", serviceConfig)
				time.Sleep(1 * time.Second)
				os.Exit(1)
			} else {
				utils.Info.Printf("Received unknown signal=%d",sig)
			}
		case serviceReq := <- serviceRegChan:
			if serviceReq.InfoType != "error" && serviceReq.InfoType != "dereg" {
				connectToService(&serviceReq, fromServiceChan)
			}
			if serviceReq.InfoType != "error" {
				updateServiceList(serviceReq)
			}
			serviceRegChan <- createServiceNameList()
		}
	}
}
