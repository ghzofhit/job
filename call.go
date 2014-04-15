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
