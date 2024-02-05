package simulator

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/docker/docker/client"
)

type WsProtocol struct {
	Addr  string
	Port  string
	Route string
	Ssl   bool
	Wg    *sync.WaitGroup
	Ctx   context.Context
	Cli   *client.Client
}

func newWs(c config.Config) WsProtocol {

	log.Println("Create new agent(s) with websockets protocol")
	log.Printf("Websockets client config: %++v", c.Mqtt)

	return WsProtocol{
		/* ----------------------- Websockets connection parameters ----------------------- */
		Addr:  c.WebSockets.Addr,
		Port:  c.WebSockets.Port,
		Route: c.WebSockets.Route,
		Ssl:   c.WebSockets.Ssl,
		/* -------------------------------------------------------------------------- */
		Ctx: c.Ctx,
		Wg:  c.Wg,
		Cli: c.Docker.Cli,
	}
}

func (w *WsProtocol) start(id int, pre string, dir string) {
	log.Printf("Device: %s-%v", pre, id)
	file := createWsFileConfig(id, pre, dir, *w)
	w.startWsAgent(file, pre, strconv.Itoa(id))
}

func createWsFileConfig(id int, pre, dir string, w WsProtocol) string {
	//TODO: create ssl agent option
	err := os.WriteFile(
		dir+"/"+pre+"-"+strconv.Itoa(id)+"-websockets.txt",
		[]byte(`
##########################################################################################################
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
# The following parameters will definitely need modifying
#

Device.LocalAgent.EndpointID "`+pre+"-"+strconv.Itoa(id)+`-ws"

# Controller's websocket server (for agent initiated sessions)
Device.LocalAgent.Controller.1.EndpointID "oktopusController"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.Host "`+w.Addr+`"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.Port "`+w.Port+`"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.Path "`+w.Route+`"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.EnableEncryption "false"

# Agent's websocket server (for controller initiated sessions)
Device.LocalAgent.MTP.1.WebSocket.Port "8080"
Device.LocalAgent.MTP.1.WebSocket.Path "/usp"
Device.LocalAgent.MTP.1.WebSocket.EnableEncryption "false"


#
# The following parameters may be modified
#
Device.LocalAgent.MTP.1.Alias "cpe-1"
Device.LocalAgent.MTP.1.Enable "true"
Device.LocalAgent.MTP.1.Protocol "WebSocket"
Device.LocalAgent.MTP.1.WebSocket.KeepAliveInterval "30"
Device.LocalAgent.Controller.1.Alias "cpe-1"
Device.LocalAgent.Controller.1.Enable "true"
Device.LocalAgent.Controller.1.AssignedRole "Device.LocalAgent.ControllerTrust.Role.1"
Device.LocalAgent.Controller.1.PeriodicNotifInterval "86400"
Device.LocalAgent.Controller.1.PeriodicNotifTime "0001-01-01T00:00:00Z"
Device.LocalAgent.Controller.1.USPNotifRetryMinimumWaitInterval "5"
Device.LocalAgent.Controller.1.USPNotifRetryIntervalMultiplier "2000"
Device.LocalAgent.Controller.1.ControllerCode ""
Device.LocalAgent.Controller.1.MTP.1.Alias "`+pre+strconv.Itoa(id)+`"
Device.LocalAgent.Controller.1.MTP.1.Enable "true"
Device.LocalAgent.Controller.1.MTP.1.Protocol "WebSocket"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.KeepAliveInterval "30"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.SessionRetryMinimumWaitInterval "5"
Device.LocalAgent.Controller.1.MTP.1.WebSocket.SessionRetryIntervalMultiplier "2000"
Internal.Reboot.Cause "LocalFactoryReset"
		`),
		0644,
	)
	if err != nil {
		log.Fatal("Error to create config file: ", err)
	}

	return dir + "/" + pre + "-" + strconv.Itoa(id) + "-websockets.txt"
}

func (w *WsProtocol) startWsAgent(file, pre, id string) {
	id, err := container.RunDockerContainer(
		w.Ctx,
		w.Cli,
		utils.DOCKER_IMG_NAME,
		pre+"-"+id+"-"+"websockets",
		file,
	)

	if err != nil {
		log.Println(err)
	}

	<-w.Ctx.Done()

	err = container.DeleteDockerContainer(context.TODO(), w.Cli, id)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Deleted docker websockets container: %s", id)
	}

	w.Wg.Done()
}
