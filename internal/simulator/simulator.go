package simulator

import (
	"io"
	"log"
	"log/slog"
	"os"
	"strconv"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/container"
)

type agentSim interface {
	startAgentDocker(string, string, string, string)
	startAgentBareMetal(string, string, string, io.Writer, chan int)
}

type mtp int

const (
	Mqtt mtp = iota
	Stomp
	Websockets
)

const DEFAULT_DIR = "/configs"

func StartDeviceSimulator(c config.Config) {

	/* ------------------------------ config parser ----------------------------- */
	mtp := getMtp(c.Mtp)                                                  // Parse the MTP to be used
	fileConfigDir := getConfigDir(c.Path)                                 // Defines where the config files for OBUSPA are
	logger := getLogger(c.BareMetal.LogToStdout, c.BareMetal.FileLogging) // Defines where the log goes to
	agent_sim := getAgentSim(mtp, c)                                      // Creates the Agent Simulator interface
	/* -------------------------------------------------------------------------- */

	stopCounting := c.SimNumber + c.NumToStartId

	pid := make(chan int, c.SimNumber)

	if c.BareMetal.Enable {
		slog.Info("Starting bare metal agent(s)", "number", c.SimNumber)
		for i := c.NumToStartId; i < stopCounting; i++ {
			c.Wg.Add(1)
			go agent_sim.startAgentBareMetal(strconv.Itoa(i), c.Prefix, fileConfigDir, logger, pid)
		}
		slog.Info("Bare metal agent(s) started")

		if c.EnableMonitor {
			pids := make([]int, c.SimNumber)
			for i := 0; i < c.SimNumber; i++ {
				pids[i] = <-pid
			}
			slog.Debug("Process monitor list", "pids", pids)
			monitorProcesses(pids)
		}

	} else {

		err := container.DownloadDockerImage(c.Ctx, c.Docker.Cli)
		if err != nil {
			log.Fatal(err)
		}

		if c.BrName == "" {
			log.Println("Bridge name not defined")
			c.BrName = "br"
		}

		j := 0
		br := 0

		for i := c.NumToStartId; i < stopCounting; i++ {
			c.Wg.Add(1)
			j++
			if j > 1000 {
				br++
				c.BrName = c.BrName + strconv.Itoa(br)
				j = 0
			}
			go agent_sim.startAgentDocker(strconv.Itoa(i), c.Prefix, c.BrName, fileConfigDir)
		}

	}

}

func getAgentSim(mtp mtp, c config.Config) agentSim {
	switch mtp {
	case Mqtt:
		mqtt := newMqtt(c)
		return &mqtt
	// TODO: Compatibilize STOMP and websockets with the new agent-sim
	case Stomp:
		slog.Info("Stomp not implemented yet")
		os.Exit(0)
		return nil
		// stomp := newStomp(c)
		// return &stomp
	case Websockets:
		slog.Info("Websockets not implemented yet")
		os.Exit(0)
		return nil
		// ws := newWs(c)
		// return &ws
	default:
		slog.Info("Invalid MTP")
		os.Exit(0)
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
		slog.Error("MTP not defined", "mtp", mtp_config)
		os.Exit(1)
	default:
		slog.Error("Invalid MTP", "mtp", mtp_config)
		os.Exit(1)
	}

	return mtp
}

func getConfigDir(path string) string {

	checkPathExists := func(dir string) {
		_, err := os.Stat(dir)
		if err != nil {
			slog.Error("Path does not exist", "path", dir, "error", err)
			os.Exit(1)
		}
	}

	if path == "" {
		path, _ = os.Getwd()
		path = path + DEFAULT_DIR
		slog.Warn("Config path not defined, using default", "path", path)
		return path
	}

	checkPathExists(path)
	return path
}

func getLogger(logToStdout bool, fileLogging config.FileLogging) io.Writer {

	if logToStdout && !fileLogging.Enable {
		return os.Stdout
	}

	if logToStdout && fileLogging.Enable {
		return io.MultiWriter(os.Stdout, getLogFile(fileLogging.LogFolder))
	}

	if !logToStdout && fileLogging.Enable {
		return getLogFile(fileLogging.LogFolder)
	}

	return nil
}

func getLogFile(folder string) *os.File {

	if folder == "" {
		folder, _ = os.Getwd()
		folder = folder + DEFAULT_DIR
		slog.Warn("Logs folder path not defined, using default", "path", folder)
	}

	logFile, err := os.OpenFile(folder+"/oktopus-agents.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) //TODO: tis log file config should be a file instead of a folder
	if err != nil {
		slog.Error("Error to open log file", "error", err)
	}

	// logFile.Close()
	return logFile
}
