package simulator

import (
	"context"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/docker/docker/client"
)

type MqttProtocol struct {
	Addr      string
	Addr2     string
	Port      string
	User      string
	Pass      string
	Ssl       bool
	Wg        *sync.WaitGroup
	Ctx       context.Context
	Cli       *client.Client
	BareMetal config.BareMetal
}

func newMqtt(c config.Config) MqttProtocol {

	log.Println("Create new agent(s) with mqtt protocol")
	log.Printf("Mqtt client config: %++v", c.Mqtt)

	return MqttProtocol{
		/* ----------------------- Mqtt connection parameters ----------------------- */
		Addr:  c.Mqtt.Addr,
		Addr2: c.Mqtt.Addr2,
		Port:  c.Mqtt.Port,
		User:  c.Mqtt.User,
		Pass:  c.Mqtt.Pass,
		Ssl:   c.Mqtt.Ssl,
		/* -------------------------------------------------------------------------- */
		Ctx:       c.Ctx,
		Wg:        c.Wg,
		Cli:       c.Docker.Cli,
		BareMetal: c.BareMetal,
	}
}

func (m *MqttProtocol) startAgentBareMetal(id, pre, br, dir string) {
	configFile := createMqttFileConfig(id, pre, dir, *m)
	dbFile := dir + "/db-" + pre + "-" + id + ".db"

	args := []string{
		"-p",
		"-v", "4",
		"-r", configFile,
		"-f", dbFile,
		"-i", m.BareMetal.EthernetInterface,
	}

	if m.Ssl {
		sslFile := dir + "/chain.pem"
		args = append(args, "-t")
		args = append(args, sslFile)
	}

	cmd := exec.CommandContext(m.Ctx, m.BareMetal.ExecutablePath, args...)

	if m.BareMetal.LogToStdout {
		cmd.Stdout = os.Stdout
	}

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Println(err)
	}

	if m.BareMetal.CleanDb {
		err = os.Remove(dbFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	m.Wg.Done()
}

func (m *MqttProtocol) startAgentDocker(id string, pre string, br string, dir string) {

	file := createMqttFileConfig(id, pre, dir, *m)

	id, err := container.RunDockerContainer(
		m.Ctx,
		m.Cli,
		utils.DOCKER_IMG_NAME,
		br,
		pre+"-"+id+"-"+"mqtt",
		file,
		"",
	)

	if err != nil {
		log.Println(err)
	}

	<-m.Ctx.Done()

	err = container.DeleteDockerContainer(context.TODO(), m.Cli, id)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Deleted docker mqtt container: %s", id)
	}

	m.Wg.Done()
}

