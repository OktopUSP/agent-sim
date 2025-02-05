package config

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
)

type Config struct {
	SimNumber    int
	NumToStartId int
	Prefix       string
	Mtp          string
	Path         string
	Ctx          context.Context
	Wg           *sync.WaitGroup
	Docker       Docker
	BrName       string
	Mqtt         Mqtt
	WebSockets   WebSockets
	Stomp        Stomp
	BareMetal    BareMetal
}

type BareMetal struct {
	Enable            bool
	ExecutablePath    string
	EthernetInterface string
	CleanDb           bool
	LogToStdout       bool
}

type Docker struct {
	Cli     *client.Client
	ImgPath string
}

type Mqtt struct {
	Addr  string
	Addr2 string
	Port  string
	User  string
	Pass  string
	Ssl   bool
}

type WebSockets struct {
	Addr  string
	Port  string
	Route string
	Ssl   bool
}

type Stomp struct {
	Addr   string
	Addr2  string
	Port   string
	User   string
	Passwd string
	Ssl    bool
}

func NewConfig(
	simNumber int,
	numToStartId int,
	prefix string,
	mtp string,
	path string,
	ctx context.Context,
	dockerCli *client.Client,
	dockerImgPath string,
	flBridgeName string,
	mqttUser string,
	mqttPass string,
	mqttSsl bool,
	mqttAddr string,
	mqttAddr2 string,
	mqttPort string,
	wsAddr string,
	wsPort string,
	flWsRoute string,
	wsSsl bool,
	stompAddr string,
	stompAddr2 string,
	stompPort string,
	stompUser string,
	stompPasswd string,
	stompSsl bool,
	bareMetal bool,
	ethernetInterface string,
	executablePath string,
	cleanDb bool,
	logToStdout bool,
) Config {
	return Config{
		SimNumber:    simNumber,
		NumToStartId: numToStartId,
		Prefix:       prefix,
		Mtp:          mtp,
		Path:         path,
		Wg:           &sync.WaitGroup{},
		Ctx:          ctx,
		Docker: Docker{
			Cli:     dockerCli,
			ImgPath: dockerImgPath,
		},
		BrName: flBridgeName,
		Mqtt: Mqtt{
			Addr:  mqttAddr,
			Addr2: mqttAddr2,
			Port:  mqttPort,
			User:  mqttUser,
			Pass:  mqttPass,
			Ssl:   mqttSsl,
		},
		WebSockets: WebSockets{
			Addr:  wsAddr,
			Port:  wsPort,
			Route: flWsRoute,
			Ssl:   wsSsl,
		},
		Stomp: Stomp{
			Addr:   stompAddr,
			Addr2:  stompAddr2,
			Port:   stompPort,
			User:   stompUser,
			Passwd: stompPasswd,
			Ssl:    stompSsl,
		},
		BareMetal: BareMetal{
			Enable:            bareMetal,
			EthernetInterface: ethernetInterface,
			ExecutablePath:    executablePath,
			CleanDb:           cleanDb,
			LogToStdout:       logToStdout,
		},
	}
}
