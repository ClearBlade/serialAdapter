package genericserialport

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	//"github.com/tarm/serial"
	serial "github.com/mikepb/go-serial"
)

//SerialPort : Struct that represents the serial port used to interface with a serial device
type SerialPort struct {
	options  serial.Options
	port     *serial.Port
	portName string
}

func createOptionsFromAdapterConfig(adapterSettings map[string]interface{}) serial.Options {
	log.Println("[INFO] Configuring serial port options...")
	// 	{
	//     Mode
	//     BitRate
	//     DataBits:    8,
	//     Parity:      PARITY_NONE,
	//     StopBits:    1,
	//     FlowControl: FLOWCONTROL_NONE,
	//     RTS
	//     CTS
	//     DTR
	//     DSR
	// }
	var portOptions = serial.RawOptions
	portOptions.Mode = serial.MODE_READ_WRITE

	//port name
	if adapterSettings["portName"] == nil || adapterSettings["portName"].(string) == "" {
		panic("createOptionsFromAdapterConfig - Cannot create serial port. Port name is empty.")
	}

	//baud rate
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying bitRate")
	if adapterSettings["bitRate"] == nil {
		panic("createOptionsFromAdapterConfig - Cannot create serial port. bitRate (Baud rate) not specified.")
	} else {
		// bitRate, err := adapterSettings["bitRate"].(json.Number).Int64()
		// if err != nil {
		// 	panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing bitRate: %w", err))
		// }
		// switch int(bitRate) {
		var bitRate = 19200
		switch bitRate {
		case 110, 300, 600, 1200, 2400, 4800, 9600, 14400, 19200, 38400, 57600, 115200, 128000, 256000:
			portOptions.BitRate = int(bitRate)
			log.Println("[DEBUG] createOptionsFromAdapterConfig - bitRate applied")
			log.Printf("[INFO] bit rate = %d\n", portOptions.BitRate)
		default:
			panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid bitRate (Baud rate) specified: %v", adapterSettings["bitRate"]))
		}
	}

	//CTS
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying clearToSend")
	if adapterSettings["clearToSend"] != nil {
		clearToSend, err := adapterSettings["clearToSend"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing clearToSend: %w", err))
		}
		if int(clearToSend) > 0 {
			switch int(clearToSend) {
			case 1, 2:
				portOptions.CTS = int(clearToSend)
				log.Println("[DEBUG] createOptionsFromAdapterConfig - clearToSend applied")
				log.Printf("[INFO] CTS = %d\n", portOptions.CTS)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid clearToSend value specified: %v", adapterSettings["clearToSend"]))
			}
		}
	}

	//DSR
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying dataSetReady")
	if adapterSettings["dataSetReady"] != nil {
		dataSetReady, err := adapterSettings["dataSetReady"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing dataSetReady: %w", err))
		}
		if int(dataSetReady) > 0 {
			switch int(dataSetReady) {
			case 1, 2:
				portOptions.DSR = int(dataSetReady)
				log.Println("[DEBUG] createOptionsFromAdapterConfig - dataSetReady applied")
				log.Printf("[INFO] DSR = %d\n", portOptions.DSR)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid dataSetReady value specified: %v", adapterSettings["dataSetReady"]))
			}
		}
	}

	//DTR
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying dataTerminalReady")
	if adapterSettings["dataTerminalReady"] != nil {
		// dataTerminalReady, err := adapterSettings["dataTerminalReady"].(json.Number).Int64()
		// if err != nil {
		// 	panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing dataTerminalReady: %w", err))
		// }
		// if int(dataTerminalReady) > 0 {
		var dataTerminalReady = 2
		if dataTerminalReady > 0 {
			switch int(dataTerminalReady) {
			case 1, 2, 3:
				portOptions.DTR = int(dataTerminalReady)
				log.Println("[DEBUG] createOptionsFromAdapterConfig - dataTerminalReady applied")
				log.Printf("[INFO] DTR = %d\n", portOptions.DTR)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid dataSetReady value specified: %v", adapterSettings["dataTerminalReady"]))
			}
		}
	}

	//data bits
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying dataBits")
	if adapterSettings["dataBits"] != nil {
		dataBits, err := adapterSettings["dataBits"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing dataBits: %w", err))
		}
		if int(dataBits) > 0 {
			switch int(dataBits) {
			case 5, 6, 7, 8:
				portOptions.DataBits = int(dataBits)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid dataBits value specified: %v", adapterSettings["dataBits"]))
			}
		}
	}

	//flow control
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying flowControl")
	if adapterSettings["flowControl"] != nil {
		// flowControl, err := adapterSettings["flowControl"].(json.Number).Int64()
		// if err != nil {
		// 	panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing flowControl: %w", err))
		// }
		// if int(flowControl) > 0 {
		var flowControl = 0
		if flowControl > 0 {
			switch int(flowControl) {
			case 1, 2, 3, 4:
				portOptions.FlowControl = int(flowControl)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid flowControl value specified: %v", adapterSettings["flowControl"]))
			}
		}
	}

	//parity
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying parity")
	if adapterSettings["parity"] != nil {
		parity, err := adapterSettings["parity"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing parity: %w", err))
		}
		if int(parity) > 0 {
			switch int(parity) {
			case 1, 2, 3, 4, 5:
				portOptions.Parity = int(parity)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid parity value specified: %v", adapterSettings["parity"]))
			}
		}
	}

	//RTS
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying requestToSend")
	if adapterSettings["requestToSend"] != nil {
		// requestToSend, err := adapterSettings["requestToSend"].(json.Number).Int64()
		// if err != nil {
		// 	panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing requestToSend: %w", err))
		// }
		// if int(requestToSend) > 0 {
		var requestToSend = 2
		if requestToSend > 0 {
			switch int(requestToSend) {
			case 1, 2, 3:
				portOptions.RTS = int(requestToSend)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid requestToSend value specified: %v", adapterSettings["requestToSend"]))
			}
		}
	}

	//stop bits
	log.Println("[DEBUG] createOptionsFromAdapterConfig - Applying stopBits")
	if adapterSettings["stopBits"] != nil {
		stopBits, err := adapterSettings["stopBits"].(json.Number).Int64()
		if err != nil {
			panic(fmt.Errorf("createOptionsFromAdapterConfig - Error parsing stopBits: %w", err))
		}
		if int(stopBits) > 0 {
			switch int(stopBits) {
			case 1, 2:
				portOptions.FlowControl = int(stopBits)
			default:
				panic(fmt.Errorf("createOptionsFromAdapterConfig - Cannot create serial port. Invalid stopBits value specified: %v", adapterSettings["stopBits"]))
			}
		}
	}

	return portOptions
}

