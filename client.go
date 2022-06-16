package main

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/pion/rtp"
)

var wrapPool = sync.Pool{New: func() interface{} {
	return new(Wrapper)
}}

var pkgPool = sync.Pool{New: func() interface{} {
	return &rtp.Packet{}
}}

// Wrapper linter
type Wrapper struct {
	Duration time.Duration
	Pkg      *rtp.Packet // save rtp packet
	Data     []byte      `json:"rtp"`  // packet to write
	Type     *string     `json:"type"` // type off wrapper data - ok - ping - pong
}

// FwdClient linter
type FwdClient struct {
	pcID     string
	sub      map[string]context.CancelFunc      // save trackID - context for unsubscribe
	messages map[string]<-chan *message.Message // save msg chann foreach trackID

	mutex sync.Mutex
}

func NewFwdClient(pcID string) *FwdClient {
	f := &FwdClient{
		pcID:     pcID,
		sub:      make(map[string]context.CancelFunc),
		messages: make(map[string]<-chan *message.Message),
	}

	return f
}

func (c *FwdClient) Close() {
	c.mutex.Lock()
	for trackID, cancel := range c.sub {
		delete(c.messages, trackID)
		delete(c.sub, trackID)
		cancel()
	}
	c.mutex.Unlock()
}

func (c *FwdClient) SetSub(trackID string, cancel context.CancelFunc) {
	c.mutex.Lock()
	c.sub[trackID] = cancel
	c.mutex.Unlock()
}

func (c *FwdClient) SetMsg(trackID string, chann <-chan *message.Message) {
	c.mutex.Lock()
	c.messages[trackID] = chann
	c.mutex.Unlock()
}

func (c *FwdClient) UnRegister(trackID string) {
	if cancel := c.GetSub(trackID); cancel != nil {
		c.mutex.Lock()
		delete(c.sub, trackID)
		delete(c.messages, trackID)
		cancel()
		c.mutex.Unlock()
	}
}

func (c *FwdClient) GetSub(trackID string) context.CancelFunc {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.sub[trackID]
}
