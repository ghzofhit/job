package job

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestLogbk(t *testing.T) {
	var databk *Logbk
	var err error
	Convey("Create logbackup file.", t, func() {
		databk, err = Newbk("test.log")
		So(err, ShouldBeNil)
	})
	Convey("Add a ling data to backup file.", t, func() {
		databk.Write("this is a test!\r\n")
		databk.Write("this is a another test!\r\n")
		So(err, ShouldBeNil)

	})
	defer databk.logfile.Close()
}
