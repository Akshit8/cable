package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Akshit8/cable"
)

func main() {
	c := cable.NewProcessManager()

	c.AddRunnableProcess(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Println("Process 1")
				time.Sleep(time.Second)
			}
		}
	})

	c.AddRunnableProcess(func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				fmt.Println("Process 2")
				time.Sleep(time.Millisecond * 500)
			}
		}
	})

	c.AddCleanupProcess(func() error {
		fmt.Println("Cleanup Process 1 and wait for 1 second")
		time.Sleep(time.Second)
		return nil
	})

	c.AddCleanupProcess(func() error {
		fmt.Println("Cleanup Process 2 and wait for 2 second")
		time.Sleep(time.Second * 2)
		return nil
	})

	<-c.Done()
}
