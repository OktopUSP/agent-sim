package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/OktopUSP/agent-sim/internal/config"
	"github.com/OktopUSP/agent-sim/internal/simulator"
)

func main() {
	done := make(chan os.Signal, 1)

	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	conf := config.NewConfig(ctx)

	go simulator.StartDeviceSimulator(conf)

	<-done
	/* ----------------------------- Stop Gracefully ---------------------------- */
	cancel()
	slog.Info("Received signal to stop the simulator. Waiting for all agents to stop")
	conf.Wg.Wait()

	if !conf.BareMetal.Enable {
		conf.Docker.Cli.Close()
	}
	/* -------------------------------------------------------------------------- */

	slog.Info("(⌐■_■) Agent simulator is out!")
}
