package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Response comment
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

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Println(msg.Payload())
}

func main() {
	fmt.Println("HSL Live")

	/* // Response-struct to store HTTP GET response
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

	options := mqtt.NewClientOptions()
	options.AddBroker("tcp://mqtt.hsl.fi:1883")
	options.SetClientID("asd")
	options.SetDefaultPublishHandler(f)

	mqttClient := mqtt.NewClient(options)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	if token := mqttClient.Subscribe("hfp/journey/tram/#", 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	time.Sleep(time.Second * 5)

	if token := mqttClient.Unsubscribe("hfp/journey/tram/#"); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}

	mqttClient.Disconnect(250)
	time.Sleep(time.Second * 2)

	/* // Signals
	signCh := make(chan os.Signal, 1)
	signal.Notify(signCh, os.Interrupt, os.Kill)

	// MQTT Client
	mqttClient := client.New(&client.Options{
		ErrorHandler: func(err error) {
			fmt.Println(err)
		},
	})
	defer mqttClient.Terminate()

	// Connect to HSL Live
	err := mqttClient.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  "mqtt.hsl.fi:1883",
		ClientID: []byte("hsl-client"),
	})
	if err != nil {
		panic(err)
	}

	// Subscribe to tram data
	err = mqttClient.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte("/hfp/journey/tram/#"),
				QoS:         mqtt.QoS1,
				Handler: func(topic, message []byte) {
					fmt.Println("message!", string(message))
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	<-signCh

	if err := mqttClient.Disconnect(); err != nil {
		panic(err)
	}  */
}