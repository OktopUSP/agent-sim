package simulator

import (
	"log"
	"os"

	"github.com/OktopUSP/agent-sim/internal/config"
)

type agentSim interface {
	start(int, string, string)
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
	dir := getDir(c.Path)

	var agent_sim agentSim

	switch mtp {
	case Mqtt:
		mqtt := newMqtt(c)
		agent_sim = &mqtt
	case Stomp:
		log.Println("Stomp not implemented yet")
		os.Exit(0)
		//StartStompDevice(i, pre)
	case Websockets:
		log.Println("Websockets not implemented yet")
		os.Exit(0)
		//StartWebsocketsDevice(i, pre)
	}

	stopCounting := c.SimNumber + c.NumToStartId

	for i := c.NumToStartId; i < stopCounting; i++ {
		go agent_sim.start(i, c.Prefix, dir)
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
