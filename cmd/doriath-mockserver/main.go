package main

import (
	crand "crypto/rand"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/crypto/ed25519"

	"github.com/rensa-labs/doriath"
	"github.com/rensa-labs/doriath/internal/libkataware"
	"github.com/rensa-labs/doriath/operlog"
)

func garbageLoop(srv *doriath.Server) {
	for i := 0; ; i++ {
		pk, sk, err := ed25519.GenerateKey(crand.Reader)
		if err != nil {
			panic(err)
		}
		idsc := fmt.Sprintf(".ed25519 %x .quorum 1. 1.", pk)
		idscBin, err := operlog.AssembleID(idsc)
		if err != nil {
			panic(err)
		}
		name := fmt.Sprintf("name-%v", i)
		newop := operlog.Operation{
			PrevHash: make([]byte, 32),
			NextID:   idscBin,
			Data:     []byte(fmt.Sprintf("garbage-data-%v", name)),
		}
		crand.Read(newop.PrevHash)
		signature := ed25519.Sign(sk, newop.SignedPart())
		newop.Signatures = [][]byte{signature}
		srv.StageOperation(name, newop)
		time.Sleep(time.Second)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	mbc, bogus := doriath.NewMockBitcoinClient()
	log.Printf("%x", bogus)
	var waa libkataware.Transaction
	err := waa.FromBytes(bogus)
	if err != nil {
		panic("waaaa")
	}
	srv, err := doriath.NewServer(mbc,
		"foobar", bogus,
		time.Second*10,
		fmt.Sprintf("/tmp/doriath-mock-%v.db", time.Now().Unix()))
	if err != nil {
		panic(err.Error())
	}
	hserv := &http.Server{
		Addr:           "127.0.0.1:18888",
		Handler:        srv,
		MaxHeaderBytes: 1024 * 4,
		ReadTimeout:    time.Second * 2,
	}
	log.Println("MOCK SERVER STARTED at 127.0.0.1:18888, point nginx here")
	go garbageLoop(srv)
	err = hserv.ListenAndServe()
	if err != nil {
		panic(err.Error())
	}
}
