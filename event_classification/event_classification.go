package event_classification

import (
	"context"
	"sync"

	"github.com/anngdinh/operator-helper/contexts"
	"github.com/anngdinh/operator-helper/multilock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EventType string

const (
	CreateEvent EventType = "CREATE"
	DeleteEvent EventType = "DELETE"
	SyncEvent   EventType = "SYNC"
)

// Event holds the context of an event
type Event struct {
	Type   EventType
	Obj    client.Object
	OldObj client.Object
}

type EventClassification struct {
	multiLock *multilock.MultiLock // lock by key

	mu    sync.RWMutex // use for the map below
	cache map[string]client.Object

	getResourceByKey func(key string) (client.Object, bool)
	isValid          func(obj client.Object) bool
}

func NewEventClassification(getResourceByKey func(key string) (client.Object, bool), isValid func(obj client.Object) bool) *EventClassification {
	return &EventClassification{
		cache:            make(map[string]client.Object),
		getResourceByKey: getResourceByKey,
		isValid:          isValid,
		multiLock:        multilock.NewMultipleLock(),
	}
}

func (ec *EventClassification) Classify(ctx context.Context, key string) *Event {
	logger := contexts.NewContext(ctx).Log()

	if key == "" {
		logger.Errorf("event_classification: received empty key")
		return nil
	}

	ec.multiLock.Lock(key)
	defer ec.multiLock.Unlock(key)

	objGet, okGet := ec.getResourceByKey(key)

	objCache, okCache := ec.readCache(key)

	objGetValid, objCacheValid := ec.isValid(objGet), ec.isValid(objCache)

	// logger.Infof("objGet: %v, objCache: %v", objGet, objCache)
	logger.Debug("Event Classification")
	logger.Debugf("   okGet:   %v, objGetValid:   %v", okGet, objGetValid)
	logger.Debugf("   okCache: %v, objCacheValid: %v", okCache, objCacheValid)

	// if objGet is deleted, but objCache is exist, then delete
	if okCache && !okGet {
		ec.deleteCache(key)
		return &Event{
			Type: DeleteEvent,
			Obj:  objCache,
		}
	}

	// if objGet, objCache is exist, but objGet have deletionTimestamp, then delete
	if okGet && isHaveDeleteTimestamp(objGet) {
		logger.Debug("Object have deletionTimestamp, delete event.")
		ec.deleteCache(key)
		return &Event{
			Type: DeleteEvent,
			Obj:  objGet,
		}
	}

	if !okCache && okGet {
		if !objGetValid {
			return nil
		}
		ec.writeCache(key, objGet)
		return &Event{
			Type: CreateEvent,
			Obj:  objGet,
		}
	}

	if !okCache && !okGet {
		return nil
	}

	if okCache && okGet {
		if !objGetValid && !objCacheValid {
			return nil
		}

		if !objGetValid {
			ec.deleteCache(key)
			return &Event{
				Type: DeleteEvent,
				Obj:  objCache,
			}
		}

		if !objCacheValid {
			ec.writeCache(key, objGet)
			return &Event{
				Type: CreateEvent,
				Obj:  objGet,
			}
		}
	}

	// okCache && okGet && objGetValid && objCacheValid
	ec.writeCache(key, objGet)
	return &Event{
		Type:   SyncEvent,
		Obj:    objGet,
		OldObj: objCache,
	}
}

func (ec *EventClassification) writeCache(key string, value client.Object) {
	if key == "" {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.cache[key] = value
}

func (ec *EventClassification) readCache(key string) (client.Object, bool) {
	if key == "" {
		return nil, false
	}
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	objCache, okCache := ec.cache[key]
	return objCache, okCache
}

func (ec *EventClassification) deleteCache(key string) {
	if key == "" {
		return
	}
	ec.mu.Lock()
	defer ec.mu.Unlock()
	delete(ec.cache, key)
}

func isHaveDeleteTimestamp(obj client.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}
