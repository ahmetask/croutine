package croutine

import (
	"runtime"
	"sync"
)

type AsyncFunc func(params ...interface{})
type SupplyAsyncFunc func(params ...interface{}) (interface{}, error)
type ErrorHandler func(error interface{}, params ...interface{})

type CRoutine interface {
	Get() Optional
	RunAsync(f AsyncFunc, params ...interface{})
	SupplyAsync(f SupplyAsyncFunc, params ...interface{}) CRoutine
	Exceptionally(h ErrorHandler) CRoutine
}

type cRoutine struct {
	run          chan bool
	result       interface{}
	err          error
	wg           *sync.WaitGroup
	params       []interface{}
	errorHandler ErrorHandler
}

func (c *cRoutine) Get() Optional {
	c.run <- true
	c.wg.Wait()
	return &Data{V: c.result}
}

func (c *cRoutine) SupplyAsync(f SupplyAsyncFunc, params ...interface{}) CRoutine {
	c.wg.Add(1)
	go func() {
		<-c.run
		defer func() {
			if r := recover(); r != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				params = append(params, string(buf))
				c.errorHandler(r, params...)
				c.wg.Done()
			}
		}()
		c.result, c.err = f(params...)
		if c.err != nil {
			c.errorHandler(c.err, params)
		}
		c.wg.Done()
	}()
	return c
}

func (c *cRoutine) RunAsync(f AsyncFunc, params ...interface{}) {
	c.wg.Add(1)
	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			params = append(params, string(buf))
			c.wg.Done()
		}
	}()
	go func() {
		<-c.run
		f(params...)
		c.wg.Done()
	}()
}

func (c *cRoutine) Exceptionally(handler ErrorHandler) CRoutine {
	c.errorHandler = handler
	return c
}

func New() CRoutine {
	wg := &sync.WaitGroup{}
	return &cRoutine{
		wg:  wg,
		run: make(chan bool),
	}
}
