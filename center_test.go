package center

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestDiscovery(t *testing.T) {
	for i := 0; i < 10; i++ {
		fmt.Println(time.Now().Format("2006-01-02 15:04:05 ") + "this is at " + strconv.Itoa(i))
		ss, err := Discovery("target")
		if err != nil {
			t.Fatal(err)
		}
		for _, s := range ss {
			fmt.Println(*s)
		}
	}

}
