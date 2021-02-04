package random_ticker

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type tickerQueue struct {
	name    string
	minWait time.Duration
	maxWait time.Duration
	stop    chan bool
}

type RTicker struct {
	lock     sync.Mutex
	tQueues  []*tickerQueue
	receiver chan string
}

func New() *RTicker {
	return &RTicker{
		receiver: make(chan string),
	}
}

func (rt *RTicker) Receive() <-chan string {
	return rt.receiver
}

func (rt *RTicker) SetQueue(name string, tMin, tMax time.Duration) {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	for i := range rt.tQueues {
		if rt.tQueues[i].name == name {
			rt.tQueues[i].minWait = tMin
			rt.tQueues[i].maxWait = tMax
			return
		}
	}

	tq := &tickerQueue{
		name:    name,
		minWait: tMin,
		maxWait: tMax,
		stop:    make(chan bool),
	}

	rt.tQueues = append(rt.tQueues, tq)

	go func(t *tickerQueue) {
		stop := false
		for !stop {
			select {
			case <-t.stop:
				stop = true
			case <-waitRandDuration(t.minWait, t.maxWait):
				rt.receiver <- t.name
			}
		}
	}(tq)
}

func (rt *RTicker) RemoveQueue(name string) error {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	index := -1
	for i := range rt.tQueues {
		if rt.tQueues[i].name == name {
			index = i
			break
		}
	}
	if index > -1 {
		rt.tQueues[index].stop <- true
		rt.tQueues[index] = nil
		rt.tQueues = append(rt.tQueues[0:index], rt.tQueues[index+1:]...)
		return nil
	}
	return errors.New("ticker queue not found")
}

func (rt *RTicker) Stop() {
	rt.lock.Lock()
	defer rt.lock.Unlock()

	for i := range rt.tQueues {
		rt.tQueues[i].stop <- true
		rt.tQueues[i] = nil
	}
	rt.tQueues = nil
	close(rt.receiver)
}

func (rt *RTicker) QueueCount() int {
	rt.lock.Lock()
	defer rt.lock.Unlock()
	return len(rt.tQueues)
}

func waitRandDuration(tMin, tMax time.Duration) <-chan time.Time {
	minSec := int64(tMin)
	maxSec := int64(tMax)

	secWait := minSec
	if maxSec > minSec {
		secWait = minSec + rand.Int63n(maxSec-minSec)
	}

	return time.After(time.Duration(secWait))
}
