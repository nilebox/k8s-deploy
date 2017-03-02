package release

import "log"

// ReleaseEventHandler can handle notifications for events to a Release resource
type ReleaseEventHandler struct {
}

func (h *ReleaseEventHandler) OnAdd(obj interface{}) {
	log.Printf("[REH] OnAdd")
}

func (h *ReleaseEventHandler) OnUpdate(oldObj, newObj interface{}) {
	log.Printf("[REH] OnUpdate")
}

func (h *ReleaseEventHandler) OnDelete(obj interface{}) {
	log.Printf("[REH] OnDelete")
}
