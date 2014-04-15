package main

import (
	"encoding/json"
	"net/http"
)

type Result struct {
	Ret    int
	Reason string
	Data   interface{}
}

var (
	MainCron *Cron
	databk   *Logbk
)

func main() {
	MainCron := New()
	MainCron.Start()

	//*
	http.HandleFunc("/add/cron/", cronHandler)
	http.HandleFunc("/add/now/", nowHandler)
	http.HandleFunc("/add/once/", onceHandler)
	http.ListenAndServe(":8888", nil)
	// */
}

//*
func cronHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	err := r.ParseForm()
	if err != nil {
		OutputJson(w, 0, "参数错误", nil)
		return
	}

	//	time := r.FormValue("time")
	var id int64
	id = MainCron.AddFunc("@hourly", func() {})
	OutputJson(w, 1, string(id), nil)

	return
}

func nowHandler(w http.ResponseWriter, r *http.Request) {

}

func onceHandler(w http.ResponseWriter, r *http.Request) {

}

func ajaxHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	err := r.ParseForm()
	if err != nil {
		OutputJson(w, 0, "参数错误", nil)
		return
	}

	time := r.FormValue("time")

	OutputJson(w, 1, time, nil)

	return
}

func OutputJson(w http.ResponseWriter, ret int, reason string, i interface{}) {
	out := &Result{ret, reason, i}
	b, err := json.Marshal(out)
	if err != nil {
		return
	}
	w.Write(b)
}

//*/
