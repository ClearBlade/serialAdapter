package main

//Deadline needs to be specified before every read/write invocation
//timeoutDuration := 5 * time.Second
//conn.SetReadDeadline(time.Now().Add(timeoutDuration))
//

// TODO - Leave serial port open for now, but we may need to add code to close after reads
// TODO - May need to add code to close serial port, possibly restart adapter, on read errors

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	genericserialport "serialAdapter/genericSerialPort"

	adapter_library "github.com/clearblade/adapter-go-library"
	mqttTypes "github.com/clearblade/mqtt_parsing"
)

const (
	msgSubscribeQos = 0
	msgPublishQos   = 0
	serialRead      = "receive"
	serialWrite     = "send"
)

var (
	//adapter_library configuration structure
	adapterConfig      *adapter_library.AdapterConfig
	deviceName         string //Defaults to serialAdapter
	serialPort         *genericserialport.SerialPort
	readTimeout        int
	cbSubscribeChannel <-chan *mqttTypes.Publish
	endWorkersChannel  chan string
	isHalfDuplex       = false
	topicRoot          string
	adapterSettings    map[string]interface{}

	serialPortLock = &sync.Mutex{}
)

func init() {
	flag.IntVar(&readTimeout, "readTimeout", 500, "The number of millisconds to wait before timing out the serial port read.")
}

func main() {
	fmt.Println("Starting serialAdapter...")

	err := adapter_library.ParseArguments(deviceName)
	if err != nil {
		log.Fatalf("[FATAL] Failed to parse arguments: %s\n", err.Error())
	}

	// Initialize all things ClearBlade, includes authenticating if needed, and fetching the
	// relevant adapter_config collection entry
	adapterConfig, err = adapter_library.Initialize()
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize: %s\n", err.Error())
	}

	applyAdapterSettings(adapterConfig)

	//This will result in the serial port being opened
	initializeSerialPort(adapterSettings)

	//Connect to ClearBlade MQTT
	err = adapter_library.ConnectMQTT(adapterConfig.TopicRoot, cbMessageHandler)
	if err != nil {
		log.Fatalf("[FATAL] Failed to Connect MQTT: %s\n", err.Error())
	}

	log.Println("[INFO] Connected to ClearBlade Platform MQTT broker")

	log.Println("[DEBUG] Flushing serial port")
	//Flush serial port one last time
	if err := serialPort.FlushSerialPort(); err != nil {
		log.Println("[ERROR] Error flushing serial port: " + err.Error())
	}

	//Start read loop
	go readWorker()

	defer serialPort.CloseSerialPort()
	defer close(endWorkersChannel)
	endWorkersChannel = make(chan string)

	//Handle OS interrupts to shut down gracefully
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT, syscall.SIGTERM)
	sig := <-interruptChannel

	log.Printf("[INFO] OS signal %s received, ending go routines.", sig)

	//End the existing goRoutines
	endWorkersChannel <- "Stop Channel"
	endWorkersChannel <- "Stop Channel"
	os.Exit(0)
}

// ClearBlade Client init helper
func initializeSerialPort(adapterSettings map[string]interface{}) *genericserialport.SerialPort {
	serialPort = genericserialport.CreateSerialPort(adapterSettings)

	return nil
}

func cbMessageHandler(message *mqttTypes.Publish) {
	//Process incoming MQTT messages as needed here

	//Determine if a read or write request was received
	if strings.HasSuffix(message.Topic.Whole, "/read") {
		log.Println("[INFO] cbMessageHandler - Handling read request...")
		go readPort()
	} else if strings.HasSuffix(message.Topic.Whole, "/write") {
		// If write request...
		log.Println("[INFO] cbMessageHandler - Handling write request...")
		go writePort(string(message.Payload))
	} else {
		log.Printf("[DEBUG] cbMessageHandler - Unknown request received: topic = %s, payload = %#v\n", message.Topic.Whole, message.Payload)
	}
}

func readWorker() {
	log.Println("[INFO] readWorker - Starting readWorker")

	for {
		select {
		case <-endWorkersChannel:
			log.Println("[DEBUG] readWorker - stopping read worker")
			return
		default:
			readPort()
		}
	}
}

// Publishes data to a topic
func publish(topic string, data string) error {
	log.Printf("[DEBUG] publish - Publishing to topic %s\n", topic)
	adapter_library.Publish(topic, []byte(data))

	log.Printf("[DEBUG] publish - Successfully published message to = %s\n", topic)
	return nil
}

func applyAdapterSettings(adapterConfig *adapter_library.AdapterConfig) {
	if err := json.Unmarshal([]byte(adapterConfig.AdapterSettings), &adapterSettings); err != nil {
		log.Printf("[ERROR] applyAdapterSettings - Error while unmarshalling json: %s. Defaulting all adapter settings.\n", err.Error())
	}

	//topic root
	if adapterConfig.TopicRoot != "" {
		log.Printf("[DEBUG] applyAdapterSettings - Setting topicRoot to %s\n", adapterConfig.TopicRoot)
		topicRoot = adapterConfig.TopicRoot
	} else {
		log.Printf("[DEBUG] getAdapterConfig - Topic root is empty. Using default value %s\n", topicRoot)
	}

	if adapterSettings["serialPortName"] == nil || adapterSettings["serialPortName"] == "" {
		panic("serialPortName not provided. ")
	}
}

