package center

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestDiscovery(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(FetchService("target"))
		time.Sleep(time.Second)
	}
}

func TestLockMap(t *testing.T) {
	p := 100
	times := 100 * 10000
	start := time.Now().UnixNano()
	test2(p, times)
	end := time.Now().UnixNano()
	fmt.Printf("used: %v\n", (end - start))
}

func test1(p int, times int) {
	m := new(sync.Map)
	wg := new(sync.WaitGroup)
	for j := 0; j < p; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < times; i++ {
				m.Store(i, i)
				m.Load(i)
				if i > 100 {
					m.Delete(i)
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()
}

type Map struct {
	Data map[int]int
	Mutx *sync.RWMutex
}

func test2(p int, times int) {
	m := &Map{
		Data: make(map[int]int),
		Mutx: new(sync.RWMutex),
	}
	wg := new(sync.WaitGroup)
	for j := 0; j < p; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < times; i++ {
				m.Mutx.Lock()
				m.Data[i] = i
				m.Mutx.Unlock()
				m.Mutx.RLock()
				_ = m.Data[i]
				m.Mutx.RUnlock()
				if i > 100 {
					m.Mutx.Lock()
					delete(m.Data, i)
					m.Mutx.Unlock()
				}
				runtime.Gosched()
			}
		}()
	}
	wg.Wait()
}
