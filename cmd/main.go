package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/simulator"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

const FILENAME = "oktopus-agent-sim"

func main() {
	done := make(chan os.Signal, 1)

	err := godotenv.Load()

	localEnv := ".env.local"
	if _, err := os.Stat(localEnv); err == nil {
		_ = godotenv.Overload(localEnv)
		slog.Info("Loaded variables from '.env.local'")
	} else {
		slog.Info("Loaded variables from '.env'")
	}

	if err != nil {
		slog.Warn("Error to load environment variables:", "error", err)
	}

	slog.SetDefault(
		slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: getLogLevelFromEnv(),
		})),
	)

	// Locks app running until it receives a stop command as Ctrl+C.
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	/*
		App variables priority:
		1º - Flag through command line.
		2º - Env variables.
		3º - Default flag value.
	*/

	slog.Info("Starting Oktopus TR-369 Agent Simulator")

	flSimNum := flag.Int("sim_number", utils.LookupEnvOrInt("SIM_NUM", 1), "Number of simulated devices")
	flNumToStartIds := flag.Int("num_to_start_ids", utils.LookupEnvOrInt("NUM_TO_START_IDS", 0), "From where to start your IDs")
	flMtp := flag.String("protocol", utils.LookupEnvOrString("MTP", ""), "MTP to use (mqtt, stomp, websockets)")
	flMqttAddr := flag.String("mqtt_addr", utils.LookupEnvOrString("MQTT_ADDR", "localhost"), "Address of the mqtt broker")
	flMqttPort := flag.String("mqtt_port", utils.LookupEnvOrString("MQTT_PORT", "1883"), "Port of the mqtt broker")
	flMqttUser := flag.String("mqtt_user", utils.LookupEnvOrString("MQTT_USER", ""), "Mqtt user")
	flMqttPasswd := flag.String("mqtt_passwd", utils.LookupEnvOrString("MQTT_PASSWD", ""), "Mqtt password")
	flMqttSsl := flag.Bool("mqtt_ssl", utils.LookupEnvOrBool("MQTT_SSL", false), "Mqtt with tls/ssl")
	flWsAddr := flag.String("ws_addr", utils.LookupEnvOrString("WS_ADDR", "localhost"), "Address of the websockets server")
	flWsPort := flag.String("ws_port", utils.LookupEnvOrString("WS_PORT", "8080"), "Port of the websockets server")
	flWsRoute := flag.String("ws_route", utils.LookupEnvOrString("WS_ROUTE", "/ws/agent"), "Route of the websockets server")
	flBridgeName := flag.String("br_name", utils.LookupEnvOrString("BR_NAME", "bridge"), "Bridge name of docker network")
	flWsSsl := flag.Bool("ws_ssl", utils.LookupEnvOrBool("WS_SSL", false), "Websockets with tls/ssl")
	flStompAddr := flag.String("stomp_addr", utils.LookupEnvOrString("STOMP_ADDR", "localhost"), "Address of the stomp broker")
	flStompPort := flag.String("stomp_port", utils.LookupEnvOrString("STOMP_PORT", "61613"), "Port of the stomp broker")
	flStompUser := flag.String("stomp_user", utils.LookupEnvOrString("STOMP_USER", ""), "Stomp user")
	flStompPasswd := flag.String("stomp_passwd", utils.LookupEnvOrString("STOMP_PASSWD", ""), "Stomp password")
	flStompSsl := flag.Bool("stomp_ssl", utils.LookupEnvOrBool("STOMP_SSL", false), "Stomp with tls/ssl")
	flBareMetal := flag.Bool("bare_metal", utils.LookupEnvOrBool("BARE_METAL", false), "Run simulator in bare metal")
	flEthernetInterface := flag.String("ethernet_interface", utils.LookupEnvOrString("ETH_INTERFACE", "eth0"), "Ethernet interface to use for obuspa connections")
	flExecutablePath := flag.String("executable_path", utils.LookupEnvOrString("EXECUTABLE_PATH", "/usr/local/bin/obuspa"), "Path to obuspa executable")
	flCleanDb := flag.Bool("clean_db", utils.LookupEnvOrBool("CLEAN_DB", false), "Clean obuspa database at the end of execution")
	flLogToStdout := flag.Bool("log_to_stdout", utils.LookupEnvOrBool("LOG_TO_STDOUT", false), "Log to stdout")
	flPath := flag.String("path", utils.LookupEnvOrString("PATH", ""), "Folder path to save configurations")
	flImgPath := flag.String("imgpath", utils.LookupEnvOrString("DOCKERFILE_PATH", ""), "Path to Dockerfile")
	flPrefix := flag.String("prefix", utils.LookupEnvOrString("PREFIX", "oktopus"), "Prefix of device id")
	flHelp := flag.Bool("help", false, "Help")

	flag.Parse()

	if *flHelp {
		flag.Usage()
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var cli *client.Client

	if !*flBareMetal {
		cli, err = container.CreateDockerClient()
		if err != nil {
			slog.Error("Error to create docker client:", "error", err)
			os.Exit(1)
		}
	} else {
		err = exec.Command("/usr/local/bin/obuspa", "-h").Run()
		if err != nil {
			slog.Error("Error to execute OBUSPA:", "error", err)
			os.Exit(1)
		}
	}

	conf := config.NewConfig(
		*flSimNum,
		*flNumToStartIds,
		*flPrefix,
		*flMtp,
		*flPath,
		/* ----------------------------- Docker Configs ----------------------------- */
		ctx,
		cli,
		*flImgPath,
		*flBridgeName,
		/* -------------------------------------------------------------------------- */

		/* ------------------------------ Mqtt Configs ------------------------------ */
		*flMqttUser,
		*flMqttPasswd,
		*flMqttSsl,
		*flMqttAddr,
		*flMqttPort,
		/* -------------------------------------------------------------------------- */

		/* ------------------------------ Websockets Configs ------------------------ */
		*flWsAddr,
		*flWsPort,
		*flWsRoute,
		*flWsSsl,
		/* -------------------------------------------------------------------------- */

		/* ------------------------------ Stomp Configs ----------------------------- */
		*flStompAddr,
		*flStompPort,
		*flStompUser,
		*flStompPasswd,
		*flStompSsl,
		/* -------------------------------------------------------------------------- */

		/* ------------------------------ Bare Metal Configs ------------------------ */
		*flBareMetal,
		*flEthernetInterface,
		*flExecutablePath,
		*flCleanDb,
		*flLogToStdout,
	)

	go simulator.StartDeviceSimulator(conf)

	<-done
	slog.Info("Received signal to stop the simulator")

	/* ----------------------------- Stop Gracefully ---------------------------- */
	cancel()
	slog.Info("Waiting for all agents to stop")
	conf.Wg.Wait()

	if !*flBareMetal {
		cli.Close()
	}
	/* -------------------------------------------------------------------------- */

	slog.Info("(⌐■_■) Agent simulator is out!")
}

func getLogLevelFromEnv() slog.Level {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL")) // Read and normalize

	switch levelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo // Default level
	}
}
