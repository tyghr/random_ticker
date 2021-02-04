package random_ticker_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/tyghr/random_ticker"
)

func TestTicker(t *testing.T) {

	rt := random_ticker.New()

	tqList := []string{"sys", "plugin", "bo", "pos", "reposync"}

	tickerQueues := struct {
		lock   sync.Mutex
		queues map[string]bool
	}{}

	tickerQueues.queues = make(map[string]bool)
	for _, tqName := range tqList {
		tickerQueues.queues[tqName] = false
	}

	for _, q := range tqList {
		go func(qName string) {
			rt.SetQueue(
				qName,
				2*time.Second,
				4*time.Second,
			)

			tickerQueues.lock.Lock()
			tickerQueues.queues[qName] = true
			tickerQueues.lock.Unlock()

			time.Sleep(5 * time.Second)
			err := rt.RemoveQueue(qName)
			assert.NoError(t, err)
		}(q)
	}

END:
	for {
		select {
		case tName := <-rt.Receive():
			tickerQueues.lock.Lock()
			tickerQueues.queues[tName] = false
			tickerQueues.lock.Unlock()
		case <-time.After(6 * time.Second):
			break END
		}
	}

	for _, isSet := range tickerQueues.queues {
		assert.Equal(t, false, isSet)
	}

	rt.Stop()
}
