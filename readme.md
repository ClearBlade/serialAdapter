# serial Adapter

The __serialAdapter__ provides the ability for the ClearBlade platform to communicate with a generic serial interface on a linux gateway.

The adapter subscribes to MQTT topics which are used to interact with the serial port. The adapter publishes any data retrieved from the serial port to MQTT topics so that the ClearBlade Platform is able to retrieve and process serial data or write data to serial devices.

# MQTT Topic Structure
The serialAdapter utilizes MQTT messaging to communicate with the ClearBlade Platform. The serialAdapter will subscribe to a specific topic in order to handle serial port requests. Additionally, the serialAdapter will publish messages to MQTT topics in order to communicate the results of requests to client applications. The topic structures utilized by the serialAdapter are as follows:

  * Read serial port data request: {__TOPIC ROOT__}/read
  * Write serial port data request: {__TOPIC ROOT__}/write
  * Read serial port data response: {__TOPIC ROOT__}/response
  * Serial port errors: {__TOPIC ROOT__}/error

## ClearBlade Platform Dependencies
The serialAdapter was constructed to provide the ability to communicate with a _System_ defined in a ClearBlade Platform instance. Therefore, the adapter requires a _System_ to have been created within a ClearBlade Platform instance.

Once a System has been created, artifacts must be defined within the ClearBlade Platform system to allow the adapter to function properly. At a minimum: 

  * A device needs to be created in the Auth --> Devices collection. The device will represent the adapter, or more importantly, the serial device or gateway on which the adapter is executing. The _name_ and _active key_ values specified in the Auth --> Devices collection will be used by the adapter to authenticate to the ClearBlade Platform or ClearBlade Edge. 
  * An adapter configuration data collection needs to be created in the ClearBlade Platform _system_ and populated with the data appropriate to the serial adapter. The schema of the data collection should be as follows:


| Column Name      | Column Datatype |
| ---------------- | --------------- |
| adapter_name     | string          |
| topic_root       | string          |
| adapter_settings | string (json)   |

### adapter_settings
The adapter_settings column will need to contain a JSON object containing the following optional attributes:

#### portName
* REQUIRED
* The operating system name for the device
* ex. /dev/ttymxc0

##### isHalfDuplex
* Most serial devices are full-duplex, ie. they can transmit and receive at the same time.
* Set to true if connecting to a device that operates half-duplex
* true/false - false is DEFAULT

##### bitRate
* number of bits per second (baud rate)

##### clearToSend (CTS)
* 0 --> CTS_INVALID - Special value to indicate setting should be left alone
* 1 --> CTS_IGNORE - CTS ignored
* 2 --> CTS_FLOW_CONTROL - CTS used for flow control

##### dataSetReady(DSR)
* 0 --> DSR_INVALID - Special value to indicate setting should be left alone
* 1 --> DSR_IGNORE - DSR ignored
* 2 --> DSR_FLOW_CONTROL - DSR used for flow control

##### dataTerminalReady (DTR)
* 0 --> DTR_INVALID - Special value to indicate setting should be left alone
* 1 --> DTR_OFF - DTR off
* 2 --> DTR_ON - DTR on
* 3 --> DTR_FLOW_CONTROL - DTR used for flow control

##### dataBits
* number of data bits
* 5, 6, 7, or 8
* DEFAULT = 8

##### deadline
* See go docs for net.Conn.SetDeadline
  * The read and write deadlines associated with the connection
  * Specified in milliseconds
  * A deadline is an absolute time after which I/O operations fail instead of blocking. The deadline applies to all future and pending I/O
  * A zero value for t means I/O operations will not time out.

##### flowControl
* 1 --> FLOWCONTROL_NONE - No flow control: DEFAULT
* 2 --> FLOWCONTROL_XONXOFF - Software flow control using XON/XOFF characters
* 3 --> FLOWCONTROL_RTSCTS - Hardware flow control using RTS/CTS signals
* 4 --> FLOWCONTROL_DTRDSR - Hardware flow control using DTR/DSR signals

##### parity
* 0 --> PARITY_INVALID - Special value to indicate setting should be left alone
* 1 --> PARITY_NONE - No parity: DEFAULT value
* 2 --> PARITY_ODD - Odd parity
* 3 --> PARITY_EVEN - Even parity
* 4 --> PARITY_MARK - Mark parity
* 5 --> PARITY_SPACE - Space parity

