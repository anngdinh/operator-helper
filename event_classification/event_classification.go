package event_classification

import (
	"context"

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
	cache            map[string]client.Object
	getResourceByKey func(key string) (client.Object, bool)
	isValid          func(obj client.Object) bool
	multiLock        *multilock.MultiLock
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
	ec.multiLock.Lock(key)
	defer ec.multiLock.Unlock(key)

	objGet, okGet := ec.getResourceByKey(key)
	objCache, okCache := ec.cache[key]

	objGetValid, objCacheValid := ec.isValid(objGet), ec.isValid(objCache)

	// logger.Infof("objGet: %v, objCache: %v", objGet, objCache)
	logger.Debug("Event Classification")
	logger.Debugf("   okGet:   %v, objGetValid:   %v", okGet, objGetValid)
	logger.Debugf("   okCache: %v, objCacheValid: %v", okCache, objCacheValid)

	// if objGet is deleted, but objCache is exist, then delete
	if okCache && !okGet {
		delete(ec.cache, key)
		return &Event{
			Type: DeleteEvent,
			Obj:  objCache,
		}
	}

	// if objGet, objCache is exist, but objGet have deletionTimestamp, then delete
	if okGet && isHaveDeleteTimestamp(objGet) {
		logger.Debug("Object have deletionTimestamp, delete event.")
		delete(ec.cache, key)
		return &Event{
			Type: DeleteEvent,
			Obj:  objGet,
		}
	}

	if !okCache && okGet {
		if !objGetValid {
			return nil
		}
		ec.cache[key] = objGet
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
			delete(ec.cache, key)
			return &Event{
				Type: DeleteEvent,
				Obj:  objCache,
			}
		}

		if !objCacheValid {
			ec.cache[key] = objGet
			return &Event{
				Type: CreateEvent,
				Obj:  objGet,
			}
		}
	}

	// okCache && okGet && objGetValid && objCacheValid
	ec.cache[key] = objGet
	return &Event{
		Type:   SyncEvent,
		Obj:    objGet,
		OldObj: objCache,
	}
}

func isHaveDeleteTimestamp(obj client.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}
