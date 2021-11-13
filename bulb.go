package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/koron/go-ssdp"
)

type Bulb struct {
	ID              *big.Int
	Model           string
	Name            string
	PowerStatus     bool
	Brightness      int
	Saturation      int
	Hue             int
	Temperature     int
	RGB             color.RGBA
	ColorMode       int
	Scheme          string
	Host            string
	Port            string
	Support         []string
	FirmwareVersion int
	*ssdp.Service
}

// Extracts bulb info
func ExtractBulbInfo(service *ssdp.Service) (*Bulb, error) {
	url, err := url.Parse(service.Location)
	if err != nil {
		return nil, err
	}

	header := service.Header()

	powerStatus := header.Get("Power") == "on"

	concatRGB := ParseHeaderInt("Rgb", header)
	rgb := color.RGBA{
		B: uint8(concatRGB & 0xFF),
		G: uint8((concatRGB >> 8) & 0xFF),
		R: uint8((concatRGB >> 16) & 0xFF),
	}

	bulb := &Bulb{
		ID:              ParseHeaderHex("Id", header),
		Scheme:          url.Scheme,
		Host:            url.Hostname(),
		Port:            url.Port(),
		Support:         strings.Split(header.Get("Support"), " "),
		FirmwareVersion: ParseHeaderInt("Fw_ver", header),
		ColorMode:       ParseHeaderInt("Color_mode", header),
		Brightness:      ParseHeaderInt("Bright", header),
		Temperature:     ParseHeaderInt("Ct", header),
		Hue:             ParseHeaderInt("Hue", header),
		Saturation:      ParseHeaderInt("Sat", header),
		Model:           header.Get("Model"),
		Name:            header.Get("Name"),
		PowerStatus:     powerStatus,
		RGB:             rgb,

		Service: service,
	}
	return bulb, nil
}

func ParseHeaderInt(headerName string, header http.Header) int {
	headerInt, err := strconv.Atoi(header.Get(headerName))
	if err != nil {
		log.Println(fmt.Sprintf("Can't extract header(%s)", headerName))
		headerInt = -1
	}
	return headerInt
}

func ParseHeaderHex(headerName string, header http.Header) *big.Int {
	bigint := new(big.Int)
	bigint.SetString(header.Get(headerName)[2:], 16)
	return bigint
}

func (bulb *Bulb) sendCommand(commandName string, params ...interface{}) {
	requestParameters := make(map[string]interface{})
	requestParameters["id"] = bulb.ID
	requestParameters["method"] = commandName
	requestParameters["params"] = params

	jsonRequest, err := json.Marshal(requestParameters)
	jsonRequestBytes := append([]byte(jsonRequest))
	jsonRequestBytes = append(jsonRequestBytes, '\r')
	jsonRequestBytes = append(jsonRequestBytes, '\n')
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", bulb.Host, bulb.Port))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer conn.Close()
	conn.Write([]byte(jsonRequestBytes))
}

func (bulb *Bulb) SetBrightness(brightness int, effect string, duration int) {
	bulb.sendCommand("set_bright", brightness, effect, duration)
}

func (bulb *Bulb) SetPower(powerStatus bool, effect string, duration int, mode int) {
	powerString := "off"
	if powerStatus == true {
		powerString = "on"
	}
	bulb.sendCommand("set_power", powerString, effect, duration, mode)
}
