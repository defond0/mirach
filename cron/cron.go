// Package cron aliases robfig/cron to add methods.
package cron

import (
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

func (c *MirachCron) AddFuncDelayed(spec string, cmd func(), delay time.Duration, res chan<- interface{}) {
	go func() {
		time.Sleep(delay)
		if err := c.AddFunc(spec, cmd); err != nil {
			res <- err
		}
		res <- nil
	}()
}
