package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type serviceState struct {
	name    string
	unknown bool
	err     string
	meta    string
}

func (s serviceState) Commit() string {
	n := strings.Index(s.meta, "Build Version: ")
	if n == -1 {
		return ""
	}

	commit, _, _ := strings.Cut(s.meta[n+len("Build Version: "):], " ")
	return commit
}

func fetchServiceState(ctx context.Context, serviceName string) serviceState {
	serviceState := serviceState{name: serviceName}
	port, ok := servicePorts[serviceName]
	if !ok {
		serviceState.unknown = true
		return serviceState
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/debug/info", port)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		serviceState.err = fmt.Sprintf("building request: %v", err)
		return serviceState
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		serviceState.err = fmt.Sprintf("get service info: %v", err)
		return serviceState
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		serviceState.err = fmt.Sprintf("returned non-ok status code: %d", resp.StatusCode)
		return serviceState
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		serviceState.err = fmt.Sprintf("read response body: %v", err)
		return serviceState
	}
	serviceState.meta = string(body)
	return serviceState
}