func createMqttFileConfig(id string, pre, dir string, m MqttProtocol) string {
	//TODO: create ssl agent option
	//TODO: create mqqt client version
	err := os.WriteFile(
		dir+"/"+pre+"-"+id+"-mqtt.txt",
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

Device.LocalAgent.EndpointID "`+pre+"-"+id+`-mqtt"


## Adding boot params
Device.LocalAgent.Controller.1.BootParameter.1.Enable true
Device.LocalAgent.Controller.1.BootParameter.1.ParameterName "Device.LocalAgent.EndpointID"

Device.LocalAgent.MTP.1.MQTT.Reference "Device.MQTT.Client.1"
Device.LocalAgent.MTP.2.MQTT.ResponseTopicConfigured "oktopus/v1/controller"
Device.LocalAgent.MTP.2.MQTT.Reference "Device.MQTT.Client.2"
Device.MQTT.Client.1.RequestResponseInfo true
Device.MQTT.Client.1.BrokerAddress "`+m.Addr+`"
Device.MQTT.Client.1.ProtocolVersion "5.0"
Device.MQTT.Client.1.BrokerPort "`+m.Port+`"
Device.MQTT.Client.1.TransportProtocol "`+isTLS(m.Ssl)+`"
Device.MQTT.Client.1.Username "`+m.User+`"
Device.MQTT.Client.1.Password "`+m.Pass+`"
Device.MQTT.Client.1.Alias "cpe-1"
Device.MQTT.Client.1.Enable true
Device.MQTT.Client.1.ClientID ""
Device.MQTT.Client.1.KeepAliveTime "60"

Device.MQTT.Client.2.BrokerAddress "`+m.Addr2+`"
Device.MQTT.Client.2.ProtocolVersion "5.0"
Device.MQTT.Client.2.BrokerPort "`+m.Port+`"
Device.MQTT.Client.2.TransportProtocol "TCP/IP"
Device.MQTT.Client.2.Username "`+m.User+`"
Device.MQTT.Client.2.Password "`+m.Pass+`"
Device.MQTT.Client.2.Alias "cpe-1"
Device.MQTT.Client.2.Enable true
Device.MQTT.Client.2.ClientID ""
Device.MQTT.Client.2.KeepAliveTime "60"

Device.MQTT.Client.1.ConnectRetryTime "5"
Device.MQTT.Client.1.ConnectRetryIntervalMultiplier   "2000"
Device.MQTT.Client.1.ConnectRetryMaxInterval "60"

Device.MQTT.Client.2.ConnectRetryTime "5"
Device.MQTT.Client.2.ConnectRetryIntervalMultiplier   "2000"
Device.MQTT.Client.2.ConnectRetryMaxInterval "60"


Device.LocalAgent.Controller.1.Alias "cpe-1"
Device.LocalAgent.Controller.1.Enable true
Device.LocalAgent.Controller.1.PeriodicNotifInterval "86400"
Device.LocalAgent.Controller.1.PeriodicNotifTime "0001-01-01T00:00:00Z"
Device.LocalAgent.Controller.1.ControllerCode ""
Device.LocalAgent.Controller.1.MTP.1.Alias "`+pre+id+`"
Device.LocalAgent.Controller.1.MTP.1.Enable true
Device.LocalAgent.Controller.1.MTP.1.Protocol "MQTT"
Device.LocalAgent.Controller.1.EndpointID "proto::oktopus"
Device.LocalAgent.Controller.1.MTP.1.MQTT.Reference "Device.MQTT.Client.1"
Device.LocalAgent.Controller.1.MTP.1.MQTT.Topic "oktopus/v1/controller"
Device.LocalAgent.Controller.1.MTP.2.MQTT.Reference "Device.MQTT.Client.2"
Device.LocalAgent.Controller.1.MTP.2.MQTT.Topic "oktopus/v1/controller"


#
# The following parameters may be modified
#
Device.LocalAgent.MTP.1.Alias "`+pre+id+`"
Device.LocalAgent.MTP.1.Enable true
Device.LocalAgent.MTP.1.Protocol "MQTT"
Device.LocalAgent.MTP.2.Alias "`+pre+id+`"
Device.LocalAgent.MTP.2.Enable true
Device.LocalAgent.MTP.2.Protocol "MQTT"
Device.DeviceInfo.SerialNumber "`+pre+"-"+id+`"

Internal.Reboot.Cause "LocalFactoryReset"
		`),
		0644,
	)
	if err != nil {
		log.Fatal("Error to create config file: ", err)
	}

	return dir + "/" + pre + "-" + id + "-mqtt.txt"
}

func isTLS(isTLS bool) string {
	if isTLS {
		return "TLS"
	}
	return "TCP/IP"
}

func (m *MqttProtocol) startMqttAgent(file, pre, br, id string) {
	id, err := container.RunDockerContainer(
		m.Ctx,
		m.Cli,
		utils.DOCKER_IMG_NAME,
		br,
		pre+"-"+id+"-"+"mqtt",
		file,
		"",
	)

	if err != nil {
		log.Println(err)
	}

	<-m.Ctx.Done()

	err = container.DeleteDockerContainer(context.TODO(), m.Cli, id)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Deleted docker mqtt container: %s", id)
	}

	m.Wg.Done()
}
