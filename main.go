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

func skipLines(s string, n int) string {
	return strings.SplitN(s, "\n", n+1)[n]
}
func stripHeadersForEos(prefixList string) string {
	output := skipLines(prefixList, 2)

	if strings.Contains(output, "deny") {
		output = skipLines(prefixList, 1)
	}

	return output
}

func handle(c *gin.Context) {
	path := strings.Split(c.Request.URL.String(), "/")

	c.Header("Content-Type", "text/plain")

	// Minimum path length is 4
	// /vendor/addressfamily/AS1234:AS-SET

	if len(path) != 4 {
		c.String(404, "Not found")
		return
	}

	routerOs := vendorShorthand(path[1])
	addressFamily := addrFamilyShorthand(path[2])
	asnOrAsSet := strings.ToUpper(strings.Split(path[3], "?")[0])
	nameParam := strings.Split(path[3], "?") // Optional

	prefixListName := "NN"

	if len(nameParam) > 1 {
		prefixListName = strings.Split(nameParam[1], "=")[1]
	}

	if routerOs == "" || addressFamily == "" || !strings.HasPrefix((asnOrAsSet), "AS") {
		c.String(400, "Bad request")
		return
	}

	// Make sure asnOrAsSet is correct format
	// AS\d{1,5} or AS-SET

	isASN, _ := regexp.MatchString("^AS\\d{1,6}$", asnOrAsSet)
	isAsSet, _ := regexp.MatchString("^AS[A-Z0-9:-]{1,48}$", asnOrAsSet)
	isEosAsSet, _ := regexp.MatchString("^AS[A-Z0-9_-]{1,48}$", asnOrAsSet)

	if !isASN && !isAsSet && !isEosAsSet {
		c.String(400, "Bad request")
		return
	}

	// Check if the prefix list is in the cache
	//	- If it is, return it
	//	- If it is not, call getPrefixList and store the result in the cache
	//	- Return the result

	cacheData := cache.get(routerOs, addressFamily, asnOrAsSet)

	if cacheData != "" {
		if prefixListName != "" {
			cacheData = strings.ReplaceAll(cacheData, "NN", prefixListName)
		}

		if path[1] == "eos" {
			cacheData = stripHeadersForEos(cacheData)
		}

		c.String(200, cacheData)
		return
	}

	output := queryBgpq4(addressFamily, routerOs, asnOrAsSet, isEosAsSet)

	if output == "" {
		c.String(500, "Internal server error")
		return
	}

	cache.set(routerOs, addressFamily, asnOrAsSet, output)

	if prefixListName != "" {
		output = strings.ReplaceAll(output, "NN", prefixListName)
	}

	if path[1] == "eos" {
		output = stripHeadersForEos(output)
	}

	c.String(200, output)
}

func main() {
	router := gin.Default()
	router.NoRoute(handle)

	router.Run(conf.listen)
}
