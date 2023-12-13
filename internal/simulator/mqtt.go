package simulator

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/OktopUSP/agent-sim/internal/config"
)

type MqttProtocol struct {
	Addr string
	Port string
	Ssl  bool
}

func newMqtt(c config.Config) MqttProtocol {
	return MqttProtocol{
		Addr: c.Address,
		Port: c.Port,
	}
}

func (m *MqttProtocol) start(id int, pre string, dir string) {
	log.Printf("Device: %s-%v", pre, id)
	file := createMqttFileConfig(id, pre, dir, m.Port, m.Addr)
	startMqttAgent(file, pre, strconv.Itoa(id))
}

func createMqttFileConfig(id int, pre string, dir, port, addr string) string {
	err := os.WriteFile(
		dir+"/"+pre+"-"+strconv.Itoa(id)+"-mqtt.txt",
		[]byte(`
#
# This file contains a factory reset database in text format
#
# If no USP database exists when OB-USP-AGENT starts, then OB-USP-AGENT will create a database containing
# the parameters specified in a text file located by the '-r' option.
# Example:
#    obuspa -p -v 4 -r factory_reset_example.txt
#
# Each line of this file contains either a comment (denoted by '#' at the start of the line)
# or a USP data model parameter and its factory reset value.
# The parameter and value are separated by whitespace.
# The value may optionally be enclosed in speech marks "" (this is the only way to specify an empty string)
#
##########################################################################################################
#
# Adding MQTT parameters to test the datamodel interface
#

Device.LocalAgent.EndpointID "`+pre+"-"+strconv.Itoa(id)+`-mqtt"


## Adding boot params
Device.LocalAgent.Controller.1.BootParameter.1.Enable true
Device.LocalAgent.Controller.1.BootParameter.1.ParameterName "Device.LocalAgent.EndpointID"
Device.LocalAgent.Subscription.1.Alias cpe-1
Device.LocalAgent.Subscription.1.Enable true
Device.LocalAgent.Subscription.1.ID default-boot-event-ACS
Device.LocalAgent.Subscription.1.Recipient Device.LocalAgent.Controller.1
Device.LocalAgent.Subscription.1.NotifType Event
Device.LocalAgent.Subscription.1.ReferenceList Device.Boot!
Device.LocalAgent.Subscription.1.Persistent true

Device.LocalAgent.MTP.1.MQTT.ResponseTopicConfigured "oktopus/v1/controller"
Device.LocalAgent.MTP.1.MQTT.Reference "Device.MQTT.Client.1"
Device.MQTT.Client.1.BrokerAddress "`+addr+`"
Device.MQTT.Client.1.ProtocolVersion "5.0"
Device.MQTT.Client.1.BrokerPort "`+port+`"
Device.MQTT.Client.1.TransportProtocol "TCP/IP"
Device.MQTT.Client.1.Username ""
Device.MQTT.Client.1.Password ""
Device.MQTT.Client.1.Alias "cpe-1"
Device.MQTT.Client.1.Enable true
Device.MQTT.Client.1.ClientID ""
Device.MQTT.Client.1.KeepAliveTime "60"

Device.MQTT.Client.1.ConnectRetryTime "5"
Device.MQTT.Client.1.ConnectRetryIntervalMultiplier   "2000"
Device.MQTT.Client.1.ConnectRetryMaxInterval "60"


Device.LocalAgent.Controller.1.Alias "cpe-1"
Device.LocalAgent.Controller.1.Enable true
Device.LocalAgent.Controller.1.PeriodicNotifInterval "86400"
Device.LocalAgent.Controller.1.PeriodicNotifTime "0001-01-01T00:00:00Z"
Device.LocalAgent.Controller.1.ControllerCode ""
Device.LocalAgent.Controller.1.MTP.1.Alias "`+pre+strconv.Itoa(id)+`"
Device.LocalAgent.Controller.1.MTP.1.Enable true
Device.LocalAgent.Controller.1.MTP.1.Protocol "MQTT"
Device.LocalAgent.Controller.1.EndpointID "oktopusController"
Device.LocalAgent.Controller.1.MTP.1.MQTT.Reference "Device.MQTT.Client.1"
Device.LocalAgent.Controller.1.MTP.1.MQTT.Topic "oktopus/v1/controller"



#
# The following parameters may be modified
#
Device.LocalAgent.MTP.1.Alias "`+pre+strconv.Itoa(id)+`"
Device.LocalAgent.MTP.1.Enable true
Device.LocalAgent.MTP.1.Protocol "MQTT"
Device.DeviceInfo.SerialNumber "`+pre+"-"+strconv.Itoa(id)+`"

Internal.Reboot.Cause "LocalFactoryReset"
		`),
		0644,
	)
	if err != nil {
		log.Fatal("Error to create config file: ", err)
	}

	return dir + "/" + pre + "-" + strconv.Itoa(id) + "-mqtt.txt"
}

func startMqttAgent(file, pre, id string) {

	//TODO: run all agents in a container

	cmd := exec.Command("/usr/local/bin/obuspa", "-p", "-v", "4", "-i", "lo", "-r", file)

	// Create pipes to capture command output
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err)
		return
	}

	// Create a scanner for command output
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	// Use goroutines to read and print output in real-time
	go printOutput(stdoutScanner, pre+"-"+id+"-"+"mqtt")
	go printOutput(stderrScanner, pre+"-"+id+"-"+"mqtt")

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for command:", err)
		return
	}
}

// printOutput reads lines from the scanner and prints them with a prefix
func printOutput(scanner *bufio.Scanner, prefix string) {
	for scanner.Scan() {
		fmt.Printf("[%s] %s\n", prefix, scanner.Text())
	}
}
