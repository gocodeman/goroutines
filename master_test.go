package routines

import (
	"log"
	"os"
	"testing"
	"time"
)

func fib(num int) int {
	if num <= 2 {
		return 1
	}
	return fib(num-2) + fib(num-1)
}
func TestWork(t *testing.T) {
	logger := log.New(os.Stdout, "", log.Lshortfile|log.LstdFlags)
	F := func(args interface{}) error {
		// t.Logf("call func: args: %v\n", args)
		num, _ := args.(int)
		ret := fib(num)
		logger.Printf("fib(%d) = %d\n", num, ret)
		// t.Logf("fib(%d) = %d\n", num, ret)
		return nil
	}
	cfg := Config{
		Num:    1000,
		Logger: logger,
	}
	ms := New(&cfg)
	for i := 0; i < 10000; i++ {
		ms.Commit(DoFunc{F: F, Args: i % 20})
	}
	time.Sleep(time.Second * time.Duration(60))
	// ms.Done()
	ms.Stop(-1)
}
