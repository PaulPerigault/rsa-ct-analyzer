package cli

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/PaulPerigault/rsa-ct-analyzer/internal/ct/mocklog"
	"github.com/PaulPerigault/rsa-ct-analyzer/internal/keys"
)

func newMockLogCmd(streams IO) *cobra.Command {
	var o mockLogOptions
	cmd := &cobra.Command{
		Use:   "mock-log path...",
		Short: "Serve local certificates as a simulated CT log (offline demos and tests)",
		Long: `Starts a read-only RFC 6962 log (get-sth, get-entries) over the certificates
found in the given paths. It listens on loopback by default. Use it with
"crawl --log-url http://127.0.0.1:8080" to exercise the whole pipeline with no
network access.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMockLog(cmd.Context(), streams, args, o)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&o.addr, "addr", envOr("RSA_CT_MOCK_ADDR", "127.0.0.1:8080"), "listen address (env RSA_CT_MOCK_ADDR)")
	fl.IntVar(&o.pageSize, "page-size", 256, "max entries per get-entries answer")
	fl.IntVar(&o.precertEvery, "precert-every", 0, "publish every Nth certificate as a precertificate (0 = never)")
	return cmd
}

type mockLogOptions struct {
	addr                   string
	pageSize, precertEvery int
}

func runMockLog(ctx context.Context, streams IO, paths []string, o mockLogOptions) error {
	loaded := keys.Load(paths, streams.In)
	var entries []mocklog.Entry
	for i, k := range loaded.Keys {
		if k.Raw == nil {
			continue // public keys and private keys are not certificates
		}
		entries = append(entries, mocklog.Entry{
			DER:     k.Raw,
			Precert: o.precertEvery > 0 && (i+1)%o.precertEvery == 0,
		})
	}
	if len(entries) == 0 {
		return errors.New("no certificate found to publish")
	}

	ln, err := net.Listen("tcp", o.addr)
	if err != nil {
		return err
	}
	srv := &http.Server{Handler: mocklog.Handler(entries, o.pageSize), ReadHeaderTimeout: 5 * time.Second}
	fmt.Fprintf(streams.Err, "mock-log: serving %d certificates on http://%s\n", len(entries), ln.Addr())

	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()
	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return srv.Shutdown(shutdown)
	}
}
