package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

var (
	errMissedCommandName  = errors.New("missed command name")
	errUnknownCommandName = errors.New("unknown command name")
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errMissedCommandName
	}

	cmdName := os.Args[1]
	os.Args = os.Args[1:]

	switch cmdName {
	case "versions":
		return runVersions()
	default:
		return fmt.Errorf("%w: %s", errUnknownCommandName, cmdName)
	}
}

func runVersions() error {
	serviceNames := flag.String("s", "", "service name(s)")
	flag.Parse()

	services := strings.Split(*serviceNames, ",")
	if *serviceNames == "" {
		services = allServices()
	}

	serviceStates := make([]serviceState, len(services))
	for i, s := range services {
		serviceStates[i] = fetchServiceState(s)
	}

	for _, s := range serviceStates {
		if s.unknown {
			fmt.Printf("\033[1;33m%s  [UNKNOWN]\033[0m\n", s.name)
			continue
		}
		if s.err != "" {
			fmt.Printf("\033[0;31m%s  [ERROR] %s\033[0m\n", s.name, s.err)
			continue
		}
		fmt.Printf("\033[0;32m%s  [OK] %s\033[0m\n", s.name, s.meta)
	}

	return nil
}

var servicePorts = map[string]int{
	"api-server": 8090,
}

func allServices() []string {
	services := make([]string, 0, len(servicePorts))
	for name := range servicePorts {
		services = append(services, name)
	}
	sort.Strings(services)
	return services
}

type serviceState struct {
	name    string
	unknown bool
	err     string
	meta    string
}

func fetchServiceState(serviceName string) serviceState {
	serviceState := serviceState{name: serviceName}
	port, ok := servicePorts[serviceName]
	if !ok {
		serviceState.unknown = true
		return serviceState
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/debug/info", port)
	resp, err := http.Get(url)
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
