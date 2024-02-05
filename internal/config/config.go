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
	Mqtt         Mqtt
	WebSockets   WebSockets
}

type Docker struct {
	Cli     *client.Client
	ImgPath string
}

type Mqtt struct {
	Addr string
	Port string
	User string
	Pass string
	Ssl  bool
}

type WebSockets struct {
	Addr  string
	Port  string
	Route string
	Ssl   bool
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
	mqttUser string,
	mqttPass string,
	mqttSsl bool,
	mqttAddr string,
	mqttPort string,
	wsAddr string,
	wsPort string,
	flWsRoute string,
	wsSsl bool,
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
		Mqtt: Mqtt{
			Addr: mqttAddr,
			Port: mqttPort,
			User: mqttUser,
			Pass: mqttPass,
			Ssl:  mqttSsl,
		},
		WebSockets: WebSockets{
			Addr:  wsAddr,
			Port:  wsPort,
			Route: flWsRoute,
			Ssl:   wsSsl,
		},
	}
}
