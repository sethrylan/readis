package main

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

func randtype() string {
	types := []string{
		"set",
		"zset",
		"hash",
		"string",
		"list",
	}
	n := rand.Int() % len(types)
	return types[n]
}

var allkeys = [...]list.Item{
	Key{name: "Raspberry Pi’s", datatype: randtype(), size: uint64(rand.Intn(100)), ttl: time.Duration(rand.Intn(100000000000))},
	Key{name: "Nutella", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Bitter melon", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Nice socks", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Eight hours of sleep", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Cats", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Plantasia, the album", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Pour over coffee", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "VR", datatype: randtype(), size: 12, ttl: 0},
	Key{name: "Noguchi Lamps", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Linux", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Business school", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Pottery", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Shampoo", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Table tennis", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Milk crates", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Afternoon tea", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Stickers", datatype: "hash", size: 12, ttl: 0},
	Key{name: "20° Weather", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Warm light", datatype: "hash", size: 12, ttl: 0},
	Key{name: "The vernal equinox", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Gaffer’s tape", datatype: "hash", size: 12, ttl: 0},
	Key{name: "Terrycloth", datatype: "hash", size: 12, ttl: 0},
}

func (*Data) ScanMock(n int) (int, int, []list.Item) {
	n = min(n, len(allkeys))
	return n, n * 100, allkeys[:n]
}
