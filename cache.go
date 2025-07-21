package main

import "time"

type prefixCache struct {
	data map[string]map[string]map[string]string
}

func (c *prefixCache) init() {
	c.data = make(map[string]map[string]map[string]string)
}

func (c *prefixCache) purgeEvery(t time.Duration) {
	ticker := time.NewTicker(t)
	for {
		<-ticker.C
		c.init()
	}
}

func (c prefixCache) get(vendor string, addrFamily string, asnOrAsSet string) string {
	return c.data[vendor][addrFamily][asnOrAsSet]
}

func (c *prefixCache) set(vendor string, addrFamily string, asnOrAsSet string, v string) {
	if c.data[vendor] == nil {
		c.data[vendor] = make(map[string]map[string]string)
	}

	if c.data[vendor][addrFamily] == nil {
		c.data[vendor][addrFamily] = make(map[string]string)
	}

	c.data[vendor][addrFamily][asnOrAsSet] = v
}
