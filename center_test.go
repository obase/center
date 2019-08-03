package center

import (
	"fmt"
	"testing"
)

func TestDiscovery(t *testing.T) {
	ss, err := Discovery("pvpbroker")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ss)
}
