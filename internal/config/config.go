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
	Address      string
	Port         string
	Mtp          string
	Path         string
	Ctx          context.Context
	Wg           *sync.WaitGroup
	Docker       Docker
}

type Docker struct {
	Cli     *client.Client
	ImgPath string
}

func NewConfig(
	simNumber int,
	numToStartId int,
	prefix string,
	addr string,
	port string,
	mtp string,
	path string,
	ctx context.Context,
	dockerCli *client.Client,
	dockerImgPath string,
) Config {
	return Config{
		SimNumber:    simNumber,
		NumToStartId: numToStartId,
		Prefix:       prefix,
		Address:      addr,
		Port:         port,
		Mtp:          mtp,
		Path:         path,
		Wg:           &sync.WaitGroup{},
		Ctx:          ctx,
		Docker: Docker{
			Cli:     dockerCli,
			ImgPath: dockerImgPath,
		},
	}
}
