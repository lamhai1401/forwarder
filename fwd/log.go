package fwd

import (
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/lamhai1401/gologs/logs"
)

type FwdLog struct {
}

func NewFwdLog() *FwdLog {
	return &FwdLog{}
}

func (l *FwdLog) Error(msg string, err error, fields watermill.LogFields) {
	logs.Error(fmt.Printf("[%s], msg: %s, fields: %v", err.Error(), msg, fields))
}
func (l *FwdLog) Info(msg string, fields watermill.LogFields) {
	logs.Info(fmt.Sprintf("msg: %s, fields: %v", msg, fields))
}
func (l *FwdLog) Debug(msg string, fields watermill.LogFields) {
	logs.Debug(fmt.Sprintf("msg: %s, fields: %v", msg, fields))
}
func (l *FwdLog) Trace(msg string, fields watermill.LogFields) {
	// logs.Debug(fmt.Sprintf("msg: %s, fields: %v", msg, fields))
}
func (l *FwdLog) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return l
}
