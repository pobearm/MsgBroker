package broker

import (
	"sync"
	"time"
)

type TimeWheel struct {
	sync.Mutex
	interval time.Duration
	ticker   *time.Ticker
	quit     chan struct{}
	timePos  int //时间轮转位置
	addPos   int //顺序写入的位置
	data     []*clients
	buckets  int //轮的数量
}

func NewTimeWheel(interval time.Duration, buckets int) *TimeWheel {
	tw := &TimeWheel{
		interval: interval,
		timePos:  0,
		addPos:   0,
		ticker:   time.NewTicker(interval),
		quit:     make(chan struct{}),
		data:     make([]*clients, buckets),
		buckets:  buckets,
	}
	for i := range tw.data {
		tw.data[i] = NewClients()
	}
	//启动wheel
	go tw.run()
	return tw
}

func (t *TimeWheel) run() {
	for {
		select {
		case <-t.ticker.C:
			t.onTicker()
		case <-t.quit:
			// log.Debug("wheel.run", "time wheel get quit channel signal")
			t.ticker.Stop()
			for _, cs := range t.data {
				cs.StopAll()
			}
			return
		}
	}
}

func (t *TimeWheel) Stop() {
	close(t.quit)
}

func (t *TimeWheel) onTicker() {
	t.Lock()
	t.timePos++
	if t.timePos >= t.buckets {
		t.timePos = 0
	}
	t.Unlock()
	//一定时间触发时, 清理一个clients集合.
	t.data[t.timePos].ClearTimeOut()
}

func (t *TimeWheel) Find(id string) (*Client, bool) {
	for _, cs := range t.data {
		if ok := cs.Exist(id); ok {
			return cs.list[id], true
		}
	}
	return nil, false
}

func (t *TimeWheel) RemoveClient(id string) {
	for _, cs := range t.data {
		if ok := cs.Exist(id); ok {
			cs.Remove(id)
			break
		}
	}
}

func (t *TimeWheel) AddClient(id string, client *Client) {
	//由于timePos的buncket是排它锁, 不能在锁住的时候, 进行添加.
	if t.addPos == t.timePos {
		t.addPos++
	}
	if t.addPos >= t.buckets {
		t.addPos = 0
	}
	cs := t.data[t.addPos]
	cs.Add(id, client)
	t.addPos++
}
