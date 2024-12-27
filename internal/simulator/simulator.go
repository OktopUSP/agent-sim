package simulator

import (
	"log"
	"os"
	"strconv"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
)

type agentSim interface {
	startAgentDocker(string, string, string, string)
	startAgentBareMetal(string, string, string, string)
}

type mtp int

const (
	Mqtt mtp = iota
	Stomp
	Websockets
)

const DEFAULT_DIR = "/configs"

func StartDeviceSimulator(c config.Config) {

	mtp := getMtp(c.Mtp)
	fileConfigDir := getDir(c.Path)

	agent_sim := getAgentSim(mtp, c)

	stopCounting := c.SimNumber + c.NumToStartId

	if c.BareMetal.Enable {

		for i := c.NumToStartId; i < stopCounting; i++ {
			c.Wg.Add(1)
			go agent_sim.startAgentBareMetal(strconv.Itoa(i), c.Prefix, c.BrName, fileConfigDir)
			// time.Sleep(time.Duration(100) * time.Millisecond)
		}

	} else {

		err := container.DownloadDockerImage(c.Ctx, c.Docker.Cli)
		if err != nil {
			log.Fatal(err)
		}

		for i := c.NumToStartId; i < stopCounting; i++ {
			c.Wg.Add(1)
			go agent_sim.startAgentDocker(strconv.Itoa(i), c.Prefix, c.BrName, fileConfigDir)
			// time.Sleep(time.Duration(100) * time.Millisecond)
		}

	}

}

func getAgentSim(mtp mtp, c config.Config) agentSim {
	switch mtp {
	case Mqtt:
		mqtt := newMqtt(c)
		return &mqtt
	case Stomp:
		stomp := newStomp(c)
		return &stomp
	case Websockets:
		ws := newWs(c)
		return &ws
	default:
		log.Fatal("Invalid MTP")
		return nil
	}
}

func getMtp(mtp_config string) mtp {

	var mtp mtp

	switch mtp_config {
	case "mqtt":
		mtp = Mqtt
	case "stomp":
		mtp = Stomp
	case "websockets":
		mtp = Websockets
	case "":
		log.Println("MTP not defined")
		os.Exit(1)
	default:
		log.Println("Invalid MTP")
		os.Exit(1)
	}

	return mtp
}

func getDir(path string) string {

	checkPathExists := func(dir string) {
		_, err := os.Stat(dir)
		if err != nil {
			log.Printf("Path: %s does not exist", path)
			os.Exit(1)
		}
	}

	if path == "" {
		path, _ = os.Getwd()
		path = path + DEFAULT_DIR
		log.Printf(
			"Path not defined, using current directory: %s",
			path,
		)
	}

	checkPathExists(path)
	return path
}
