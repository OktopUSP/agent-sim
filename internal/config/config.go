package config

import (
	"context"
	"flag"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	"github.com/OktopUSP/agent-sim/internal/container"
	"github.com/OktopUSP/agent-sim/internal/utils"
	"github.com/docker/docker/client"
	"github.com/joho/godotenv"
)

type Config struct {
	SimNumber     int
	NumToStartId  int
	Prefix        string
	Mtp           string
	Path          string
	Ctx           context.Context
	Wg            *sync.WaitGroup
	Docker        Docker
	BrName        string
	Mqtt          Mqtt
	WebSockets    WebSockets
	Stomp         Stomp
	BareMetal     BareMetal
	ProtoTrace    bool
	EnableMonitor bool
}

type BareMetal struct {
	Enable            bool
	ExecutablePath    string
	EthernetInterface string
	CleanDb           bool
	LogToStdout       bool
	Logger            io.Writer
	FileLogging       FileLogging
}

type FileLogging struct {
	Enable    bool
	LogFolder string
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

type Stomp struct {
	Addr   string
	Port   string
	User   string
	Passwd string
	Ssl    bool
}

func NewConfig(ctx context.Context) Config {

	localEnv := ".env.local"
	_, err := os.Stat(localEnv)
	if err == nil {
		_ = godotenv.Overload(localEnv)
	}

	/*
		App variables priority:
		1º - Flag through command line.
		2º - Env variables.
		3º - Default flag value.
	*/

	//TODO: find better names for env variables and flags
	//TODO: folders/paths/files consig to work in a similar way
	//TODO: dynamic configure obuspa log verbosity level

	mainProcessLogFile := flag.String("main_process_log_file", utils.LookupEnvOrString("MAIN_PROCESS_LOG_FILE", ""), "Main ulation process log file")
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
	flLogToStdout := flag.Bool("log_to_stdout", utils.LookupEnvOrBool("LOG_TO_STDOUT", false), "Enable log to stdout")
	flLogToFile := flag.Bool("log_to_file", utils.LookupEnvOrBool("LOG_TO_FILE", false), "Enable log to file")
	flLogFile := flag.String("log_file", utils.LookupEnvOrString("LOG_FOLDER_PATH", ""), "Log folder path")
	flLogLevel := flag.String("log_level", utils.LookupEnvOrString("LOG_LEVEL", "info"), "Log level")
	flPath := flag.String("path_cfg", utils.LookupEnvOrString("PATH_CFG", ""), "Folder path to save configurations")
	flImgPath := flag.String("imgpath", utils.LookupEnvOrString("DOCKERFILE_PATH", ""), "Path to Dockerfile")
	flProtoTrace := flag.Bool("proto_trace", utils.LookupEnvOrBool("PROTO_TRACE", false), "Enable OBUSPA protobuffer tracing")
	flPrefix := flag.String("prefix", utils.LookupEnvOrString("PREFIX", "oktopus"), "Prefix of device id")
	flEnableMonitor := flag.Bool("enable_monitor", utils.LookupEnvOrBool("ENABLE_MONITOR", false), "Enable process monitoring")
	flHelp := flag.Bool("help", false, "Help")

	setupLogging(*flLogLevel, *mainProcessLogFile)

	flag.Parse()

	if *flHelp {
		flag.Usage()
		os.Exit(0)
	}

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

	return Config{
		SimNumber:     *flSimNum,
		NumToStartId:  *flNumToStartIds,
		Prefix:        *flPrefix,
		Mtp:           *flMtp,
		Path:          *flPath,
		Wg:            &sync.WaitGroup{},
		ProtoTrace:    *flProtoTrace,
		Ctx:           ctx,
		EnableMonitor: *flEnableMonitor,
		Docker: Docker{
			Cli:     cli,
			ImgPath: *flImgPath,
		},
		BrName: *flBridgeName,
		Mqtt: Mqtt{
			Addr: *flMqttAddr,
			Port: *flMqttPort,
			User: *flMqttUser,
			Pass: *flMqttPasswd,
			Ssl:  *flMqttSsl,
		},
		WebSockets: WebSockets{
			Addr:  *flWsAddr,
			Port:  *flWsPort,
			Route: *flWsRoute,
			Ssl:   *flWsSsl,
		},
		Stomp: Stomp{
			Addr:   *flStompAddr,
			Port:   *flStompPort,
			User:   *flStompUser,
			Passwd: *flStompPasswd,
			Ssl:    *flStompSsl,
		},
		BareMetal: BareMetal{
			Enable:            *flBareMetal,
			EthernetInterface: *flEthernetInterface,
			ExecutablePath:    *flExecutablePath,
			CleanDb:           *flCleanDb,
			LogToStdout:       *flLogToStdout,
			FileLogging: FileLogging{
				Enable:    *flLogToFile,
				LogFolder: *flLogFile,
			},
		},
	}
}

func getLogLevel(leveStr string) slog.Level {

	switch leveStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func setupLogging(logLevel string, logFile string) {
	var logWriter io.Writer = os.Stdout
	if logFile != "" {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Error("Failed to open log file:", "error", err)
			os.Exit(1)
		}
		logWriter = io.MultiWriter(os.Stdout, file)
	}
	slog.SetDefault(
		slog.New(slog.NewJSONHandler(logWriter, &slog.HandlerOptions{
			Level: getLogLevel(logLevel),
		})),
	)
}
