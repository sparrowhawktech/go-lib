package sql

import (
	"sync"
)

type TxListener func(txCtx *TxCtx, eventName string, data interface{})

type TxDispatcher struct {
	listeners []TxListener
	mux       sync.Mutex
}

func (o *TxDispatcher) RegisterListener(listenr TxListener) {
	o.mux.Lock()
	defer o.mux.Unlock()
	o.listeners = append(o.listeners, listenr)
}

func (o *TxDispatcher) Dispatch(txCtx *TxCtx, eventName string, data interface{}) {
	snapshot := o.takeSnapshot()
	for _, listener := range snapshot {
		listener(txCtx, eventName, data)
	}
}

func (o *TxDispatcher) takeSnapshot() []TxListener {
	o.mux.Lock()
	defer o.mux.Unlock()
	snapshot := make([]TxListener, 0)
	snapshot = append(snapshot, o.listeners...)
	return snapshot
}

func newTxDispatcher() *TxDispatcher {
	return &TxDispatcher{listeners: make([]TxListener, 0), mux: sync.Mutex{}}
}

var GlobalTxDispatcher = newTxDispatcher()
