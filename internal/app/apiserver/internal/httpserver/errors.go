package httpserver

import "errors"

var errNotReadyToAcceptConnections = errors.New("not ready to accept connections")
