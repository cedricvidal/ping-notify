package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	pushbullet "github.com/mitsuse/pushbullet-go"
	"github.com/mitsuse/pushbullet-go/requests"
	ping "github.com/sparrc/go-ping"
)

var usage = `
Usage:
	PB_KEY=YOUR_PUSHBULLET_KEY ping-notify host
	ping-notify -k YOUR_PUSHBULLET_KEY host
`

func notify(token string, title string, body string) {
	// Create a client for Pushbullet.
	pb := pushbullet.New(token)

	// Create a push. The following codes create a note, which is one of push types.
	n := requests.NewNote()
	n.Title = title
	n.Body = body

	// Send the note via Pushbullet.
	if _, err := pb.PostPushesNote(n); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		return
	}
}

func main() {
	pbKey := flag.String("k", os.Getenv("PB_KEY"), "Pushbullet key")
	flag.Usage = func() {
		fmt.Printf(usage)
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		return
	}
	hostname := flag.Arg(0)
	fmt.Printf("Pinging %s. Will notify with Pushbullet when site is back up\n", hostname)

	pinger, err := ping.NewPinger(hostname)
	if err != nil {
		panic(err)
	}
	pinger.Count = 1
	pinger.Timeout = time.Second * 1

	// listen for ctrl-C signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			pinger.Stop()
		}
	}()

	for {
		pinger.Run()                 // blocks until finished
		stats := pinger.Statistics() // get send/receive/rtt stats
		if stats.PacketLoss > 0 {
			fmt.Printf("%s is down!\n", stats.Addr)
		} else {
			message := fmt.Sprintf("%s is up!", hostname)
			fmt.Print(message)
			notify(*pbKey, message, "Yeah")
			return
		}
		time.Sleep(time.Second * 5)
	}
}
