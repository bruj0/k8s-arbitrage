package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"path/filepath"

	"github.com/davecgh/go-spew/spew"
	"github.com/fsnotify/fsnotify"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type jsonQueue struct {
	Seqno     int64
	Timestamp int64
	Payload   string
}

var queue = make(map[int64][]*gopacket.Packet)
var jq []*jsonQueue
var id = "pod0"

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	log.Printf("Starting k8s-arbitrage, watching /data\n")
	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if (event.Op & fsnotify.Create) == fsnotify.Create {
					log.Println("New file:", event.Name)
					if filepath.Ext(event.Name) == ".pcap" {
						arbitrage(event.Name)
						fn := event.Name + fmt.Sprintf("%d", time.Now().UnixNano()) + id + ".json"
						log.Printf("Saving to %s", fn)
						Save(fn, jq)
						queue = make(map[int64][]*gopacket.Packet)
						jq = nil
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add("/data")
	if err != nil {
		log.Fatal(err)
	}
	<-done

}
func arbitrage(file string) {
	if handle, err := pcap.OpenOffline(file); err != nil {
		panic(err)
	} else {
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			handlePacket(packet)
		}
	}
	for seqno, pairs := range queue {
		fmt.Printf("Found seqno=%d\n", seqno)
		for _, p := range pairs {
			fmt.Printf("ts 0=%d,payload=%s\n",
				(*p).Metadata().Timestamp.UnixNano(),
				(*p).ApplicationLayer().Payload())
			fmt.Printf("Packet=%s", spew.Sdump(p))
			jq = append(jq, &jsonQueue{
				Seqno:     seqno,
				Timestamp: (*p).Metadata().Timestamp.UnixNano(),
				Payload:   fmt.Sprintf("%s", (*p).ApplicationLayer().Payload()),
			})
		}
	}
}

func handlePacket(packet gopacket.Packet) {

	if app := packet.ApplicationLayer(); app != nil {
		payload := app.Payload()
		fields := strings.Split(string(payload), " ")
		if len(fields) == 6 {
			seqno, err := strconv.ParseInt(fields[3], 10, 64)
			if err != nil {
				fmt.Printf("Error parsing seqno:%s", err)
				return
			}
			if len(queue[seqno]) < 2 {
				queue[seqno] = append(queue[seqno], &packet)
			}

		}
	}
	fmt.Printf("Error getting applycation layer:%s", spew.Sdump(packet))
}

func Save(path string, v interface{}) error {

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	r, err := Marshal(v)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

var Marshal = func(v interface{}) (io.Reader, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}
