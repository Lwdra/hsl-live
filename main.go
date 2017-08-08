package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Response comment, TKL
type Response struct {
	Body []struct {
		Journey struct {
			Line     string `json:"lineRef"`
			Location struct {
				Lat  string `json:"latitude"`
				Long string `json:"longitude"`
			} `json:"vehicleLocation"`
			Vehicle string `json:"vehicleRef"`
		} `json:"monitoredVehicleJourney"`
	} `json:"body"`
}

// Tram comment, HSL
type Tram struct {
	VP struct {
		Line    string          `json:"desi"`
		Vehicle string          `json:"veh"`
		Lat     json.RawMessage `json:"lat"`
		Long    json.RawMessage `json:"long"`
	} `json:"VP"`
}

// Devices comment
type Devices struct {
	Size  int `json:"fullSize"`
	Items []struct {
		ID   string `json:"deviceId"`
		Name string `json:"name"`
		Manu string `json:"manufacturer"`
	} `json:"items"`
}

// TramID comment
type TramID struct {
	ID   string
	Name string
}

// TramIds map
var TramIds = make(map[string]string)

// REST Get function
func getData(response interface{}, url string) error {

	// Create a http client with a 10 sec timeout
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}

	// HTTP GET
	r, err := httpClient.Get(url)
	if err != nil {
		panic(err)
	}
	defer r.Body.Close()

	// Decode response JSON to Response-struct
	return json.NewDecoder(r.Body).Decode(response)
}

func sendToIoT(tram Tram, ID string) {
	fmt.Println("Sending data")
	url := strings.Join([]string{"https://milan_knp:Devbtiu202020@my.iot-ticket.com/api/v1/process/write/", ID}, "")
	var httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
	jsonStr := strings.Join([]string{`[{"name":"lat", "v":"`, string(tram.VP.Lat), `"}, {"name":"lon", "v":"`, string(tram.VP.Long), `"}, {"name":"num", "v":"`, tram.VP.Line, `"}]`}, "")
	jsonBa := []byte(jsonStr)

	_, err := httpClient.Post(url, "application/json", bytes.NewBuffer(jsonBa))
	if err != nil {
		fmt.Println(err)
	}
}

// Subscription message handler for MQTT
var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	tram := new(Tram)

	json.Unmarshal(msg.Payload(), &tram)

	fmt.Println(tram.VP.Line, string(tram.VP.Lat), string(tram.VP.Long), tram.VP.Vehicle)
	if val, ok := TramIds[tram.VP.Vehicle]; ok {
		go sendToIoT(*tram, val)
	}
}

func main() {
	fmt.Println("Starting...")

	// Get existing devices
	//TramIds := make(map[string]string)
	repeat := 0
	offset := 0

	for offset <= repeat {

		devicesURL := strings.Join([]string{"https://milan_knp:Devbtiu202020@my.iot-ticket.com/api/v1/devices?limit=100&offset=", strconv.Itoa(offset)}, "")
		devices := new(Devices)
		getData(&devices, devicesURL)

		for _, device := range devices.Items {
			if device.Manu == "HKL" {
				if _, ok := TramIds[device.Name]; !ok {
					TramIds[device.Name] = device.ID
				}
			}
		}

		repeat = devices.Size
		offset += 100
	}

	/* // REST
	 // Response-struct to store HTTP GET response
	response := new(Response)
	// TKL Live API
	url := "http://data.itsfactory.fi/journeys/api/1/vehicle-activity"

	for {
		getData(&response, url)

		sort.Slice(response.Body, func(i, j int) bool {
			return response.Body[i].Journey.Line < response.Body[j].Journey.Line
		})

		// Loop through elements in response.Body
		for _, val := range response.Body {
			fmt.Println(val.Journey.Line, val.Journey.Location.Lat, val.Journey.Location.Long, val.Journey.Vehicle)
		}

		time.Sleep(time.Second * 1)
	} */

	//MQTT

	signCh := make(chan os.Signal, 1)
	signal.Notify(signCh, syscall.SIGINT, syscall.SIGTERM)

	options := mqtt.NewClientOptions()
	options.AddBroker("tcp://mqtt.hsl.fi:1883")
	options.SetDefaultPublishHandler(handler)
	options.SetClientID("client")

	mqttClient := mqtt.NewClient(options)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := mqttClient.Subscribe("/hfp/journey/tram/#", 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	exit := false
	for exit != true {
		select {
		case sign := <-signCh:
			if token := mqttClient.Unsubscribe("/hfp/journey/tram/#"); token.Wait() && token.Error() != nil {
				fmt.Println(token.Error())
				os.Exit(1)
			}
			fmt.Println(sign)
			mqttClient.Disconnect(250)
			time.Sleep(time.Second * 2)
			exit = true
		default:
		}
	}
}
