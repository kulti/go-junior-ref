package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

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
