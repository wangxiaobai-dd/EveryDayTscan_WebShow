package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	mapset "github.com/deckarep/golang-set"
	"gopkg.in/ini.v1"
)

const (
	ConfigIni = "config.ini"
	Format    = "2006-01-02"
)

type FileType int

const (
	SCAN FileType = iota
	AI
	DUMP
)

var cfg *ini.File
var checkToday bool

// scan record on that day
var recordMap map[string]string
var recordAIMap map[string]string

type DumpVersion struct {
	Version int    `json:"version"`
	Date    string `json:"date"`
	File    string `json:"file"`
}

var recordDumpMap map[string][]DumpVersion

// diff with previous day
var resultMap sync.Map

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var err error
	cfg, err = ini.Load(ConfigIni)
	if err != nil {
		panic(err)
	}
	dir := getConfigStr("ScanTool", "OutputDir")
	files, err := os.ReadDir(dir)
	if err != nil {
		//panic(err)
		log.Println(err)
	}

	if isOpen, _ := getConfigBool("LogToFile", "Open"); isOpen {
		file, err := os.OpenFile(getConfigStr("LogToFile", "Path"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err == nil {
			log.SetOutput(file)
		}
	}

	recordMap = make(map[string]string)
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), getConfigStr("ScanTool", "ResultFile")) {
			continue
		}
		log.Println(f.Name())
		readRecord(SCAN, dir+f.Name())
	}
	// AI Result
	recordAIMap = make(map[string]string)
	aiDir := getConfigStr("ScanAI", "OutputDir")
	files, err = os.ReadDir(aiDir)
	if err != nil {
		log.Println(err)
	}
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), getConfigStr("ScanAI", "ResultFile")) {
			continue
		}
		log.Println(f.Name())
		readRecord(AI, aiDir+f.Name())
	}
	// Dump
	recordDumpMap = make(map[string][]DumpVersion)
	dumpDir := getConfigStr("Dump", "UploadDir")
	files, err = os.ReadDir(dumpDir)
	if err != nil {
		log.Println(err)
	}
	for _, f := range files {
		if version, date, ok := parseDumpFilename(f.Name()); ok {
			recordDumpMap[date] = append(recordDumpMap[date], DumpVersion{
				Version: version,
				Date:    date,
				File:    dumpDir + f.Name(),
			})
			log.Println(f.Name())
		}
	}
}

func parseDumpFilename(name string) (version int, date string, ok bool) {
	pattern := regexp.MustCompile(`^(\d+)-dump\.(\d{4}-\d{2}-\d{2})\.html$`)
	matches := pattern.FindStringSubmatch(name)
	if len(matches) != 3 {
		return 0, "", false
	}

	version, _ = strconv.Atoi(matches[1])
	date = matches[2]
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return 0, "", false
	}
	return version, date, true
}

func readRecord(fileType FileType, path string) {
	f, err := os.ReadFile(path)
	if err != nil {
		log.Println(err)
		return
	}
	if fileType == AI {
		recordAIMap[getDateFromPath(path)] = string(f)
	} else if fileType == SCAN {
		recordMap[getDateFromPath(path)] = string(f)
	} else if fileType == DUMP {

	}
	log.Printf("readRecord,path:%s,date:%s,fileType:%v", path, getDateFromPath(path), fileType)
}

func getDateFromPath(path string) string {
	slice := strings.Split(path, ".")
	if len(slice) > 0 {
		return slice[len(slice)-1]
	}
	return ""
}

func checkScan() {
	scanHour, _ := getConfigInt("ScanTime", "Hour")
	scanMin, _ := getConfigInt("ScanTime", "Minute")

	localTime := time.Now()
	if localTime.Hour() == scanHour {
		if !checkToday && localTime.Minute() <= scanMin {
			checkToday = true
			callScan(localTime)
		} else if localTime.Minute() > scanMin {
			checkToday = false
		}
	}
}

