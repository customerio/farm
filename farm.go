package farm

import (
	"errors"
)

type Generator func(chan<- interface{}) error

type Worker func(in interface{}) (interface{}, error)

type Runner struct {
	Results <-chan interface{}
	Errors  <-chan error
}

type context struct {
	concurrency int
	generate    Generator
	work        Worker
	in          chan interface{}
	out         chan interface{}
	done        chan bool
	errs        chan error
}

func Run(concurrency int, generate Generator, work Worker) Runner {
	c := context{
		concurrency,
		generate,
		work,
		make(chan interface{}),
		make(chan interface{}),
		make(chan bool),
		make(chan error, concurrency+1),
	}

	go populate(c)

	for i := 0; i < concurrency; i++ {
		go process(c)
	}

	go wait(c)

	return Runner{c.out, c.errs}
}

func populate(c context) {
	defer func() {
		handlePanic(c, recover(), "generating")
		close(c.in)
	}()

	if err := c.generate(c.in); err != nil {
		c.errs <- err
	}
}

func process(c context) {
	defer func() {
		handlePanic(c, recover(), "processing")
		c.done <- true
	}()

	for msg := range c.in {
		if o, err := c.work(msg); err != nil {
			c.errs <- err
			break
		} else {
			c.out <- o
		}
	}
}

func wait(c context) {
	for i := 0; i < c.concurrency; i++ {
		<-c.done
	}

	close(c.out)
	close(c.errs)
}

func handlePanic(c context, r interface{}, prefix string) {
	if r != nil {
		c.errs <- toError(r, prefix)
	}
}

func toError(r interface{}, prefix string) error {
	if err, ok := r.(error); ok {
		return err
	} else if err, ok := r.(string); ok {
		return errors.New(prefix + " work for farm paniced: " + err)
	} else {
		return errors.New(prefix + "work for farm paniced.")
	}
}
