package simulator

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/docker/docker/client"
)

type WsProtocol struct {
	Addr      string
	Port      string
	Route     string
	Ssl       bool
	Wg        *sync.WaitGroup
	Ctx       context.Context
	Cli       *client.Client
	BareMetal config.BareMetal
}

func newWs(c config.Config) WsProtocol {

	log.Println("Create new agent(s) with websockets protocol")
	log.Printf("Websockets client config: %++v", c.WebSockets)

	return WsProtocol{
		/* ----------------------- Websockets connection parameters ----------------------- */
		Addr:  c.WebSockets.Addr,
		Port:  c.WebSockets.Port,
		Route: c.WebSockets.Route,
		Ssl:   c.WebSockets.Ssl,
		/* -------------------------------------------------------------------------- */
		Ctx:       c.Ctx,
		Wg:        c.Wg,
		Cli:       c.Docker.Cli,
		BareMetal: c.BareMetal,
	}
}

func (w *WsProtocol) startAgentBareMetal(id, pre, dir string, logger io.Writer) {
	configFile := createWsFileConfig(id, pre, dir, *w)
	dbFile := dir + "/db-" + pre + "-" + id + ".db"

	args := []string{
		"-p",
		"-v", "4",
		"-r", configFile,
		"-f", dbFile,
		"-i", w.BareMetal.EthernetInterface,
	}

	if w.Ssl {
		sslFile := dir + "/chain.pem"
		args = append(args, "-t")
		args = append(args, sslFile)
	}

	cmd := exec.CommandContext(w.Ctx, w.BareMetal.ExecutablePath, args...)

	if w.BareMetal.LogToStdout {
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

	if w.BareMetal.CleanDb {
		err = os.Remove(dbFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	w.Wg.Done()
}

func (w *WsProtocol) startAgentDocker(id, pre, br, dir string) {
	file := createWsFileConfig(id, pre, dir, *w)
	id, err := container.RunDockerContainer(
		w.Ctx,
		w.Cli,
		utils.DOCKER_IMG_NAME,
		br,
		pre+"-"+id+"-"+"websockets",
		file,
		"",
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

func createWsFileConfig(id string, pre, dir string, w WsProtocol) string {
	//TODO: create ssl agent option
	err := os.WriteFile(
		dir+"/"+pre+"-"+id+"-websockets.txt",
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

Device.LocalAgent.EndpointID "`+pre+"-"+id+`-ws"

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
Device.LocalAgent.Controller.1.MTP.1.Alias "`+pre+id+`"
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

	return dir + "/" + pre + "-" + id + "-websockets.txt"
}
