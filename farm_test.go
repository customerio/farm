package farm

import (
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
