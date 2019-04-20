package routines

import (
	"context"
	"log"
)

const (
	STATUS_IDLE = iota
	STATUS_BUSY
)

const (
	CMD_EXIT = 1
)

type ctrlCmd struct {
	cmd  int
	args interface{}
}
type DoFunc struct {
	F    func(args interface{}) error
	Args interface{}
}

type Worker struct {
	funCh  chan DoFunc
	ctrlCh chan ctrlCmd
	freeCh chan int
	ctx    context.Context
	status int
	id     int
	logger *log.Logger
}

func newWorker(ctx context.Context, id int, l *log.Logger, freeCh chan int) *Worker {
	w := Worker{
		ctx:    ctx,
		status: STATUS_IDLE,
		id:     id,
		logger: l,
		freeCh: freeCh,
	}
	w.funCh = make(chan DoFunc, 1)
	w.ctrlCh = make(chan ctrlCmd, 1)
	w.freeCh <- id
	go w.run()
	return &w
}

func (w *Worker) run() {
	var err error
	for {
		select {
		case doFunc := <-w.funCh:
			w.status = STATUS_BUSY
			if doFunc.F != nil {
				// w.logger.Printf("recv dofunc: %v\n", doFunc)
				if err = doFunc.F(doFunc.Args); err != nil {
					w.logger.Printf("worker %d dofunc error: %v\n", w.id, err)
				}
			}
			w.status = STATUS_IDLE
			w.freeCh <- w.id

		case <-w.ctx.Done():
			w.logger.Printf("worker recv ctx.Done: %v\n", w.ctx.Err())
			return

		case c := <-w.ctrlCh:
			w.logger.Printf("recv cmd: %v\n", c)
			switch c.cmd {
			case CMD_EXIT:
				return
			}
		}
	}
}
