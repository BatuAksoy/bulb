package main

import (
	"github.com/koron/go-ssdp"
)

func main() {
	ssdp.SetMulticastSendAddrIPv4("239.255.255.250:1982")
	list, err := ssdp.Search("wifi_bulb", 5, "")
	if err != nil {
		return
	}

	if len(list) == 0 {
		panic("Cannot spot the bulb")
	}

	bulb, err := ExtractBulbInfo(&list[0])
	if err != nil {
		panic(err)
	}

	bulb.SetPower(true, "smooth", 500, 0)
}
