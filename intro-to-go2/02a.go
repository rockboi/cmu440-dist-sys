package main

import "fmt"

const (
	N_INCREMENTS = 20
)

func main() {

	counter := 0
	donechan := make(chan bool)
	lock_chan := make(chan int, 1)
	lock_chan <- 1

	go func(done chan<- bool, lock_chan chan int) {
		for i := 0; i < N_INCREMENTS; i++ {
			_ = <-lock_chan
			counter++
			lock_chan <- 1
		}
		done <- true
	}(donechan, lock_chan)

	for i := 0; i < N_INCREMENTS; i++ {
		_ = <-lock_chan
		counter++
		lock_chan <- 1
	}

	_ = <-donechan

	fmt.Println("Count: ", counter)

}
