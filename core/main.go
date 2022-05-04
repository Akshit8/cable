package main

import (
	"context"
	"log"
	"sync"
	"time"
)

func main() {
	wg := &sync.WaitGroup{}

	ctx1, cancel1 := context.WithCancel(context.Background())

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx1.Done():
				log.Println("context cancelled for gr 1:", ctx1.Err())
				return
			default:
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()

	go func() {
		time.Sleep(time.Millisecond * 300)
		cancel1()
	}()

	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel2()

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			select {
			case <-ctx2.Done():
				log.Println("context cancelled for gr 2:", ctx2.Err())
				return
			default:
				time.Sleep(time.Millisecond * 200)
			}
		}
	}()

	wg.Wait()
}
