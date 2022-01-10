package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/awgh/bencrypt/ecc"
	"github.com/awgh/netutils"
	transport "github.com/awgh/ratnet-transports/dns"
	"github.com/awgh/ratnet/api"
	"github.com/awgh/ratnet/api/events/defaultlogger"
	"github.com/awgh/ratnet/nodes/ram"
	"github.com/awgh/ratnet/policy/server"
)

const (
	domain      = "nonexistant.tld"
	interval    = 500 // milliseconds
	jitter      = 10
	keyfilename = "server_cid"
)

var (
	testNode       *TestNode
	nameserverHost string
)

type TestNode struct {
	Node       api.Node
	ClientConv uint32
	ServerConv uint32
	tport      api.Transport
}

func writeKeyFile(cid *ecc.KeyPair) {
	ioutil.WriteFile(keyfilename, []byte(cid.ToB64()), 0644)
}

func loadKeyFile() (cid *ecc.KeyPair) {
	cid = new(ecc.KeyPair)
	cid.GenerateKey()
	keyfile, err := ioutil.ReadFile(keyfilename)
	if err != nil {
		writeKeyFile(cid) // if there is no config, write one
		return
	}
	cid.FromB64(string(keyfile))

	return cid
}

func initNode() {
	t := &TestNode{}

	contentKey := loadKeyFile()
	routingKey := new(ecc.KeyPair)
	t.Node = ram.New(contentKey, routingKey)

	defaultlogger.StartDefaultLogger(t.Node, api.Info)

	t.ClientConv = 0xffffffff
	t.ServerConv = 0xffffffff
	tmap := make(map[string]interface{})
	tmap["ListenStr"] = ":53"
	tmap["UpstreamStr"] = nameserverHost
	tmap["Domain"] = domain

	t.tport = transport.NewFromMap(t.Node, tmap)

	testNode = t
}

func (t *TestNode) CID() (ret string, err error) {
	cid, err := t.Node.CID()
	if err != nil {
		return
	}
	ret = cid.ToB64()
	return
}

func (t *TestNode) Serve() {
	policy := server.New(t.tport, ":53", false)
	t.Node.SetPolicy(policy)
}

func findNamserver() string {
	nameservers, err := netutils.GetDNSForIP(netutils.GetOutboundIP())
	if err != nil {
		panic("namesever1")
	}
	if len(nameservers) < 1 {
		panic("nameserver2")
	}
	host := fmt.Sprintf("%s:53", nameservers[0].String())
	return host
}

func main() {
	log.Println("server starting up")
	nameserverHost = findNamserver()
	log.Println("upstream nameserver:", nameserverHost)

	initNode()

	cid, err := testNode.CID()
	if err != nil {
		panic(err)
	}
	log.Println("CID:", cid)

	testNode.Serve()
	err = testNode.Node.Start()
	if err != nil {
		panic(err)
	}

	for {
		res := <-testNode.Node.Out()
		log.Printf("New message: '%s'\n", res.Content.String())
	}
}
