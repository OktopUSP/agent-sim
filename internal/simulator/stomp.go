package simulator

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/docker/docker/client"
)

type StompProtocol struct {
	Addr      string
	Port      string
	User      string
	Passwd    string
	Ssl       bool
	Wg        *sync.WaitGroup
	Ctx       context.Context
	Cli       *client.Client
	BareMetal config.BareMetal
}

func newStomp(c config.Config) StompProtocol {
	log.Println("Create new agent(s) with stomp protocol")
	log.Printf("Stomp client config: %++v", c.Stomp)

	return StompProtocol{
		/* ----------------------- Stomp connection parameters ----------------------- */
		Addr:   c.Stomp.Addr,
		Port:   c.Stomp.Port,
		User:   c.Stomp.User,
		Passwd: c.Stomp.Passwd,
		Ssl:    c.Stomp.Ssl,
		/* -------------------------------------------------------------------------- */
		Ctx:       c.Ctx,
		Wg:        c.Wg,
		Cli:       c.Docker.Cli,
		BareMetal: c.BareMetal,
	}
}

func (s *StompProtocol) startAgentBareMetal(id string, pre, dir string, logger io.Writer) {
	configFile := createStompFileConfig(id, pre, dir, *s)
	dbFile := dir + "/db-" + pre + "-" + id + ".db"

	args := []string{
		"-p",
		"-v", "4",
		"-r", configFile,
		"-f", dbFile,
		"-i", s.BareMetal.EthernetInterface,
	}

	if s.Ssl {
		sslFile := dir + "/chain.pem"
		args = append(args, "-t")
		args = append(args, sslFile)
	}

	cmd := exec.CommandContext(s.Ctx, s.BareMetal.ExecutablePath, args...)

	if s.BareMetal.LogToStdout {
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

	if s.BareMetal.CleanDb {
		err = os.Remove(dbFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	s.Wg.Done()
}

func (s *StompProtocol) startAgentDocker(id string, pre, br, dir string) {
	file := createStompFileConfig(id, pre, dir, *s)

	sslFile := ""
	if s.Ssl {
		if s.Ssl {
			sslFile = dir + "/chain.pem"
		}
		log.Println("SSL enabled")
		log.Println("SSL file path:", sslFile)
	}

	id, err := container.RunDockerContainer(
		s.Ctx,
		s.Cli,
		utils.DOCKER_IMG_NAME,
		br,
		pre+"-"+id+"-"+"stomp",
		file,
		sslFile,
	)

	if err != nil {
		log.Println(err)
	}

	<-s.Ctx.Done()

	err = container.DeleteDockerContainer(context.TODO(), s.Cli, id)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Deleted docker stomp container: %s", id)
	}

	s.Wg.Done()
}

func createStompFileConfig(id string, pre, dir string, s StompProtocol) string {
	err := os.WriteFile(
		dir+"/"+pre+"-"+id+"-stomp.txt",
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

Device.LocalAgent.EndpointID "`+pre+"-"+id+`-stomp"

#
# The following parameters will definitely need modifying
#
Device.LocalAgent.Controller.1.EndpointID "proto::oktopus"
Device.STOMP.Connection.1.Host "`+s.Addr+`"
Device.STOMP.Connection.1.Username "`+s.User+`"
Device.STOMP.Connection.1.Password "`+s.Passwd+`"

#
# The following parameters may be modified
#
Device.LocalAgent.MTP.1.Alias "`+pre+id+`"
Device.LocalAgent.MTP.1.Enable "true"
Device.LocalAgent.MTP.1.Protocol "STOMP"
Device.LocalAgent.MTP.1.STOMP.Reference "Device.STOMP.Connection.1"
Device.LocalAgent.MTP.1.STOMP.Destination "oktopus/usp/v1/agent"
Device.LocalAgent.Controller.1.Alias "cpe-1"
Device.LocalAgent.Controller.1.Enable "true"
Device.LocalAgent.Controller.1.AssignedRole "Device.LocalAgent.ControllerTrust.Role.1"
Device.LocalAgent.Controller.1.PeriodicNotifInterval "300"
Device.LocalAgent.Controller.1.PeriodicNotifTime "0001-01-01T00:00:00Z"
Device.LocalAgent.Controller.1.USPNotifRetryMinimumWaitInterval "5"
Device.LocalAgent.Controller.1.USPNotifRetryIntervalMultiplier "2000"
Device.LocalAgent.Controller.1.ControllerCode ""
Device.LocalAgent.Controller.1.MTP.1.Alias "`+pre+id+`"
Device.LocalAgent.Controller.1.MTP.1.Enable "true"
Device.LocalAgent.Controller.1.MTP.1.Protocol "STOMP"
Device.LocalAgent.Controller.1.MTP.1.STOMP.Reference "Device.STOMP.Connection.1"
Device.LocalAgent.Controller.1.MTP.1.STOMP.Destination "controller-notify-dest"
Device.STOMP.Connection.1.Alias "cpe-1"
Device.STOMP.Connection.1.Enable "true"
Device.STOMP.Connection.1.Port "`+s.Port+`"
Device.STOMP.Connection.1.EnableEncryption "false"
Device.STOMP.Connection.1.VirtualHost "/"
Device.STOMP.Connection.1.EnableHeartbeats "true"
Device.STOMP.Connection.1.OutgoingHeartbeat "30000"
Device.STOMP.Connection.1.IncomingHeartbeat "300000"
Device.STOMP.Connection.1.ServerRetryInitialInterval "60"
Device.STOMP.Connection.1.ServerRetryIntervalMultiplier "2000"
Device.STOMP.Connection.1.ServerRetryMaxInterval "30720"
Device.DeviceInfo.SerialNumber "`+pre+"-"+id+`"
Device.STOMP.Connection.1.EnableEncryption "`+strconv.FormatBool(s.Ssl)+`"
Internal.Reboot.Cause "LocalFactoryReset"
		`),
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}

	return dir + "/" + pre + "-" + id + "-stomp.txt"
}
