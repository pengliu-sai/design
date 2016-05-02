package http_api

import (
	"net"
	"net/http"
	"log"
	"fmt"
	"strings"
	"framework/design/tools/app"
)

type logWriter struct {
	app.Logger
}

func (l logWriter) Write(p []byte) (int, error) {
	l.Logger.Output(2, string(p))
	return len(p), nil
}

func Serve(listener net.Listener, handler http.Handler, proto string, l app.Logger) {
	l.Output(2, fmt.Sprintf("%s: listening on %s", proto, listener.Addr()))

	server := &http.Server{
		Handler: handler,
		ErrorLog: log.New(logWriter{l}, "", 0),
	}
	err := server.Serve(listener)

	if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
		l.Output(2, fmt.Sprintf("ERROR: http.Serve() - %s", err))
	}

	l.Output(2, fmt.Sprintf("%s: closing %s", proto, listener.Addr()))
}
