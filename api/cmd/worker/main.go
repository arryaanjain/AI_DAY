package main

import "log/slog"

// Long-running generation work will run in this separate process.
func main() {
	slog.Info("worker started", "status", "queue consumers will be configured in the generation phase")
}
