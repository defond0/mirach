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
func (c *MirachCron) AddFuncDelayed(spec string, cmd func(), delay time.Duration, res chan<- interface{}) {
	go func() {
		time.Sleep(delay)
		if err := c.AddFunc(spec, cmd); err != nil {
			res <- err
		}
		res <- nil
	}()
}

// AddFuncRandDelay delays the call to AddFunc by a random delay.
// You must also pass in a chan for the purpose of returning either a
// result or an err that the caller can chose to utilize. You must
// provide at least one time.Duration parameter and optionally a second.
// The first duration parameter will be the maximum delay. The second,
// if provided will be the minimum time to wait. Remaining arguments
// are ignored.
func (c *MirachCron) AddFuncRandDelay(spec string, cmd func(), res chan<- interface{}, t ...time.Duration) {
	var max, min int64
	for i, x := range t {
		switch i {
		case 0:
			max = int64(x)
		case 1:
			min = int64(x)
			break
		}
	}
	rand.Seed(time.Now().UTC().UnixNano())
	delay := time.Duration(rand.Int63n(max-min) + min)
	c.AddFuncDelayed(spec, cmd, delay, res)
}
