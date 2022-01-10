#!/bin/bash

exec go test -race -tags debug -v -timeout 0 github.com/awgh/ratnet-transports/dns/dnstest
