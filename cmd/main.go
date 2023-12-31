package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/simulator"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/joho/godotenv"
)

const FILENAME = "oktopus-agent-sim"
const VERSION = "0.0.1"

func main() {
	done := make(chan os.Signal, 1)

	err := godotenv.Load()

	localEnv := ".env.local"
	if _, err := os.Stat(localEnv); err == nil {
		_ = godotenv.Overload(localEnv)
		log.Println("Loaded variables from '.env.local'")
	} else {
		log.Println("Loaded variables from '.env'")
	}

	if err != nil {
		log.Println("Error to load environment variables:", err)
	}

	// Locks app running until it receives a stop command as Ctrl+C.
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	/*
		App variables priority:
		1º - Flag through command line.
		2º - Env variables.
		3º - Default flag value.
	*/

	log.Println("Starting Oktopus TR-369 Agent Simulator Version:", VERSION)

	flSimNum := flag.Int("sim_number", utils.LookupEnvOrInt("SIM_NUM", 1), "Number of simulated devices")
	flNumToStartIds := flag.Int("num_to_start_ids", utils.LookupEnvOrInt("NUM_TO_START_IDS", 0), "From where to start your IDs")
	flMtp := flag.String("protocol", utils.LookupEnvOrString("MTP", ""), "MTP to use (mqtt, stomp, websockets)")
	flAddr := flag.String("addr", utils.LookupEnvOrString("ADDR", "localhost"), "Address of the broker/server")
	flPort := flag.String("port", utils.LookupEnvOrString("PORT", "1883"), "Port of the broker/server")
	flPath := flag.String("path", utils.LookupEnvOrString("PATH", ""), "Folder path to save configurations")
	flPrefix := flag.String("prefix", utils.LookupEnvOrString("PREFIX", "oktopus"), "Prefix of device id")
	flHelp := flag.Bool("help", false, "Help")

	flag.Parse()

	if *flHelp {
		flag.Usage()
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())

	dockerCli, err := container.CreateDockerClient()
	if err != nil {
		log.Fatal(err)
	}

	conf := config.NewConfig(
		*flSimNum,
		*flNumToStartIds,
		*flPrefix,
		*flAddr,
		*flPort,
		*flMtp,
		*flPath,
		ctx,
	)

	simulator.StartDeviceSimulator(conf)

	<-done
	cancel()

	log.Println("(⌐■_■) Agent simulator is out!")
}
