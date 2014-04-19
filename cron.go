// This library implements a cron spec parser and runner.  See the README for
// more details.
package main

import (
	"sort"
	"sync/atomic"
	"time"
)

// Cron keeps track of any number of entries, invoking the associated func as
// specified by the schedule. It may be started, stopped, and the entries may
// be inspected while running.
type Cron struct {
	entries   []*Entry
	stop      chan struct{}
	add       chan *Entry
	del       chan int64
	snapshot  chan []*Entry
	running   bool
	increment int64
}

// Job is an interface for submitted cron jobs.
type Job interface {
	Run(int64)
}

// The Schedule describes a job's duty cycle.
type Schedule interface {
	// Return the next activation time, later than the given time.
	// Next is invoked initially, and then each time the job is run.
	Next(time.Time) time.Time
}

type OnceSchedule struct {
	runed   bool
	thetime time.Time
}

func (s *OnceSchedule) Next(t time.Time) time.Time {
	if t.Before(s.thetime) && !s.runed {
		return s.thetime
	}
	return time.Time{}
}

// Entry consists of a schedule and the func to execute on that schedule.
type Entry struct {
	// The schedule on which this job should be run.
	Schedule Schedule

	// The next time the job will run. This is the zero time if Cron has not been
	// started or this entry's schedule is unsatisfiable
	Next time.Time

	// The last time this job was run. This is the zero time if the job has never
	// been run.
	Prev time.Time

	// The Job to run.
	Job Job

	// The Job id
	Id int64
}

// byTime is a wrapper for sorting the entry array by time
// (with zero time at the end).
type byTime []*Entry

func (s byTime) Len() int      { return len(s) }
func (s byTime) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byTime) Less(i, j int) bool {
	// Two zero times should return false.
	// Otherwise, zero is "greater" than any other time.
	// (To sort it at the end of the list.)
	if s[i].Next.IsZero() {
		return false
	}
	if s[j].Next.IsZero() {
		return true
	}
	return s[i].Next.Before(s[j].Next)
}

// New returns a new Cron job runner.
func New() *Cron {
	return &Cron{
		entries:   nil,
		add:       make(chan *Entry),
		del:       make(chan int64),
		stop:      make(chan struct{}),
		snapshot:  make(chan []*Entry),
		running:   false,
		increment: time.Now().UnixNano(),
	}
}

// A wrapper that turns a func() into a cron.Job
type FuncJob func(id int64)

func (f FuncJob) Run(id int64) { f(id) }

// AddFunc adds a func to the Cron to be run on the given schedule.
func (c *Cron) AddFunc(spec string, cmd func(id int64)) int64 {
	return c.AddJob(spec, FuncJob(cmd))
}

func (c *Cron) DelJob(id int64) {
	if !c.running {
		for i, entry := range c.entries {
			if entry.Id == id {
				c.entries = append(c.entries[:i], c.entries[i+1:]...)
				return
			}
		}
	}
	c.del <- id
}

// AddFunc adds a Job to the Cron to be run on the given schedule.
func (c *Cron) AddJob(spec string, cmd Job) int64 {
	return c.Schedule(Parse(spec), cmd)
}

func (c *Cron) AddOncejob(once time.Time, cmd Job) int64 {
	schedule := &OnceSchedule{
		runed:   false,
		thetime: once,
	}
	return c.Schedule(schedule, cmd)
}

// Schedule adds a Job to the Cron to be run on the given schedule.
func (c *Cron) Schedule(schedule Schedule, cmd Job) int64 {
	var increment int64
	increment = c.getIncrement()
	entry := &Entry{
		Schedule: schedule,
		Job:      cmd,
		Id:       increment,
	}
	if !c.running {
		c.entries = append(c.entries, entry)
		return increment
	}
	c.add <- entry
	return increment
}

// Entries returns a snapshot of the cron entries.
func (c *Cron) Entries() []*Entry {
	if c.running {
		c.snapshot <- nil
		x := <-c.snapshot
		return x
	}
	return c.entrySnapshot()
}

// Start the cron scheduler in its own go-routine.
func (c *Cron) Start() {
	c.running = true
	go c.run()
}

// Run the scheduler.. this is private just due to the need to synchronize
// access to the 'running' state variable.
func (c *Cron) run() {
	// Figure out the next activation times for each entry.
	now := time.Now().Local()

	for _, entry := range c.entries {
		entry.Next = entry.Schedule.Next(now)
	}
	for {
		// Determine the next entry to run.
		sort.Sort(byTime(c.entries))
		var effective time.Time
		if len(c.entries) == 0 || c.entries[0].Next.IsZero() {
			// If there are no entries yet, just sleep - it still handles new entries
			// and stop requests.
			effective = now.AddDate(10, 0, 0)
		} else {
			effective = c.entries[0].Next
		}

		select {
		case now = <-time.After(effective.Sub(now)):
			// Run every entry whose next time was this effective time.
			for i := 0; i < len(c.entries); i++ {
				if c.entries[i].Next != effective {
					break
				}
				go c.entries[i].Job.Run(c.entries[i].Id)
				c.entries[i].Prev = c.entries[i].Next
				c.entries[i].Next = c.entries[i].Schedule.Next(effective)
				if c.entries[i].Next.IsZero() {
					c.entries = append(c.entries[:i], c.entries[i+1:]...)
					i -= 1
				}
			}
			continue

		case newEntry := <-c.add:
			c.entries = append(c.entries, newEntry)
			newEntry.Next = newEntry.Schedule.Next(now)
		case id := <-c.del:
			for i, entry := range c.entries {
				if entry.Id == id {
					c.entries = append(c.entries[:i], c.entries[i+1:]...)
				}
			}
			continue
		case <-c.snapshot:
			c.snapshot <- c.entrySnapshot()

		case <-c.stop:
			return
		}

		// 'now' should be updated after newEntry and snapshot cases.
		now = time.Now().Local()
	}
}

// Stop the cron scheduler.
func (c *Cron) Stop() {
	c.stop <- struct{}{}
	c.running = false
}

func (c *Cron) getIncrement() int64 {
	atomic.AddInt64(&c.increment, 1)
	//c.increment = c.increment + 1
	return c.increment
}

// entrySnapshot returns a copy of the current cron entry list.
func (c *Cron) entrySnapshot() []*Entry {
	entries := []*Entry{}
	for _, e := range c.entries {
		entries = append(entries, &Entry{
			Schedule: e.Schedule,
			Next:     e.Next,
			Prev:     e.Prev,
			Job:      e.Job,
		})
	}
	return entries
}
