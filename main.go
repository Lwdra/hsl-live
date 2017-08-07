package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

// REST Get function
func getData(response interface{}) error {
	// TKL Live API
	url := "http://data.itsfactory.fi/journeys/api/1/vehicle-activity"

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

// Subscription message handler for MQTT
var handler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	tram := new(Tram)

	json.Unmarshal(msg.Payload(), &tram)
	fmt.Println(tram.VP.Line, string(tram.VP.Lat), string(tram.VP.Long), tram.VP.Vehicle)
}

func main() {
	fmt.Println("Starting...")
	/* // REST
	 // Response-struct to store HTTP GET response
	response := new(Response)

	for {
		getData(&response)

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

	//time.Sleep(time.Second * 20)

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
