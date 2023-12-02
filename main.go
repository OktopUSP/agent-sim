package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/OktopUSP/agent-sim/internal/simulator"
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

	flSimNum := flag.Int("sim_number", lookupEnvOrInt("SIM_NUM", 1), "Number of simulated devices")
	flNumToStartIds := flag.Int("num_to_start_ids", lookupEnvOrInt("NUM_TO_START_IDS", 0), "From where to start your IDs")
	flMtp := flag.String("protocol", lookupEnvOrString("MTP", "mqtt"), "MTP to use (mqtt, stomp, websockets)")
	flPrefix := flag.String("prefix", lookupEnvOrString("PREFIX", "oktopus"), "Prefix of device id")
	flHelp := flag.Bool("help", false, "Help")

	var mtp simulator.MTP
	switch *flMtp {
	case "mqtt":
		mtp = simulator.Mqtt
	case "stomp":
		mtp = simulator.Stomp
	case "websockets":
		mtp = simulator.Websockets
	default:
		log.Printf("Incorrect MTP parameter: %s", *flMtp)
		os.Exit(0)
	}

	stopCounting := *flSimNum + *flNumToStartIds

	for i := *flNumToStartIds; i < stopCounting; i++ {
		go simulator.StartDeviceSimulator(mtp, i, *flPrefix)
	}

	flag.Parse()

	if *flHelp {
		flag.Usage()
		os.Exit(0)
	}

	<-done
	log.Println("(⌐■_■) Agent simulator is out!")
}

func lookupEnvOrString(key string, defaultVal string) string {
	if val, _ := os.LookupEnv(key); val != "" {
		return val
	}
	return defaultVal
}

func lookupEnvOrInt(key string, defaultVal int) int {
	if val, _ := os.LookupEnv(key); val != "" {
		v, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalf("LookupEnvOrInt[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}

func lookupEnvOrBool(key string, defaultVal bool) bool {
	if val, _ := os.LookupEnv(key); val != "" {
		v, err := strconv.ParseBool(val)
		if err != nil {
			log.Fatalf("LookupEnvOrBool[%s]: %v", key, err)
		}
		return v
	}
	return defaultVal
}
