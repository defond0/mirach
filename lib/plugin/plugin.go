package plugin

import (
	"encoding/json"
	"io"

	"github.com/cleardataeng/mirach/lib/cron"
)

type Runner interface {
	Run() []byte
}

// Plugin is a routine used to collect data.
type PlugConf struct {
	Name      string
	LoadDelay string `mapstructure:"load_delay"`
	RunAtLoad bool   `mapstructure:"run_at_load"`
	Schedule  string
	Plugin    Runner
}

func (p PlugConf) cronStat() {
	bytes := p.Plugin.Run()
	res := struct {
		name string
		data []byte
	}{
		name: p.Name,
		data: bytes,
	}
	realres := json.Marshal(res)
	for _, b := range p.Backend {
		b.Write(realres)
	}
}

type Ctrl struct {
	plugins []PlugConf
	cron    *cron.MirachCron
	backend map[PlugConf][]io.ReadWriter
}

func NewCtrl(b []io.ReadWriter) *Ctrl {
	return &Ctrl{
		cron:    cron.New(),
		backend: b,
	}
}

func (c *Ctrl) Start() {
	c.cron.Start()
}

func (c *Ctrl) Stop() {
	c.cron.Stop()
}
