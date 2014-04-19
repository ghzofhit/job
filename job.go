package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
	//	"strconv"
	"strings"
)

type Result struct {
	Ret    int
	Reason string
	Data   interface{}
}

var (
	MainCron *Cron
	databk   *Logbk
	logs     *Logbk
)

func main() {
	MainCron = New()
	MainCron.Start()
	var err error
	databk, err = Newbk("data.log")
	if err != nil {
		fmt.Println("数据日志创建错误")
		return
	}
	logs, err = Newbk("info.log")
	if err != nil {
		fmt.Println("日志创建错误")
		return
	}
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
	feed := r.FormValue("url")
	schedule := r.FormValue("schedule")

	if !strings.HasPrefix(feed, "http") {
		feed = "http://" + feed
	}

	//判断url是否合理
	host, err := url.ParseRequestURI(feed)
	if err != nil {

	}

	//判断是否能解析到对应的host记录
	_, err = net.LookupIP(host.Host)
	if err != nil {

	}

	//	time := r.FormValue("time")
	var jid int64
	jid = MainCron.AddFunc("@hourly", func(id int64) { CallUrl(feed, id) })

	databk.WriteBin(Bean{Id: jid,
		Time:     time.Now(),
		Schedule: schedule,
		Method:   "cron",
		Url:      "feed"})

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
