package plugin

import (
	"io"

	"github.com/cleardataeng/mirach/lib/cron"
)

// Plugin is a routine used to collect data.
type Plugin struct {
	Disabled  bool
	Label     string
	LoadDelay string `mapstructure:"load_delay"`
	RunAtLoad bool   `mapstructure:"run_at_load"`
	Schedule  string
	Type      string
}

type Ctrl struct {
	cron    *cron.MirachCron
	backend []io.ReadWriter
}

func NewCtrl(c *cron.MirachCron, b []io.ReadWriter) *Ctrl {
	return &Ctrl{
		cron:    c,
		backend: b,
	}
}

func (c *Ctrl) Start() {
	c.cron.Start()
}

func (c *Ctrl) Stop() {
	c.cron.Stop()
}
