package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	FrontEnd = "../FrontEnd/"
	MainPage = FrontEnd + "main.html"
	Plugin   = "../../Plugin/"
)

func showPage(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(MainPage)
	if err != nil {
		panic(err)
	}
	t.Execute(w, nil)
}

// by date
func getDayResult(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	result, ok := resultMap.Load(date)
	log.Println("getDayResult,date:", date)
	if ok {
		log.Println(result)
		fmt.Fprintf(w, result.(string))
	} else {
		fmt.Fprintf(w, generateDayResult(date))
	}
}

func getDayAIResult(w http.ResponseWriter, r *http.Request) {
	date := r.FormValue("date")
	result, ok := recordAIMap[date]
	log.Println("getDayAIResult,date:", date)
	if ok {
		fmt.Fprintf(w, result)
	}
}

// all records by now = record in the latest day
func getAllResult(w http.ResponseWriter, r *http.Request) {
	date := time.Now().Format(Format)
	for i := 0; i >= -100; i-- {
		latestDate := getOffsetDate(date, i)
		record, ok := recordMap[latestDate]
		if ok {
			// todo sort
			log.Println("getAllResult", latestDate)
			fmt.Fprintf(w, getRecordDiff(record, ""))
			return
		}
	}
}

func getRecordDate(w http.ResponseWriter, r *http.Request) {
	var dates []string
	for date, _ := range recordMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	bytes, _ := json.Marshal(dates)
	log.Println("getRecordDate", dates)
	fmt.Fprintf(w, string(bytes))
}

//func getRTC() time.Time {
//	rtc, err := rtc.GetTime("")
//	if err != nil {
//		log.Println(err)
//	}
//	return rtc.Local()
//}

func uploadResult(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "failed to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "failed to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := getConfigStr("ScanAI", "UploadDir")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "failed to create upload directory", http.StatusInternalServerError)
		return
	}

	filePath := filepath.Join(uploadDir, handler.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "failed to create file on server", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "upload file successfully: %s\n", handler.Filename)
	log.Printf("receive file successfully, file:%s", handler.Filename)
}

func main() {

	http.HandleFunc("/", showPage)
	http.HandleFunc("/getDay", getDayResult)
	http.HandleFunc("/getAll", getAllResult)
	http.HandleFunc("/getRecord", getRecordDate)
	http.HandleFunc("/getAIDay", getDayAIResult)
	http.HandleFunc("/upload", uploadResult)
	http.Handle("/Plugin/", http.StripPrefix("/Plugin/", http.FileServer(http.Dir(Plugin))))
	http.Handle("/FrontEnd/", http.StripPrefix("/FrontEnd/", http.FileServer(http.Dir(FrontEnd))))
	//generateDayResult(getFormatRTC())
	//getRTC()
	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			//checkScan()
		}
	}()

	err := http.ListenAndServe(":"+getConfigStr("Port", "Port"), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}

}
