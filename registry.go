package pb

import (
	"fmt"
	"sync"
)

var elementsM sync.Mutex

var elements = map[string]Element{
	"percent":  ElementPercent,
	"counters": ElementCounters,
	"bar":      adaptiveWrap(ElementBar),
	"speed":    ElementSpeed,
	"rtime":    ElementRemainingTime,
	"etime":    ElementElapsedTime,
	"string":   ElementString,
	"cycle":    ElementCycle,
}

// RegisterElement give you a chance to use custom elements
func RegisterElement(name string, el Element, adaptive bool) {
	if adaptive {
		el = adaptiveWrap(el)
	}
	elementsM.Lock()
	elements[name] = el
	elementsM.Unlock()
}

func getElement(name string, additional map[string]Element) (el Element, err error) {
	if additional != nil && additional[name] != nil {
		return additional[name], nil
	}
	elementsM.Lock()
	el = elements[name]
	elementsM.Unlock()
	if el == nil {
		err = fmt.Errorf("Unexpected element '%s'", name)
	}
	return
}
