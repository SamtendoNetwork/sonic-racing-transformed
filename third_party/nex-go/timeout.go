package nex

import (
	"context"
	"time"
)



type Timeout struct {
	timeout time.Duration
	ctx     context.Context
	cancel  context.CancelFunc
}


func (t *Timeout) SetRTO(timeout time.Duration) {
	t.timeout = timeout
}


func (t *Timeout) RTO() time.Duration {
	return t.timeout
}


func NewTimeout() *Timeout {
	return &Timeout{}
}
