package center

import (
	"strings"
	"unsafe"
)

var robin uint

func Robin(name string) (*Service, error) {
	services, _, err := FetchService(name)
	if err != nil {
		return nil, err
	}
	size := len(services)
	if size == 0 {
		return nil, nil
	}
	robin++
	return services[robin%uint(size)], nil
}
func Hash(name string, key string) (*Service, error) {
	services, _, err := FetchService(name)
	if err != nil {
		return nil, err
	}
	size := len(services)
	if size == 0 {
		return nil, nil
	}
	idx := MMHash([]byte(key))
	return services[idx%uint32(size)], nil
}

func HttpName(name string) string {
	if strings.HasSuffix(name, ".http") {
		return name
	}
	return name + ".http"
}

func GrpcName(name string) string {
	if strings.HasSuffix(name, ".grpc") {
		return name
	}
	return name + ".grpc"
}

func TcpName(name string) string {
	if strings.HasSuffix(name, ".tcp") {
		return name
	}
	return name + ".tcp"
}

func ProxyName(name string) string {
	if strings.HasSuffix(name, ".proxy") {
		return name
	}
	return name + ".proxy"
}

const (
	c1_32 uint32 = 0xcc9e2d51
	c2_32 uint32 = 0x1b873593
)

// GetHash returns a murmur32 hash for the data slice.
func MMHash(data []byte) uint32 {
	// Seed is set to 37, same as C# version of emitter
	var h1 uint32 = 37

	nblocks := len(data) / 4
	var p uintptr
	if len(data) > 0 {
		p = uintptr(unsafe.Pointer(&data[0]))
	}

	p1 := p + uintptr(4*nblocks)
	for ; p < p1; p += 4 {
		k1 := *(*uint32)(unsafe.Pointer(p))

		k1 *= c1_32
		k1 = (k1 << 15) | (k1 >> 17) // rotl32(k1, 15)
		k1 *= c2_32

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19) // rotl32(h1, 13)
		h1 = h1*5 + 0xe6546b64
	}

	tail := data[nblocks*4:]

	var k1 uint32
	switch len(tail) & 3 {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1_32
		k1 = (k1 << 15) | (k1 >> 17) // rotl32(k1, 15)
		k1 *= c2_32
		h1 ^= k1
	}

	h1 ^= uint32(len(data))

	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return (h1 << 24) | (((h1 >> 8) << 16) & 0xFF0000) | (((h1 >> 16) << 8) & 0xFF00) | (h1 >> 24)
}

func nvl(val int64, def int64) int64 {
	if val == 0 {
		return def
	}
	return val
}