//CreateSerialPort :
func CreateSerialPort(adapterSettings map[string]interface{}) *SerialPort {
	var err error
	thePort := SerialPort{}

	log.Println("[DEBUG] CreateSerialPort - creating serial port options from adapter config settings")
	thePort.options = createOptionsFromAdapterConfig(adapterSettings)

	if adapterSettings["portName"] == nil || adapterSettings["portName"].(string) == "" {
		panic("CreateSerialPort - Cannot create serial port. Port name is empty.")
	}
	thePort.portName = adapterSettings["portName"].(string)
	log.Printf("[INFO] CreateSerialPort - port name set to %s\n", thePort.portName)

	log.Printf("[INFO] CreateSerialPort - opening serial port %s\n", thePort.portName)
	thePort.port, err = thePort.options.Open(thePort.portName)
	if err != nil {
		panic(fmt.Errorf("CreateSerialPort - Error opening serial port %s: %w", thePort.portName, err))
	}

	if adapterSettings["xONxOFF"] != nil {
		log.Println("[DEBUG] CreateSerialPort - xONxOFF specified. Applying xONxOFF to serial port")
		// xONxOFF, err := adapterSettings["xONxOFF"].(json.Number).Int64()
		// if err != nil {
		// 	panic(fmt.Errorf("CreateSerialPort - Error parsing xONxOFF: %w", err))
		// }
		// if int(xONxOFF) > 0 {
		var xONxOFF = 0
		if xONxOFF > 0 {
			switch int(xONxOFF) {
			case 1, 2, 3, 4:
				err = thePort.port.SetXonXoff(int(xONxOFF))
				if err != nil {
					panic(fmt.Errorf("CreateSerialPort - Error setting xONxOFF value: %w", err))
				}
				err = thePort.port.Apply(&thePort.options)
				if err != nil {
					panic(fmt.Errorf("CreateSerialPort - Error applying options after xONxOFF change: %w", err))
				}
			default:
				panic(fmt.Errorf("CreateSerialPort - Cannot apply xONxOFF to port. Invalid xONxOFF value specified: %w", adapterSettings["xONxOFF"]))
			}
		}
	}

	return &thePort
}

