package httpx

import (
	"fmt"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config := LoadConfig()
	fmt.Printf("%+v\n", *config)
}
