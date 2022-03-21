// Copyright (c) 2022 Andreas Sauerwein. All rights reserved.
// Use of this source code is governed by the MIT License that can be found in
// the LICENSE file.

package main

// Connect to the broker, subscribe, and write messages received to a file

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	hec "github.com/jhop310/splunk-hec-go"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

var conf config

// config is a stuct for holding runtime configuration
type config struct {
	Broker             string   `yaml:"broker" envconfig:"BROKER"`
	MqttUsername       string   `yaml:"mqtt_username" envconfig:"MQTT_USERNAME"`
	MqttPassword       string   `yaml:"mqtt_password" envconfig:"MQTT_PASSWOWRD"`
	HecUrl             string   `yaml:"hec_url" envconfig:"HEC_URL"`
	HecToken           string   `yaml:"hec_token" envconfig:"HEC_TOKEN"`
	ClientId           string   `yaml:"client_id" envconfig:"CLIENT_ID"`
	WriteToConsole     bool     `yaml:"write_to_console" envconfig:"WRITE_TO_CONSOLE"`
	WriteToSplunk      bool     `yaml:"write_to_splunk" envconfig:"WRITE_TO_SPLUNK"`
	Topics             []string `yaml:"topics" envconfig:"TOPICS"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify" envconfig:"INSECURE_SKIP_VERIFY"`
}

// handler is a simple struct that provides a function to be called when a message is received. The message is parsed
// and the count followed by the raw message is written to the file (this makes it easier to sort the file)
type handler struct {
	spl hec.HEC
}

// read config.yaml file and fill conf variable
func ReadconfigFile(cfg *config) {
	yfile, err := ioutil.ReadFile("config.yaml")

	if err != nil {
		fmt.Println("Error opening file: ", err)
	}
	err2 := yaml.Unmarshal(yfile, &conf)
	if err2 != nil {

		log.Fatal(err2)
	}
}

// read config from environment variables, overwrites config file
func ReadconfigEnv(cfg *config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		fmt.Println(err)
	}
}

func NewHandler() *handler {
	var spl hec.HEC
	if conf.WriteToSplunk {
		spl = hec.NewClient(conf.HecUrl, conf.HecToken)
		spl.SetHTTPClient(&http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.InsecureSkipVerify},
		}})
	}
	return &handler{spl: spl}
}

// handle is called when a message is received
func (o *handler) handle(_ mqtt.Client, msg mqtt.Message) {
	// We extract the count and write that out first to simplify checking for missing values
	var message map[string]interface{}
	err := json.Unmarshal(msg.Payload(), &message)
	if err != nil {
		fmt.Printf("Message could not be parsed (%s): %s", msg.Payload(), err)
	}

	if conf.WriteToConsole {
		fmt.Printf("received message: %s\n", msg.Payload())
	}
	if conf.WriteToSplunk {
		message["TOPIC"] = msg.Topic()
		payload, jsonErr := json.Marshal(message)
		if jsonErr != nil {
			fmt.Printf("ERROR creating json payload: %s\n", err.Error())
		}
		event1 := hec.NewEvent(string(payload))
		err := o.spl.WriteEvent(event1)
		if err != nil {
			fmt.Printf("ERROR writing to splunk: %s\n", err.Error())
		}
	}
}

func main() {
	// read config from file
	ReadconfigFile(&conf)
	// overwrite config with env variable
	ReadconfigEnv(&conf)

	// print some information
	fmt.Println("MQTT Broker: ", conf.Broker)
	fmt.Println("MQTT Username: ", conf.MqttUsername)
	fmt.Println("Splunk HEC URL: ", conf.HecUrl)
	fmt.Println("Write to Console enabled: ", conf.WriteToConsole)
	fmt.Println("Write to Splunk enabled: ", conf.WriteToSplunk)

	// Enable logging by uncommenting the below
	// mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	// mqtt.CRITICAL = log.New(os.Stdout, "[CRITICAL] ", 0)
	// mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	// Create a handler that will deal with incoming messages
	h := NewHandler()

	// Now we establish the connection to the mqtt broker
	opts := mqtt.NewClientOptions()
	opts.AddBroker(conf.Broker)
	opts.SetClientID(conf.ClientId)

	opts.SetOrderMatters(false)       // Allow out of order messages (use this option unless in order delivery is essential)
	opts.ConnectTimeout = time.Second // Minimal delays on connect
	opts.WriteTimeout = time.Second   // Minimal delays on writes
	opts.KeepAlive = 10               // Keepalive every 10 seconds so we quickly detect network outages
	opts.PingTimeout = time.Second    // local broker so response should be quick
	opts.Username = conf.MqttUsername // mqtt broker username
	opts.Password = conf.MqttPassword // mqtt broker password

	// Automate connection management (will keep trying to connect and will reconnect if network drops)
	opts.ConnectRetry = true
	opts.AutoReconnect = true

	// If using QOS2 and CleanSession = FALSE then it is possible that we will receive messages on topics that we
	// have not subscribed to here (if they were previously subscribed to they are part of the session and survive
	// disconnect/reconnect). Adding a DefaultPublishHandler lets us detect this.
	opts.DefaultPublishHandler = func(_ mqtt.Client, msg mqtt.Message) {
		fmt.Printf("UNEXPECTED MESSAGE: %s\n", msg)
	}

	// Log events
	opts.OnConnectionLost = func(cl mqtt.Client, err error) {
		fmt.Println("connection lost")
	}

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("connection established")

		// Establish the subscription - doing this here means that it will happen every time a connection is established
		// (useful if opts.CleanSession is TRUE or the broker does not reliably store session data)
		for _, topic := range conf.Topics {
			t := c.Subscribe(topic, 1, h.handle)
			// the connection handler is called in a goroutine so blocking here would hot cause an issue. However as blocking
			// in other handlers does cause problems its best to just assume we should not block
			go func(newTopic string) {
				_ = t.Wait() // Can also use '<-t.Done()' in releases > 1.2.0
				if t.Error() != nil {
					fmt.Printf("ERROR SUBSCRIBING: %s\n", t.Error())
				} else {
					fmt.Println("subscribed to: ", newTopic)
				}
			}(topic)
		}
	}
	opts.OnReconnecting = func(mqtt.Client, *mqtt.ClientOptions) {
		fmt.Println("attempting to reconnect")
	}

	//
	// Connect to the broker
	//
	client := mqtt.NewClient(opts)

	// If using QOS2 and CleanSession = FALSE then messages may be transmitted to us before the subscribe completes.
	// Adding routes prior to connecting is a way of ensuring that these messages are processed
	// client.AddRoute(TOPIC, h.handle)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("Connection is up")

	// Messages will be delivered asynchronously so we just need to wait for a signal to shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	<-sig
	fmt.Println("signal caught - exiting")
	client.Disconnect(1000)
	fmt.Println("shutdown complete")
}
