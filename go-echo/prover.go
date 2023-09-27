package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func prove(sequent, memory string, timeout int) (string, error) {
	// run the command
	stdout, stderr, exitCode, err := runCommand("../prover.sh", sequent, memory, strconv.Itoa(timeout))
	if err != nil {
		return "", err
	}

	// Timeout
	if strings.Contains(stderr, "CPU time limit exceeded") {
		if stdout == "" {
			return "Proof Failed: Timeout.", nil
		} else {
			return stdout + " The proof tree is too large to output: Timeout.", nil
		}
	}
	// OutOfMemoryError
	if strings.Contains(stderr, "OutOfMemoryError") {
		if stdout == "" {
			return "Proof Failed: OutOfMemoryError.", nil
		} else {
			return stdout + " The proof tree is too large to output: OutOfMemoryError.", nil
		}
	}
	// StackOverflowError
	if strings.Contains(stderr, "StackOverflowError") {
		return "Proof Failed: StackOverflowError.", nil
	}
	// other err
	if _, err := os.Stat("out.tex"); err != nil || exitCode != 0 {
		return "", fmt.Errorf("binary execution error.\n"+"stdout: %s\n"+"stderr: %s", stdout, stderr)
	}
	// success
	return stdout, nil
}

func makeDVI() (string, error) {
	// check if out.tex exists
	if _, err := os.Stat("out.tex"); err != nil {
		return "", nil
	}

	// run the command
	stdout, stderr, exitCode, err := runCommand("latex", "-halt-on-error", "-interaction=nonstopmode", "out.tex")
	if err != nil {
		return "", err
	}

	// Dimension too large
	if strings.Contains(stdout, "Dimension too large") {
		return " The proof tree is too large to output: Dimension too large.", nil
	}
	// other err
	if _, err := os.Stat("out.dvi"); err != nil || exitCode != 0 {
		return "", fmt.Errorf("could not compile tex file.\n"+"stdout: %s\n"+"stderr: %s", stdout, stderr)
	}
	// success
	return "", nil
}

func makePNG() (string, error) {
	// check if out.dvi exists
	if _, err := os.Stat("out.dvi"); err != nil {
		return "", nil
	}

	// run the command
	stdout, stderr, exitCode, err := runCommand("dvipng", "out.dvi", "-o", "out.png")
	if err != nil {
		return "", err
	}

	// DVI stack overflow
	if strings.Contains(stderr, "DVI stack overflow") {
		return " The proof tree is too large to output: DVI stack overflow.", nil
	}
	// other err
	if _, err := os.Stat("out.png"); err != nil || exitCode != 0 {
		return "", fmt.Errorf("could not compile dvi file.\n"+"stdout: %s\n"+"stderr: %s", stdout, stderr)
	}
	// success
	return "", nil
}

func runCommand(name string, arg ...string) (string, string, int, error) {
	cmd := exec.Command(name, arg...)

	// create a buffer to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// run the command
	if err := cmd.Run(); err != nil {
		return "", "", -1, err
	}

	return stdout.String(), stderr.String(), cmd.ProcessState.ExitCode(), nil
}
