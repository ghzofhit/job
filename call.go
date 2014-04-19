package main

import (
	"net/http"
)

type Call struct {
	url string
}

func (c *Call) Run() {
	resp, err := http.Get("http://test.sohu.com/lib/r.php")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
}

func CallUrl(url string, id int64) error {
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		return err
	}
	if resp.StatusCode == 200 {
		//
		return err
	}
	defer resp.Body.Close()
	return nil
}