//OpenSerialPort :
func (serialDevice *SerialPort) OpenSerialPort() {
	var err error
	log.Println("[DEBUG] OpenSerialPort - Opening serial port")

	if serialDevice.portName == "" {
		panic("CreateSerialPort - Cannot open serial port. Port name is empty.")
	}

	if serialDevice.port != nil {
		serialDevice.port, err = serialDevice.port.OpenPort(&serialDevice.options)
		if err != nil {
			panic(fmt.Errorf("CreateSerialPort - Error opening serial port: %w", err))
		}
	} else {
		serialDevice.options.Open(serialDevice.portName)
	}

	log.Println("[INFO] OpenSerialPort - Serial port opened")
}

//CloseSerialPort :
func (serialDevice *SerialPort) CloseSerialPort() {
	var err error
	log.Println("[DEBUG] CloseSerialPort - Closing serial port")
	err = serialDevice.port.Close()
	if err != nil {
		panic(fmt.Errorf("CloseSerialPort - Error closing serial port:: %w", err))
	}
	log.Println("[INFO] CloseSerialPort - Serial port closed")
}

//ReadSerialPort :
func (serialDevice *SerialPort) ReadSerialPort() (string, error) {
	var n int
	var err error

	//buf := make([]byte, 128)
	buf := make([]byte, 2048)
	incomingData := ""

	for {
		n, err = serialDevice.port.Read(buf)
		if err != nil { // err will equal io.EOF if there is no data to read
			break
		}
		//log.Printf("[DEBUG] readSerialPort - Data read: " + string(buf))
		log.Printf("[DEBUG] readSerialPort - Hex Data read: %x\n", buf)
		log.Printf("[DEBUG] readSerialPort - Number of bytes read: %d\n", n)
		if n > 0 {
			//incomingData += string(buf[:n])
		}
		incomingData = hex.EncodeToString(buf)
		log.Println("[DEBUG] readSerialPort - Hex converted to string " + incomingData)
	}

	if n > 0 {
		//incomingData += string(buf[:n])
		incomingData = hex.EncodeToString(buf)
		log.Println("[DEBUG] readSerialPort - Hex converted to string 2 " + incomingData)
	}

	log.Println("[DEBUG] readSerialPort - Done reading")

	if err != nil {
		if err != io.EOF {
			log.Println("[ERROR] readSerialPort - Error Reading from serial port: " + err.Error())
		}
	}

	return incomingData, err
}

//WriteSerialPort :
func (serialDevice *SerialPort) WriteSerialPort(data string) error {
	//convert to Hex
	decoded, err := hex.DecodeString(data)
	if err != nil {
		log.Printf("[ERROR] WriteSerialPort - ERROR converting string to hex: %s\n", err.Error())
	} else {
		log.Printf("[DEBUG] WriteSerialPort - String converted to hex is: %x\n", decoded)
	}
	n, err := serialDevice.port.Write(decoded)
	//n, err := serialDevice.port.Write([]byte(decoded))
	if err != nil {
		log.Printf("[ERROR] WriteSerialPort - ERROR writing to serial port: %s\n", err.Error())
		return err
	} else {
		log.Printf("[DEBUG] SendATCommand - Number of bytes written: %d\n", n)
	}
	return nil
}

//FlushSerialPort :
func (serialDevice *SerialPort) FlushSerialPort() error {
	log.Println("[INFO] FlushSerialPort - resetting serial port")
	if err := serialDevice.port.Reset(); err != nil {
		log.Println("[ERROR] FlushSerialPort - Error resetting serial port: " + err.Error())
		return err
	}
	return nil
}

//SetDeadline :
func (serialDevice *SerialPort) SetDeadline(deadline int) error {
	return serialDevice.port.SetDeadline(time.Now().Add(time.Duration(deadline) * time.Millisecond))
}

//SetReadDeadline :
func (serialDevice *SerialPort) SetReadDeadline(deadline int) error {
	return serialDevice.port.SetReadDeadline(time.Now().Add(time.Duration(deadline) * time.Millisecond))
}

//SetWriteDeadline :
func (serialDevice *SerialPort) SetWriteDeadline(deadline int) error {
	return serialDevice.port.SetWriteDeadline(time.Now().Add(time.Duration(deadline) * time.Millisecond))
}
