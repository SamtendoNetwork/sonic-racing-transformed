package nex

import (
	"math"
	"sync"
	"time"
)

const (
	alpha float64 = 1.0 / 8.0
	beta  float64 = 1.0 / 4.0
	k     float64 = 4.0
)



type RTT struct {
	sync.Mutex
	lastRTT     float64
	average     float64
	variance    float64
	initialized bool
}


func (rtt *RTT) Adjust(next time.Duration) {

	rtt.Lock()
	if rtt.initialized {
		rtt.variance = (1.0-beta)*rtt.variance + beta*math.Abs(rtt.variance-float64(next))
		rtt.average = (1.0-alpha)*rtt.average + alpha*float64(next)
	} else {
		rtt.lastRTT = float64(next)
		rtt.variance = float64(next) / 2
		rtt.average = float64(next) + k*rtt.variance
		rtt.initialized = true
	}
	rtt.Unlock()
}



func (rtt *RTT) GetRTTSmoothedAvg() float64 {
	return rtt.average / 16
}



func (rtt *RTT) GetRTTSmoothedDev() float64 {
	return rtt.variance / 8
}


func (rtt *RTT) Initialized() bool {
	return rtt.initialized
}


func (rtt *RTT) Average() time.Duration {
	return time.Duration(rtt.average)
}


func NewRTT() *RTT {
	return &RTT{}
}
