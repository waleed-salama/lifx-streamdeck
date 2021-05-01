package streamdeck

import (
	"gitlab.com/wwsean08/lifx-streamdeck/lifx"
)

type RediscoverJob struct {
	controller  lifx.Controller
	description string
}

func (r RediscoverJob) Execute() {
	if r.controller != nil {
		_ = r.controller.DiscoverDevices()
	}
}

func (r RediscoverJob) Description() string {
	return r.description
}

func (r RediscoverJob) Key() int {
	return 42
}
