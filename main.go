package main

import (
	"fmt"
	"time"

	"timed-cache/timed-cache"
)

func main() {
	purgeCall := func(key, value interface{}) {
		fmt.Printf("i am purged. %v, %v\n", key, value)
	}

	tc := timed_cache.NewTimedCache(4, purgeCall)
	tc.Add(1, "a")
	time.Sleep(2 * time.Second)
	tc.Add(2, "b")
	time.Sleep(2 * time.Second)
	tc.Add(3, "c")
	time.Sleep(2 * time.Second)
	tc.Add(4, "d")

	time.Sleep(4 * time.Second)
	tc.Get(3)
	tc.PurgeExpired()

	tc.Print()
}