func callScan(localTime time.Time) {
	notScanDay := getConfigInts("ScanTool", "NotScanDay", ",")
	for _, day := range notScanDay {
		if day == int(localTime.Weekday()) {
			log.Println("not scan day:", day)
			return
		}
	}

	// 1:svn up
	err := updateCode()
	if err != nil {
		log.Println("updateCode:", err)
	}

	// 2:tscancode
	resultFileName := getFileName(localTime)
	projectPath := getConfigStr("Project", "Path")
	dirs := getConfigStrs("Project", "Dirs", ",")

	execStr := getConfigStr("ScanTool", "Path") + " "
	for _, dir := range dirs {
		execStr += projectPath + dir + " "
	}
	execStr += "2>" + resultFileName + " " + getConfigStr("ScanTool", "Param")

	log.Println(execStr)

	cmd := exec.Command("/bin/bash", "-c", execStr)
	err = cmd.Run()
	if err != nil {
		log.Println(err)
	} else {
		err := os.Chmod(resultFileName, 0666)
		if err != nil {
			log.Println(err)
		}
		readRecord(SCAN, resultFileName)
		log.Println("scan finish")
	}
}

func updateCode() error {
	_, err := user.Lookup(getConfigStr("CodeVcs", "User"))
	if err != nil {
		return err
	}
	cmd := exec.Command("/bin/bash", "-c", getConfigStr("CodeVcs", "Cmd")+" "+getConfigStr("Project", "Path"))
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	log.Println(string(output))
	return nil
}

func getFileName(time time.Time) string {
	return getConfigStr("ScanTool", "OutputDir") + getConfigStr("ScanTool", "ResultFile") + "." + time.Format(Format)
}

func getOffsetDate(date string, day int) string {
	oldTime, err := time.Parse(Format, date)
	if err != nil {
		log.Println(err)
		return ""
	}
	newTime := oldTime.AddDate(0, 0, day)
	return newTime.Format(Format)
}

func generateDayResult(date string) string {
	log.Println("generateDayResult,date:", date)
	record, ok := recordMap[date]
	if !ok {
		log.Println("noresult", date)
		return ""
	}
	var result string
	// It doesn't scan in some days, such as Saturday and Sunday, in config.ini
	for i := -1; i >= -5; i-- {
		offsetDate := getOffsetDate(date, i)
		offsetRecord, ok := recordMap[offsetDate]
		if ok {
			result = getRecordDiff(record, offsetRecord)
			resultMap.Store(date, result)
			log.Println(result, "d:", date, "offset:", offsetDate)
			break
		}
	}
	return result
}

func getRecordDiff(record string, offsetRecord string) string {
	log.Println("getRecordDiff")

	recordSet := mapset.NewSet()
	offsetSet := mapset.NewSet()

	s := strings.Split(record, "\n")
	for _, line := range s {
		recordSet.Add(line)
	}
	s = strings.Split(offsetRecord, "\n")
	for _, line := range s {
		offsetSet.Add(line)
	}
	diff := recordSet.Difference(offsetSet)
	log.Println(diff)
	iter := diff.Iterator()
	var sResult []string
	for elem := range iter.C {
		sResult = append(sResult, elem.(string))
	}
	sort.Strings(sResult)
	bytes, _ := json.Marshal(sResult)
	return string(bytes)
}

func getConfigStr(section string, key string) string {
	return cfg.Section(section).Key(key).String()
}

func getConfigBool(section string, key string) (bool, error) {
	return cfg.Section(section).Key(key).Bool()
}

func getConfigInt(section string, key string) (int, error) {
	return cfg.Section(section).Key(key).Int()
}

func getConfigInts(section string, key string, delim string) []int {
	return cfg.Section(section).Key(key).Ints(delim)
}

func getConfigStrs(section string, key string, delim string) []string {
	return cfg.Section(section).Key(key).Strings(delim)
}
