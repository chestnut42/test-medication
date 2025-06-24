package httpx

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
)

func ServeContext(ctx context.Context, handler http.Handler, addr string) error {
	srv := &http.Server{
		Handler: handler,
		BaseContext: func(net.Listener) context.Context {
			// Cancellation is actually handled below. The server not need to be canceled by the root context.
			return context.WithoutCancel(ctx)
		},
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("net.Listen: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer cancel()
		defer wg.Done()

		err = srv.Serve(l)
	}()

	<-ctx.Done()
	// nolint:errcheck
	_ = srv.Shutdown(context.Background()) // Here we want for the server to get shut down no matter what.
	wg.Wait()
	return err
}
