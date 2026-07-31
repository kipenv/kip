package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kipenv/kip/internal/cli"
)

func main() {
	os.Exit(run())
}

// run executes the command tree and returns the process exit code. It exists so
// that the signal handler's cleanup runs — os.Exit in main would skip it.
func run() int {
	// Bind the command tree to a context cancelled on SIGINT/SIGTERM, so Ctrl+C
	// aborts an in-flight upload or download instead of killing the process
	// mid-request. Commands reach it through cmd.Context().
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.NewRootCmd().ExecuteContext(ctx); err != nil {
		// On interrupt the surfacing error is whatever the in-flight call
		// reported (wrapped in *url.Error, etc.), so ask the context rather
		// than trying to unwrap down to context.Canceled.
		if ctx.Err() != nil {
			fmt.Fprintln(os.Stderr, "aborted")
			return 130
		}
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 0
}