func readPort() {
	if isHalfDuplex {
		serialPortLock.Lock()
		defer serialPortLock.Unlock()
	}

	readFromSerialPort()
}

func writePort(payload string) {
	if isHalfDuplex {
		serialPortLock.Lock()
		defer serialPortLock.Unlock()
	}

	writeToSerialPort(payload)
}

func readFromSerialPort() {
	// 1. Read all data from serial port
	// 2. Publish data to platform as string
	data, err := serialPort.ReadSerialPort()

	if err != nil && err != io.EOF {
		log.Printf("[ERROR] readFromSerialPort - ERROR reading from serial port: %s\n", err.Error())
		publishSerialIOError("read", err)
	} else {
		if data != "" {
			//If there are any slashes in the data, we need to escape them so duktape
			//doesn't throw a SyntaxError: unterminated string (line 1) error
			data = strings.Replace(data, `\`, `\\`, -1)

			log.Printf("[INFO] readFromSerialPort - Data read from serial port: %s\n", data)

			//Publish data to message broker
			err := publish(topicRoot+"/response", data)
			if err != nil {
				log.Printf("[ERROR] readFromSerialPort - ERROR publishing to topic: %s\n", err.Error())
			}
		} else {
			log.Println("[DEBUG] readFromSerialPort - No data read from serial port, skipping publish.")
		}
	}
	return
}

func writeToSerialPort(payload string) {
	log.Printf("[INFO] writeToSerialPort - Writing to serial port: %s\n", payload)
	log.Println("[DEBUG] writeToSerialPort - About to lock serialPortLock")

	err := serialPort.WriteSerialPort(string(payload))
	log.Println("[DEBUG] writeToSerialPort - Just unlocked serialPortLock")
	if err != nil {
		log.Printf("[ERROR] writeToSerialPort - ERROR writing to serial port: %s\n", err.Error())
		publishSerialIOError("read", err)
	}
	return
}

func setReadDeadline() {
	if adapterSettings["deadline"] != nil {
		log.Println("[DEBUG] CreateSerialPort - deadline specified. Applying deadline to serial port")
		deadline, err := adapterSettings["deadline"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("CreateSerialPort - Error parsing deadline: %w", err))
		}
		if int(deadline) > 0 {
			err = serialPort.SetDeadline(int(deadline))
			if err != nil {
				//TODO - Do we really want to panic?
				panic(fmt.Errorf("CreateSerialPort - Error setting deadline value: %w", err))
			}
		}
	} else {
		if adapterSettings["readDeadline"] != nil {
			log.Println("[DEBUG] CreateSerialPort - readDeadline specified. Applying readDeadline to serial port")
			readDeadline, err := adapterSettings["readDeadline"].(json.Number).Int64()
			if err != nil {
				panic(fmt.Errorf("CreateSerialPort - Error parsing readDeadline: %w", err))
			}
			if int(readDeadline) > 0 {
				err = serialPort.SetReadDeadline(int(readDeadline))
				if err != nil {
					//TODO - Do we really want to panic?
					panic(fmt.Errorf("CreateSerialPort - Error setting readDeadline value: %w", err))
				}
			}
		}
	}
}

func setWriteDeadline() {
	if adapterSettings["deadline"] != nil {
		log.Println("[DEBUG] setWriteDeadline - deadline specified. Applying deadline to serial port")
		deadline, err := adapterSettings["deadline"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("setWriteDeadline - Error parsing deadline: %w", err))
		}
		if int(deadline) > 0 {
			err = serialPort.SetDeadline(int(deadline))
			if err != nil {
				//TODO - Do we really want to panic?
				panic(fmt.Errorf("setWriteDeadline - Error setting deadline value: %w", err))
			}
		}
	} else {
		if adapterSettings["writeDeadline"] != nil {
			log.Println("[DEBUG] CreateSerialPort - writeDeadline specified. Applying writeDeadline to serial port")
			writeDeadline, err := adapterSettings["writeDeadline"].(json.Number).Int64()
			if err != nil {
				panic(fmt.Errorf("CreateSerialPort - Error parsing writeDeadline: %w", err))
			}
			if int(writeDeadline) > 0 {
				if int(writeDeadline) > 0 {
					err = serialPort.SetWriteDeadline(int(writeDeadline))
					if err != nil {
						//TODO - Do we really want to panic?
						panic(fmt.Errorf("CreateSerialPort - Error setting writeDeadline value: %w", err))
					}
				}
			}
		}
	}
}

func publishSerialIOError(operation string, ioError error) {
	var errorPayload map[string]interface{}
	errorPayload[""] = operation
	errorPayload["error"] = fmt.Errorf("%w", ioError)

	respStr, err := json.Marshal(errorPayload)
	if err != nil {
		log.Printf("[ERROR] publishSerialIOError - ERROR marshalling json: %s\n", err.Error())
	} else {
		log.Printf("[DEBUG] publishModbusResponse - Publishing error response %s to topic %s\n", string(respStr), topicRoot+"/error")
		publish(topicRoot+"/error", string(respStr))
	}
}
