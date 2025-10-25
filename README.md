# go-irr

A simple API for bgpq4, written in Go.

## Usage

```
GET /routeros/addressfamily/ASorAS-SET
GET /routeros/addressfamily/ASorAS-SET?name=myprefixlist
```

## Config

SOURCES = What gets passed to the -S field in bgpq4, default is NTTCOM,INTERNAL,LACNIC,RADB,RIPE,RIPE-NONAUTH,ALTDB,BELL,LEVEL3,APNIC,JPIRR,ARIN,BBOI,TC,AFRINIC,IDNIC,RPKI,REGISTROBR,CANARIE
MATCH_PARENT = If bgpq4 should match parent prefixes, not just the exact route object (enabled by default)
LISTEN = What go-irr should listen to (default [::]:8080)
CACHE_TIME = How long go-irr should cache prefixes for (default 1 hour)
ALLOW_CACHE_BYPASS = If the "bypassCache" query parameter is allowed
ALLOW_CACHE_CLEAR = If global cache is allowed to be cleared with a request to /clearCache

### Examples:

```
GET /arista/v4/AS208453

GET /arista/v6/AS208453:AS-SWEHOSTING

# For systems which do not permit ":" in the URI
GET /eos/v4/AS208453_AS-CUST
```

## Supported versions

```
/arista/
/eos/ # Short version without the prefix list headers
/juniper/
/bird/
/routeros6/
/routeros7/
/ios-xr/
/ext-acl/ # generate extended access-list
/cisco/
/json/
```

## Supported address families

```
/brand/v4/
/brand/v6/
```

## Hosted version

[https://irr.as208453.net/](https://irr.as208453.net/)

## Self hosting with Docker

1. Install docker
2. Clone the repo
3. Start using docker compose
4. go-irr is now reachable via `localhost:8080`

```
docker compose up -d
```
