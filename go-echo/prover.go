package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Result struct {
	Msg string
	Img []byte
	Tex string
}

func prove(sequent, memory string, timeout int, enableNotification bool) (*Result, error) {
	// lock
	mutex.Lock()
	defer mutex.Unlock()

	// create temp dir
	dir, err := os.MkdirTemp(".", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)

	// change dir
	if err := os.Chdir(dir); err != nil {
		return nil, err
	}
	defer os.Chdir("..")

	// symlink ../prover
	if err := os.Symlink("../prover", "prover"); err != nil {
		return nil, err
	}

	// run prover
	msg, err := runProver(sequent, memory, timeout)
	if err != nil {
		return nil, err
	}
	// make dvi
	msgDVI, err := makeDVI()
	if err != nil {
		return nil, err
	}
	// make png
	msgPNG, err := makePNG()
	if err != nil {
		return nil, err
	}
	msg += msgDVI + msgPNG

	// create response
	res := &Result{
		Msg: msg,
	}

	// if out.png exists
	if _, err := os.Stat("out.png"); err == nil {
		if enableNotification {
			notifyLineWithImage(msg)
		}
		// read out.png
		img, err := os.ReadFile("out.png")
		if err != nil {
			return nil, err
		}
		res.Img = img
		// read out.tex
		tex, err := os.ReadFile("out.tex")
		if err != nil {
			return nil, err
		}
		res.Tex = string(tex)
	} else {
		if enableNotification {
			notifyLine(msg)
		}
	}

	return res, nil
}

func runProver(sequent, memory string, timeout int) (string, error) {
	// run the command
	stdout, stderr, err := runCommand("../prover.sh", sequent, memory, strconv.Itoa(timeout))

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
	if err != nil {
		return "", errors.New("binary execution error.\n" + "stdout: " + stdout + "\n" + "stderr: " + stderr)
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
	stdout, stderr, commandErr := runCommand("latex", "-halt-on-error", "-interaction=nonstopmode", "out.tex")

	// Dimension too large
	if strings.Contains(stdout, "Dimension too large") {
		return " The proof tree is too large to output: Dimension too large.", nil
	}
	// other err
	if _, err := os.Stat("out.dvi"); err != nil || commandErr != nil {
		return "", errors.New("could not compile tex file.\n" + "stdout: " + stdout + "\n" + "stderr: " + stderr)
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
	stdout, stderr, commandErr := runCommand("dvipng", "out.dvi", "-o", "out.png")

	// DVI stack overflow
	if strings.Contains(stderr, "DVI stack overflow") {
		return " The proof tree is too large to output: DVI stack overflow.", nil
	}
	// other err
	if _, err := os.Stat("out.png"); err != nil || commandErr != nil {
		return "", errors.New("could not compile dvi file.\n" + "stdout: " + stdout + "\n" + "stderr: " + stderr)
	}
	// success
	return "", nil
}

func runCommand(name string, arg ...string) (string, string, error) {
	cmd := exec.Command(name, arg...)

	// create a buffer to capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// run the command
	err := cmd.Run()

	return stdout.String(), stderr.String(), err
}
