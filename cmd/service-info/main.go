package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
)

var (
	errMissedCommandName  = errors.New("missed command name")
	errUnknownCommandName = errors.New("unknown command name")
	errMissedRequiredArg  = errors.New("missed require argument")
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	switch cmdName {
	case "versions":
		return runVersions(ctx)
	case "is-deployed":
		return runIsDeployed(ctx)
	default:
		return fmt.Errorf("%w: %s", errUnknownCommandName, cmdName)
	}
}

func runVersions(ctx context.Context) error {
	serviceNames := flag.String("s", "", "service name(s)")
	flag.Parse()

	services := strings.Split(*serviceNames, ",")
	if *serviceNames == "" {
		services = allServices()
	}

	serviceStates := make([]serviceState, len(services))
	for i, s := range services {
		serviceStates[i] = fetchServiceState(ctx, s)
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

func runIsDeployed(ctx context.Context) error {
	serviceNames := flag.String("s", "", "service name(s)")
	commit := flag.String("c", "", "commit hash")
	flag.Parse()

	if *commit == "" {
		return fmt.Errorf("%w: commit hash", errMissedRequiredArg)
	}
	commitMsg, err := commitMessage(ctx, *commit)
	if err != nil {
		return fmt.Errorf("get commit message: %w", err)
	}

	services := strings.Split(*serviceNames, ",")
	if *serviceNames == "" {
		services = allServices()
	}

	type state struct {
		serviceState
		deployed bool
	}
	states := make([]state, len(services))

	for i, service := range services {
		state := state{serviceState: fetchServiceState(ctx, service)}
		serviceCommit := getCommitFromMeta(state.meta)
		if serviceCommit != "" {
			deployed, err := isAncestor(ctx, *commit, serviceCommit)
			if err != nil {
				state.err = "check is deployed: " + err.Error()
			} else {
				state.deployed = deployed
			}
		}

		states[i] = state
	}

	fmt.Printf("Looking up for commit %s:\n%s\n", *commit, commitMsg)
	fmt.Println("---------------")

	for _, state := range states {
		if state.err != "" {
			fmt.Printf("\033[0;31m%s  [ERROR] (%v)\033[0m\n", state.name, state.err)
		} else if state.unknown {
			fmt.Printf("\033[1;33m%s  [UNKNOWN]\033[0m\n", state.name)
		} else if state.deployed {
			fmt.Printf("\033[0;32m%s  [DEPLOYED]\033[0m\n", state.name)
		} else {
			fmt.Printf("\033[0;37m%s  [NOT DEPLOYED] Current=%+v\033[0m\n", state.name, state.meta)
		}
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

func getCommitFromMeta(meta string) string {
	n := strings.Index(meta, "Build Version: ")
	if n == -1 {
		return ""
	}

	commit, _, _ := strings.Cut(meta[n+len("Build Version: "):], " ")
	return commit
}

func isAncestor(ctx context.Context, a, b string) (bool, error) {
	args := []string{"merge-base", "--is-ancestor", a, b}

	cmd := exec.CommandContext(ctx, "git", args...)
	errBuf := &bytes.Buffer{}
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && ee.ExitCode() == 1 {
			return false, nil
		}

		return false, fmt.Errorf("%w (%s)", err, errBuf.Bytes())
	}

	return true, nil
}

func commitMessage(ctx context.Context, commit string) (string, error) {
	args := []string{"log", "--format=%B", "-n", "1", commit}

	cmd := exec.CommandContext(ctx, "git", args...)
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.Stdout = outBuf
	cmd.Stderr = errBuf

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w (%s)", err, errBuf.Bytes())
	}

	return strings.TrimSpace(outBuf.String()), nil
}
