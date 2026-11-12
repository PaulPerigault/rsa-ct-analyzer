// Command rsa-ct-analyzer audits RSA keys from local files and CT logs.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root := cli.NewRootCmd(cli.StdIO())
	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		stop()
		var exit *cli.ExitCodeError
		if errors.As(err, &exit) {
			os.Exit(exit.Code)
		}
		os.Exit(cli.ExitError)
	}
}
