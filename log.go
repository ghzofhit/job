package main

import (
	//	"bytes"
	//	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type Logbk struct {
	logfile *os.File
	running bool
}

type Bean struct {
	Id       int64
	Time     time.Time
	Method   string
	Url      string
	Schedule string
}

func Newbk(filename string) (_ *Logbk, err error) {
	logfile, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_SYNC|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &Logbk{
		logfile: logfile,
		running: true,
	}, nil
}

func (l *Logbk) WriteBin(bin Bean) error {

	if l.running {
		b, err := json.Marshal(bin)
		if err != nil {
			fmt.Println("error:", err)
			return err
		}

		l.logfile.Write(b)
		l.logfile.WriteString("\r\n")
		return nil
	}
	return errors.New("There is no log file link")
}

func (l *Logbk) Write(line string) error {
	if l.running {
		l.logfile.WriteString(line)
		l.logfile.WriteString("\r\n")
		return nil
	}
	return errors.New("There is no log file link")
}
