package routines

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type Master struct {
	workers  map[int]*Worker
	freeIdCh chan int
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *log.Logger
	rwmutex  sync.RWMutex
}

type Config struct {
	Num    int
	Logger *log.Logger
}

func New(cfg *Config) *Master {
	mstr := Master{}
	mstr.logger = cfg.Logger
	mstr.ctx, mstr.cancel = context.WithCancel(context.Background())
	mstr.freeIdCh = make(chan int, cfg.Num)
	mstr.workers = make(map[int]*Worker, cfg.Num)
	mstr.rwmutex.Lock()
	for i := 0; i < cfg.Num; i++ {
		mstr.workers[i] = newWorker(mstr.ctx, i, cfg.Logger, mstr.freeIdCh)
	}
	mstr.rwmutex.Unlock()
	return &mstr
}

func (ms *Master) Commit(df DoFunc, timeout ...int) error {
	var id int
	if len(timeout) > 0 && timeout[0] > 0 {
		select {
		case id = <-ms.freeIdCh:
		case <-time.After(time.Duration(timeout[0])):
			return errors.New("commit timeout")
		}
	} else {
		id = <-ms.freeIdCh
	}
	// ms.logger.Printf("commit id is: %d\n", id)
	ms.rwmutex.RLock()
	w := ms.workers[id]
	ms.rwmutex.RUnlock()
	w.funCh <- df
	return nil
}

//cancel all
func (ms *Master) Cancel() {
	ms.cancel()
}

func (ms *Master) Stop(id int) {
	//stop all
	ms.rwmutex.RLock()
	defer ms.rwmutex.RUnlock()
	if id < 0 {
		for _, w := range ms.workers {
			w.ctrlCh <- ctrlCmd{cmd: CMD_EXIT}
		}
		return
	}
	if w, ok := ms.workers[id]; ok {
		w.ctrlCh <- ctrlCmd{cmd: CMD_EXIT}
	}
}

func (ms *Master) Stats() {
	ms.rwmutex.RLock()
	defer ms.rwmutex.RUnlock()
	ms.logger.Printf("worker idle: %d, busy: %d\n", len(ms.freeIdCh), len(ms.workers)-len(ms.freeIdCh))
}
