package main

import (
	"errors"
	"github.com/ahmetask/croutine"
	"log"
	"os"
	"os/signal"
)

func A(v ...interface{}) (interface{}, error) {
	log.Printf("Test %v", v)
	if len(v) == 2 {
		log.Println(v)
		panic("panic error exist")
	} else if len(v) == 1 {
		return nil, errors.New("test")
	}

	return B{Id: 2}, nil
}

func GHandler(err interface{}, params ...interface{}) {
	log.Printf("error exist : %v, trace:%v", err, params[len(params)-1])
}

type B struct {
	Id int
}

func main() {
	c := croutine.New().
		SupplyAsync(A, 1,2).
		Exceptionally(GHandler)

	log.Printf("Result:%v", c.Get().OrElse(func() interface{} {
		return "orElse"
	}))
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}