##### readDeadline
* See go docs for net.Conn.SetReadDeadline
  * The deadline for future Read calls and any currently-blocked Read call.
  * Specified in milliseconds
  * A zero value for t means Read will not time out.

##### requestToSend (RTS)
* 0 --> RTS_INVALID - Special value to indicate setting should be left alone
* 1 --> RTS_OFF - RTS off
* 2 --> RTS_ON - RTS on
* 3 --> RTS_FLOW_CONTROL - RTS used for flow control

##### stopBits
* Number of stop bits 
* 1 or 2
* DEFAULT = 1

##### writeDeadline
* See go docs for net.Conn.SetWriteDeadline
  * The deadline for future Write calls and any currently-blocked Write call.
  * Specified in milliseconds
  * Even if write times out, it may return n > 0, indicating that some of the data was successfully written.
  * A zero value for t means Write will not time out.

##### xONxOFF
* 0 --> XONXOFF_INVALID - Special value to indicate setting should be left alone
* 1 --> XONXOFF_DISABLED - XON/XOFF disabled
* 2 --> XONXOFF_IN -XON/XOFF enabled for input only
* 3 --> XONXOFF_OUT - XON/XOFF enabled for output only
* 4 --> XONXOFF_INOUT - XON/XOFF enabled for input and output

#### adapter_settings_example
{  
  "networkAddress":"00:11:22:33",  
  "networkDataKey":"33:22:11:00:33:22:11:00:33:22:11:00:33:22:11:00",   
  "networkSessionKey":"00:11:22:33:00:11:22:33:00:11:22:33:00:11:22:33",  
  "serialPortName":"/dev/ttyAP1",  
  "transmissionDataRate":"DR8",  
  "transmissionFrequency":"915500000"  
}


## Usage

### Executing the adapter

`serialAdapter -systemKey=<SYSTEM_KEY> -systemSecret=<SYSTEM_SECRET> -platformURL=<PLATFORM_URL> -messagingURL=<MESSAGING_URL> -deviceName=<DEVICE_NAME> -password=<DEVICE_ACTIVE_KEY> -adapterConfigCollection=<COLLECTION_NAME> -logLevel=<LOG_LEVEL>`

   __*Where*__ 

   __systemKey__
  * REQUIRED
  * The system key of the ClearBLade Platform __System__ the adapter will connect to

   __systemSecret__
  * REQUIRED
  * The system secret of the ClearBLade Platform __System__ the adapter will connect to
  
   __platformURL__
  * The url of the ClearBlade Platform instance the adapter will connect to
  * OPTIONAL
  * Defaults to __http://localhost:9000__

   __messagingURL__
  * The MQTT url of the ClearBlade Platform instance the adapter will connect to
  * OPTIONAL
  * Defaults to __localhost:1883__

   __deviceName__
  * REQUIRED
  * The device name the adapter will use to authenticate to the ClearBlade Platform
  * Requires the device to have been defined in the _Auth - Devices_ collection within the ClearBlade Platform __System__
   
   __password__
  * REQUIRED
  * The active key the adapter will use to authenticate to the platform
  * Requires the device to have been defined in the _Auth - Devices_ collection within the ClearBlade Platform __System__

   __adapterConfigCollection__
  * The collection name of the data collection used to house adapter configuration data
  * OPTIONAL
  * Defaults to __adapter_config__

   __logLevel__
  * The level of runtime logging the adapter should provide.
  * Available log levels:
    * fatal
    * error
    * warn
    * info
    * debug
  * OPTIONAL
  * Defaults to __info__


## Setup
---
The serial adapter is dependent upon the ClearBlade Go SDK and its dependent libraries being installed. The serial adapter was written in Go and therefore requires Go to be installed (https://golang.org/doc/install). The code utilizes features introduced in Go 1.13. Therefore, you MUST INSTALL Go v1.13 AT MINIMUM.


### Adapter compilation
The github.com/mikepb/go-serial library referenced in this adapter utilizes embedded C calls. Therefore, CGO needs to be enabled when compiling the adapter.



In order to compile the adapter for execution on ARM 5, the following steps need to be performed:

 1. Retrieve the adapter source code  
    * ```git clone git@github.com:ClearBlade/serialAdapter.git```
 2. Navigate to the xdotadapter directory  
    * ```cd serialAdapter```
 4. Compile the adapter
    * ```CGO_ENABLED=1 GOARCH=arm GOARM=5 GOOS=linux go build```



