package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"time"
)

const (
	FrontEnd = "../../FrontEnd/"
	MainPage = FrontEnd + "main.html"
	Plugin   = "../../../Plugin/"
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

// all records by now = record in the latest day
func getAllResult(w http.ResponseWriter, r *http.Request) {
	date := getFormatRTC()
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

func main() {

	http.HandleFunc("/", showPage)
	http.HandleFunc("/getDay", getDayResult)
	http.HandleFunc("/getAll", getAllResult)
	http.HandleFunc("/getRecord", getRecordDate)
	http.Handle("/Plugin/", http.StripPrefix("/Plugin/", http.FileServer(http.Dir(Plugin))))
	http.Handle("/FrontEnd/", http.StripPrefix("/FrontEnd/", http.FileServer(http.Dir(FrontEnd))))
	//generateDayResult(getFormatRTC())

	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			checkScan()
		}
	}()

	err := http.ListenAndServe(":"+getConfigStr("Port", "Port"), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
