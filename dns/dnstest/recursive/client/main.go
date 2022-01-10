package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/awgh/bencrypt/ecc"
	"github.com/awgh/netutils"
	transport "github.com/awgh/ratnet-transports/dns"
	"github.com/awgh/ratnet/api"
	"github.com/awgh/ratnet/api/events/defaultlogger"
	"github.com/awgh/ratnet/nodes/ram"
	"github.com/awgh/ratnet/policy/poll"
	"github.com/kelseyhightower/envconfig"
)

const (
	domain   = "nonexistant.tld"
	interval = 5000 // milliseconds
	jitter   = 10
)

var (
	testNode       *TestNode
	config         Config
	nameserverHost string
)

type Config struct {
	ServerCID string `required:"true" envconfig:"SERVER_CID"`
}

type TestNode struct {
	Node       api.Node
	ClientConv uint32
	ServerConv uint32
	tport      api.Transport
}

func initNode() {
	t := &TestNode{}
	contentKey := new(ecc.KeyPair)
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

func (t *TestNode) Poll() {
	policy := poll.New(t.tport, t.Node, interval, jitter)
	t.Node.SetPolicy(policy)
	t.Node.AddPeer("server", true, nameserverHost)
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
	err := envconfig.Process("client", &config)
	if err != nil {
		panic(err)
	}

	pubKey := ecc.PubKey{}
	err = pubKey.FromB64(config.ServerCID)
	if err != nil {
		panic(err)
	}

	nameserverHost = findNamserver()
	log.Println("upstream nameserver:", nameserverHost)

	log.Println("client starting up")
	initNode()

	testNode.Node.AddContact("server", config.ServerCID)

	cid, err := testNode.CID()
	if err != nil {
		panic(err)
	}
	log.Println("CID:", cid)

	testNode.Poll()
	err = testNode.Node.Start()
	if err != nil {
		panic(err)
	}

	i := 0
	for {
		log.Printf("sending packet %d...", i)
		message := fmt.Sprintf("hello %d", i)
		bytebuf := bytes.NewBufferString(message)
		msg := api.Msg{
			Name:         "server",
			Content:      bytebuf,
			IsChan:       false,
			PubKey:       &pubKey,
			Chunked:      false,
			StreamHeader: false,
		}
		err := testNode.Node.SendMsg(msg)
		i += 1
		if err != nil {
			panic(err)
		}
		time.Sleep(5 * time.Second)
	}
}
