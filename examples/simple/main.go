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

	<-c.Done()
}
