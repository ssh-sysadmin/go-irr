package main

import (
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var cache prefixCache
var conf config

func init() {
	cache.init()
	loadConfig(&conf)
	go cache.purgeEvery(time.Hour)
}

func main() {
	router := gin.Default()
	router.NoRoute(handle)

	router.Run(conf.listen)
}

func handle(c *gin.Context) {
	c.Header("Content-Type", "text/plain")

	path := strings.Split(c.Request.URL.Path, "/")[1:]
	q := c.Request.URL.Query()

	// path has 3 segments:
	// /vendor/addressfamily/AS1234:AS-SET

	if len(path) != 3 {
		c.String(404, "Not found")
		return
	}

	vendor := path[0]
	addrFamily := path[1]
	asnOrAsSet := strings.ReplaceAll(strings.ToUpper(path[2]), "_", ":")

	if vendor == "" || addrFamily == "" || !strings.HasPrefix((asnOrAsSet), "AS") {
		c.String(400, "Bad request")
		return
	}

	// Make sure asnOrAsSet is correct format
	// AS\d{1,5} or AS-SET

	isASN, _ := regexp.MatchString("^AS\\d{1,6}$", asnOrAsSet)
	isAsSet, _ := regexp.MatchString("^AS[A-Z0-9:-]{1,48}$", asnOrAsSet)

	if !isASN && !isAsSet {
		c.String(400, "Bad request")
		return
	}

	// Check if the prefix list is already cached
	// If it isn't, look it up using bgpq4 and cache the result

	output := cache.get(vendor, addrFamily, asnOrAsSet)

	if output == "" {
		output = queryBgpq4(vendor, addrFamily, asnOrAsSet)

		if output == "" {
			c.String(500, "Internal server error")
			return
		}

		cache.set(vendor, addrFamily, asnOrAsSet, output)
	}

	if q.Has("name") {
		output = strings.ReplaceAll(output, "NN", q.Get("name"))
	}

	c.String(200, output)
}
