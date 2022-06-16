package fwd

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

// ClientDataTime linter
type ClientDataTime struct {
	id string
	t  int64
}

const (
	maxChanSize = 1024 * 2
	add         = "add"
	closing     = "closing"
	hub         = "hub"
)

// Forwarder linter
type Forwarder struct {
	id            string
	clients       map[string]*FwdClient
	pubsub        *gochannel.GoChannel
	dataTimeChann chan *ClientDataTime // to check data time foreach msg
	dataTime      map[string]int64
	ctx           context.Context
	cancelFunc    context.CancelFunc
	mutex         sync.Mutex
}

// NewForwarder linter
func NewForwarder() *Forwarder {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: maxChanSize,
		},
		// watermill.NewStdLogger(false, false),
		NewFwdLog(),
	)

	ctx, cancel := context.WithCancel(context.Background())

	f := &Forwarder{
		dataTimeChann: make(chan *ClientDataTime, maxChanSize),
		dataTime:      make(map[string]int64),
		id:            "Forwarder",
		clients:       make(map[string]*FwdClient),
		pubsub:        pubSub,
		ctx:           ctx,
		cancelFunc:    cancel,
	}

	go f.serve()
	return f
}

func (f *Forwarder) serve() {
	for {
		select {
		case data := <-f.dataTimeChann:
			f.mutex.Lock()
			f.dataTime[data.id] = data.t
			f.mutex.Unlock()
		case <-f.ctx.Done():
			return
		}
	}
}

func (f *Forwarder) Close() {
	f.cancelFunc()
}

// Push pushing data to fwd
func (f *Forwarder) Push(trackID string, data []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), data)
	f.dataTimeChann <- &ClientDataTime{
		id: trackID,
		t:  time.Now().UnixMilli(),
	}
	return f.pubsub.Publish(trackID, msg)
}

// Register register to fwd
func (f *Forwarder) Register(pcID string, trackID string) (<-chan *message.Message, error) {
	client := f.GetClient(pcID)
	if client == nil {
		client = NewFwdClient(pcID)
	}

	// check duplicate
	if sub := client.GetSub(trackID); sub != nil {
		client.UnRegister(trackID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	chann, err := f.pubsub.Subscribe(ctx, trackID)
	if err != nil {
		cancel()
		return nil, err
	}

	client.SetMsg(trackID, chann)
	client.SetSub(trackID, cancel)
	return chann, nil
}

// UnRegister linter
func (f *Forwarder) UnRegister(pcID string, trackID string) bool {
	client := f.GetClient(pcID)
	if client == nil {
		return false
	}
	client.UnRegister(trackID)
	return true
}

func (f *Forwarder) CloseClient(pcID string) bool {
	client := f.GetClient(pcID)
	if client == nil {
		return false
	}

	f.mutex.Lock()
	delete(f.clients, pcID)
	f.mutex.Unlock()

	client.Close()
	return true
}

func (f *Forwarder) GetClient(pcID string) *FwdClient {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.clients[pcID]
}

// GetLastTimeReceive linter
func (f *Forwarder) GetLastTimeReceive() map[string]int64 {
	temp := make(map[string]int64)

	f.mutex.Lock()
	for k, v := range f.dataTime {
		temp[k] = v
	}
	f.mutex.Unlock()

	return temp
}

// GetLastTimeReceiveBy linter
func (f *Forwarder) GetLastTimeReceiveBy(trackID string) int64 {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.dataTime[trackID]
}
