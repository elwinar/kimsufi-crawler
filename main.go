package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const endpoint = "https://ws.ovh.com/dedicated/r2/ws.dispatcher/getAvailability2"

var command string

func init() {
	switch runtime.GOOS {
	case "linux":
		command = "notify-send"
	case "windows":
		command = "notify-send.exe"
	}
}

type Payload struct {
	Answer struct {
		Availability []struct {
			Reference string `json:"reference"`
			MetaZones []struct {
				Availability string `json:"availability"`
				Zone         string `json:"zone"`
			} `json:"metaZones"`
		} `json:"availability"`
	} `json:"answer"`
}

func main() {
	var (
		referencesList = flag.String("references", "160sk6,160sk5,160sk4,160sk41,160sk42,160sk3,160sk31,160sk32,160sk2,160sk21,160sk22,160sk23,161sk2,160sk1", "comma-separated list of references to find")
		metazonesList  = flag.String("metazones", "fr", "comma-separated list of metazones to find")
		interval       = flag.Duration("interval", 10*time.Second, "duration to wait between two checks")
	)
	flag.Parse()

	references := strings.Split(*referencesList, ",")
	metazones := strings.Split(*metazonesList, ",")

	log.Println("references:", references)
	log.Println("metazones:", metazones)

	for {
		go func() {
			log.Println("checking")
			res, err := http.Get(endpoint)
			if err != nil {
				panic(err)
			}

			raw, err := ioutil.ReadAll(res.Body)
			if err != nil {
				panic(err)
			}

			var payload Payload
			err = json.Unmarshal(raw, &payload)
			if err != nil {
				panic(err)
			}

			for _, server := range payload.Answer.Availability {
				if !contains(references, server.Reference) {
					continue
				}

				for _, metazone := range server.MetaZones {
					if !contains(metazones, metazone.Zone) {
						continue
					}

					if metazone.Availability == "unavailable" {
						log.Println("unavailable", server.Reference, metazone.Zone)
						continue
					}

					log.Println("found", server.Reference, metazone.Zone)
					exec.Command("notify-send", "Server available", fmt.Sprintf("server %s available in %s", server.Reference, metazone.Zone)).Run()
				}
			}
		}()
		time.Sleep(*interval)
	}
}

func contains(haystack []string, needle string) bool {
	for _, straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}
