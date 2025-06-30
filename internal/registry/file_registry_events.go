package registry

import (
	"fmt"
)

// AddEventSubscriber adds an event subscriber
func (r *FileRegistry) AddEventSubscriber(subscriber EventSubscriber) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.events = append(r.events, subscriber)
}

// emitEvent sends an event to all subscribers
func (r *FileRegistry) emitEvent(event *RegistryEvent) {
	for _, subscriber := range r.events {
		// Send events asynchronously to avoid blocking
		go func(sub EventSubscriber, evt *RegistryEvent) {
			if err := sub.OnEvent(evt); err != nil {
				// Log error but don't fail the operation
				fmt.Printf("Event subscriber error: %v\n", err)
			}
		}(subscriber, event)
	}
}
