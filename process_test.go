package cable

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	t.Parallel()

	pr := newProcessRunner()

	ch := make(chan int)
	const testVal = 1

	pr.Run(func() {
		time.Sleep(time.Second)
		ch <- testVal
	})

	val := <-ch

	if val != testVal {
		t.Errorf("Expected %d, got %d", testVal, val)
	}
}

func TestWait(t *testing.T) {
	t.Parallel()

	pr := newProcessRunner()

	start := time.Now()

	ch := make(chan time.Time, 1)

	pr.Run(func() {
		time.Sleep(time.Second)
		ch <- time.Now()
	})

	pr.Wait()

	val := <-ch
	diff := val.Sub(start)

	if diff < time.Second {
		t.Errorf("Expected at least 1 second, got %s", diff)
	}
}
