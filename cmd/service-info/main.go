package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
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
		serviceCommit := state.Commit()
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
