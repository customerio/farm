package farm

import (
	"errors"
	"sort"
	"testing"
)

func TestRun(t *testing.T) {
	r := Run(
		10,
		func(in chan<- interface{}) error {
			for i := 0; i < 20; i++ {
				in <- i
			}
			return nil
		},
		func(val interface{}) (interface{}, error) {
			return val.(int) * 2, nil
		},
	)

	results := make([]int, 0, 20)

	for msg := range r.Results {
		results = append(results, msg.(int))
	}

	sort.Ints(results)

	for i, result := range results {
		if result != i*2 {
			t.Errorf("Result %d: wrong result: want: %d found: %d", i, i*2, result)
		}
	}

	if err := <-r.Errors; err != nil {
		t.Errorf("Error returned from run: %v", err)
	}
}

func TestGeneratorError(t *testing.T) {
	r := Run(
		10,
		func(in chan<- interface{}) error {
			return errors.New("whoops")
		},
		func(val interface{}) (interface{}, error) {
			return nil, nil
		},
	)

	var count int

	for _ = range r.Results {
		count += 1
	}

	if count > 0 {
		t.Errorf("Unexpected results returned.")
	}

	err := <-r.Errors

	if err == nil {
		t.Errorf("Error not returned from generator")
	}

	if err.Error() != "whoops" {
		t.Errorf("Wrong error message: want: %d found: %d", "whoops", err.Error())
	}
}

func TestGeneratorPanic(t *testing.T) {
	r := Run(
		10,
		func(in chan<- interface{}) error {
			panic("whoops")
		},
		func(val interface{}) (interface{}, error) {
			return nil, nil
		},
	)

	for _ = range r.Results {
	}

	err := <-r.Errors

	if err == nil {
		t.Errorf("Error not returned from generator")
	}

	if msg := "generating work for farm paniced: whoops"; err.Error() != msg {
		t.Errorf("Wrong error message: want: %d found: %d", msg, err.Error())
	}
}

func TestWorkerError(t *testing.T) {
	r := Run(
		10,
		func(in chan<- interface{}) error {
			for i := 0; i < 20; i++ {
				in <- i
			}
			return nil
		},
		func(val interface{}) (interface{}, error) {
			return val.(int) * 2, errors.New("whoops")
		},
	)

	var count int

	for _ = range r.Results {
		count += 1
	}

	if count > 0 {
		t.Errorf("Unexpected results returned.")
	}

	errs := make([]error, 0, 10)

	for err := range r.Errors {
		errs = append(errs, err)
	}

	if len(errs) != 10 {
		t.Errorf("Wrong number of errors returned: want: %d found: %d", 10, len(errs))
	}

	if errs[0].Error() != "whoops" {
		t.Errorf("Wrong error message: want: %d found: %d", "whoops", errs[0].Error())
	}
}

func TestWorkerPanic(t *testing.T) {
	r := Run(
		10,
		func(in chan<- interface{}) error {
			for i := 0; i < 20; i++ {
				in <- i
			}
			return nil
		},
		func(val interface{}) (interface{}, error) {
			panic("whoops")
		},
	)

	var count int

	for _ = range r.Results {
		count += 1
	}

	if count > 0 {
		t.Errorf("Unexpected results returned.")
	}

	errs := make([]error, 0, 10)

	for err := range r.Errors {
		errs = append(errs, err)
	}

	if len(errs) != 10 {
		t.Errorf("Wrong number of errors returned: want: %d found: %d", 10, len(errs))
	}

	if msg := "processing work for farm paniced: whoops"; errs[0].Error() != msg {
		t.Errorf("Wrong error message: want: %d found: %d", msg, errs[0].Error())
	}
}
