// Package cron aliases robfig/cron to add methods.
package cron

import (
	"math/rand"
	"time"

	"github.com/robfig/cron"
)

// MirachCron has robfig/cron embedded and has the only purpose of adding
// methods to that type.
type MirachCron struct {
	*cron.Cron
}

// New returns a pointer to MirachCron with an initialized Cron.
func New() *MirachCron {
	c := cron.New()
	return &MirachCron{c}
}

// AddFuncDelayed delays the call to AddFunc by the given delay.
// You must also pass in a chan for the purpose of returning either a
// result or an err that the caller can chose to utilize.
func (c *MirachCron) AddFuncDelayed(spec string, cmd func(), delay time.Duration, errChan chan<- error) {
	go func() {
		time.Sleep(delay)
		if err := c.AddFunc(spec, cmd); err != nil {
			errChan <- err
		}
		errChan <- nil
	}()
}

// AddFuncRandDelay delays the call to AddFunc by a random delay.
func (c *MirachCron) AddFuncRandDelay(spec string, cmd func(), max, min time.Duration, errChan chan<- error) {
	rand.Seed(time.Now().UTC().UnixNano())
	delay := time.Duration(rand.Int63n(int64(max)-int64(min)) + int64(min))
	c.AddFuncDelayed(spec, cmd, delay, errChan)
}

// AddJobDelayed delays the call to AddJob by the given delay.
func (c *MirachCron) AddJobDelayed(spec string, cmd cron.Job, delay time.Duration, errChan chan<- error) {
	go func() {
		time.Sleep(delay)
		if err := c.AddJob(spec, cmd); err != nil {
			errChan <- err
		}
		errChan <- nil
	}()
}

// AddJobRandDelay delays the call to AddJob by a random delay.
func (c *MirachCron) AddJobRandDelay(spec string, cmd cron.Job, max, min time.Duration, errChan chan<- error) {
	rand.Seed(time.Now().UTC().UnixNano())
	delay := time.Duration(rand.Int63n(int64(max)-int64(min)) + int64(min))
	c.AddJobDelayed(spec, cmd, delay, errChan)
}
