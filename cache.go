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

func (c prefixCache) get(routerOs string, addrFamily string, asnOrAsSet string) string {
	return c.data[routerOs][addrFamily][asnOrAsSet]
}

func (c *prefixCache) set(routerOs string, addrFamily string, asnOrAsSet string, v string) {
	if c.data[routerOs] == nil {
		c.data[routerOs] = make(map[string]map[string]string)
	}

	if c.data[routerOs][addrFamily] == nil {
		c.data[routerOs][addrFamily] = make(map[string]string)
	}

	c.data[routerOs][addrFamily][asnOrAsSet] = v
}
