package job

import (
	"errors"
	"os"
)

type Logbk struct {
	logfile *os.File
	running bool
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

func (l *Logbk) Write(line string) error {
	if l.running {
		l.logfile.WriteString(line)
		return nil
	}
	return errors.New("There is no log file link")
}
