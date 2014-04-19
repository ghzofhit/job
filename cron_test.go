package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// Many tests schedule a job for every second, and then wait at most a second
// for it to run.  This amount is just slightly larger than 1 second to
// compensate for a few milliseconds of runtime.
const ONE_SECOND = 1*time.Second + 10*time.Millisecond
const FIVE_SECOND = 5*time.Second + 10*time.Millisecond

// Start and stop cron with no entries.
func TestNoEntries(t *testing.T) {

	cron := New()
	cron.Start()

	Convey("When have no entries.", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-stop(cron):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

// Start, stop, then add an entry. Verify entry doesn't run.
func TestStopCausesJobsToNotRun(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.Start()
	cron.Stop()
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })
	Convey("Start, stop, then add an entry. Verify entry doesn't run.", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):
			// No job ran!
			tag = true
		case <-wait(wg):

		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestAddBeforeRunning(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })
	cron.Start()
	defer cron.Stop()

	// Give cron 2 seconds to run our job (which is always activated).
	Convey("Add a job, start cron, expect it runs.", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestDelBeforeRunning(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	id := cron.AddFunc("*/5 * * * * ?", func(id int64) { wg.Done() })
	cron.AddFunc("0 0 0 1 1 ?", func(id int64) {})
	cron.DelJob(id)
	cron.Start()
	defer cron.Stop()

	// Give cron 2 seconds to run our job (which is always activated).
	Convey("Del a job,del cron,expect it not runs.", t, func() {
		tag := false
		select {
		case <-time.After(FIVE_SECOND):
			tag = true
		case <-wait(wg):

		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestAddWhileRunning(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.Start()
	defer cron.Stop()
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })
	Convey("Start cron, add a job, expect it runs.", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestDelWhileRunning(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	id := cron.AddFunc("*/5 * * * * ?", func(id int64) { wg.Done() })
	cron.AddFunc("0 0 0 1 1 ?", func(id int64) {})
	cron.Start()
	cron.DelJob(id)

	defer cron.Stop()
	Convey("Start cron, add a job ,and del it, expect not runs.", t, func() {
		// Give cron 2 seconds to run our job (which is always activated).
		tag := false
		select {
		case <-time.After(FIVE_SECOND):
			tag = true
		case <-wait(wg):

		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestSnapshotEntries(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.AddFunc("@every 2s", func(id int64) { wg.Done() })
	cron.Start()
	defer cron.Stop()

	// Cron should fire in 2 seconds. After 1 second, call Entries.
	select {
	case <-time.After(ONE_SECOND):
		cron.Entries()
	}
	Convey("Test timing with Entries.", t, func() {
		// Even though Entries was called, the cron should fire at the 2 second mark.
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

// Test that the entries are correctly sorted.
// Add a bunch of long-in-the-future entries, and an immediate entry, and ensure
// that the immediate entry runs immediately.
// Also: Test that multiple jobs run in the same instant.
func TestMultipleEntries(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	cron := New()
	cron.AddFunc("0 0 0 1 1 ?", func(id int64) {})
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })
	cron.AddFunc("0 0 0 31 12 ?", func(id int64) {})
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })

	cron.Start()
	defer cron.Stop()

	Convey("Test that the entries are correctly sorted.", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestRunningJobTwice(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	cron := New()
	cron.AddFunc("0 0 0 1 1 ?", func(id int64) {})
	cron.AddFunc("0 0 0 31 12 ?", func(id int64) {})
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })

	cron.Start()
	defer cron.Stop()
	Convey("Test running the same job twice.", t, func() {
		tag := false
		select {
		case <-time.After(2 * ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})

}

func TestRunningMultipleSchedules(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(2)

	cron := New()
	cron.AddFunc("0 0 0 1 1 ?", func(id int64) {})
	cron.AddFunc("0 0 0 31 12 ?", func(id int64) {})
	cron.AddFunc("* * * * * ?", func(id int64) { wg.Done() })
	cron.Schedule(Every(time.Minute), FuncJob(func(id int64) {}))
	cron.Schedule(Every(time.Second), FuncJob(func(id int64) { wg.Done() }))
	cron.Schedule(Every(time.Hour), FuncJob(func(id int64) {}))

	cron.Start()
	defer cron.Stop()
	Convey("Test running multischedules", t, func() {
		tag := false
		select {
		case <-time.After(2 * ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

//
func TestLocalTimezone(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	now := time.Now().Local()
	spec := fmt.Sprintf("%d %d %d %d %d ?",
		now.Second()+1, now.Minute(), now.Hour(), now.Day(), now.Month())

	cron := New()
	cron.AddFunc(spec, func(id int64) { wg.Done() })
	cron.Start()
	defer cron.Stop()
	Convey("Test that the cron is run in the local time zone (as opposed to UTC).", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

type testJob struct {
	wg   *sync.WaitGroup
	name string
}

func (t testJob) Run(id int64) {
	t.wg.Done()
}

//
func TestJob(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.AddJob("0 0 0 30 Feb ?", testJob{wg, "job0"})
	cron.AddJob("0 0 0 1 1 ?", testJob{wg, "job1"})
	cron.AddJob("* * * * * ?", testJob{wg, "job2"})
	cron.AddJob("1 0 0 1 1 ?", testJob{wg, "job3"})
	cron.Schedule(Every(5*time.Second+5*time.Nanosecond), testJob{wg, "job4"})
	cron.Schedule(Every(5*time.Minute), testJob{wg, "job5"})

	cron.Start()
	defer cron.Stop()
	Convey("Simple test using Runnables.", t, func() {
		tag := false
		select {
		case <-time.After(ONE_SECOND):

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)

		// Ensure the entries are in the right order.
		expecteds := []string{"job2", "job4", "job5", "job1", "job3", "job0"}

		var actuals []string
		for _, entry := range cron.Entries() {
			actuals = append(actuals, entry.Job.(testJob).name)
		}

		for i, expected := range expecteds {
			So(actuals[i], ShouldResemble, expected)
		}
	})
}

func TestOncejob(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	cron := New()
	cron.Start()
	cron.AddOncejob(time.Now().Add(time.Second*2), FuncJob(func(id int64) { wg.Done() }))
	defer cron.Stop()
	Convey("Start, Once job.", t, func() {
		tag := false
		select {
		case <-time.After(FIVE_SECOND):
			// No job ran!

		case <-wait(wg):
			tag = true
		}
		So(tag, ShouldBeTrue)
	})
}

func wait(wg *sync.WaitGroup) chan bool {
	ch := make(chan bool)
	go func() {
		wg.Wait()
		ch <- true
	}()
	return ch
}

func stop(cron *Cron) chan bool {
	ch := make(chan bool)
	go func() {
		cron.Stop()
		ch <- true
	}()
	return ch
}
