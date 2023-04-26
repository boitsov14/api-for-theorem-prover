package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

func makeProofTree(id, sequent, size string, timeout time.Duration) string {
	return prove(id, sequent, size, timeout) + makeDVI(id) + makeImg(id)
}

func prove(id, sequent, size string, timeout time.Duration) string {

	stdout, stderr, err := CommandExecWithTimeout(timeout, "../prover", "-Xmx"+size, id, sequent)

	// Timeoutしたとき
	if err == context.DeadlineExceeded {
		if stdout == "" {
			return "Proof Failed: Timeout."
		} else {
			return stdout + " The proof tree is too large to output: Timeout."
		}
	}

	// OutOfMemoryErrorしたとき
	if strings.Contains(stderr, "OutOfMemoryError") {
		if stdout == "" {
			return "Proof Failed: OutOfMemoryError."
		} else {
			return stdout + " The proof tree is too large to output: OutOfMemoryError."
		}
	}

	// StackOverflowErrorしたとき
	if strings.Contains(stderr, "StackOverflowError") {
		return "Proof Failed: StackOverflowError."
	}

	// その他のエラーが発生したとき
	if stdout == "" || stderr != "" || err != nil {
		fmt.Println(stdout, stderr, err)
		return stdout + "An unexpected error has occurred: Java exec failure."
	}

	return stdout
}

func makeDVI(id string) string {

	if !exists(id + ".tex") {
		return ""
	}

	stdout, stderr, err := CommandExec("latex", "-halt-on-error", id+".tex")

	// Dimension too largeのとき
	if strings.Contains(stdout, "Dimension too large") {
		return " The proof tree is too large to output: Dimension too large."
	}

	// その他の理由によりDVIが生成されないとき
	if !exists(id + ".dvi") {
		fmt.Println(stdout, stderr, err)
		return " An unexpected error has occurred: Could not compile tex file."
	}

	return ""
}

func makeImg(id string) string {

	if !exists(id + ".dvi") {
		return ""
	}

	stdout, stderr, err := CommandExec("dvipng", id+".dvi", "-o", id+".png")

	// DVI stack overflowのとき
	if strings.Contains(stderr, "DVI stack overflow") {
		return " The proof tree is too large to output: DVI stack overflow."
	}

	// その他の理由によりPNGが生成されないとき
	if !exists(id + ".png") {
		fmt.Println(stdout, stderr, err)
		return " An unexpected error has occurred: Could not compile dvi file."
	}

	return ""
}
