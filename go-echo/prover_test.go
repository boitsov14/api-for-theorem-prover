package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func proveForTest(sequent, memory string, timeout int) (string, error) {
	// create temp dir
	dir, err := os.MkdirTemp(".", "")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	// change dir
	if err := os.Chdir(dir); err != nil {
		return "", err
	}
	defer os.Chdir("..")

	// symlink ../prover
	if err := os.Symlink("../prover", "prover"); err != nil {
		return "", err
	}

	// run prover
	msg, err := prove(sequent, memory, timeout)
	if err != nil {
		return "", err
	}
	// make dvi
	msgDVI, err := makeDVI()
	if err != nil {
		return "", err
	}
	// make png
	msgPNG, err := makePNG()
	if err != nil {
		return "", err
	}
	msg += msgDVI + msgPNG

	return msg, nil
}

func TestProve(t *testing.T) {
	assert := assert.New(t)

	sequent := "P or not P"
	msg, err := proveForTest(sequent, "2g", 10)
	assert.NoError(err)
	assert.Contains(msg, "Provable.")

	sequent = "P"
	msg, err = proveForTest(sequent, "2g", 10)
	assert.NoError(err)
	assert.Contains(msg, "Unprovable.")

	sequent = "(((((((((a⇔b)⇔c)⇔d)⇔e)⇔f)⇔g)⇔h)⇔i)⇔(a⇔(b⇔(c⇔(d⇔(e⇔(f⇔(g⇔(h⇔i)))))))))"

	msg, err = proveForTest(sequent, "2g", 1)
	assert.NoError(err)
	assert.Equal("Proof Failed: Timeout.", msg)

	msg, err = proveForTest(sequent, "2g", 5)
	assert.NoError(err)
	assert.Contains(msg, "The proof tree is too large to output: Timeout.")

	msg, err = proveForTest(sequent, "10m", 20)
	assert.NoError(err)
	assert.Equal("Proof Failed: OutOfMemoryError.", msg)

	sequent = "((o11 or o12 or o13) and (o21 or o22 or o23) and (o31 or o32 or o33) and (o41 or o42 or o43)) -> ((o11 and o21) or (o11 and o31) or (o11 and o41) or (o21 and o31) or (o21 and o41) or (o31 and o41) or (o12 and o22) or (o12 and o32) or (o12 and o42) or (o22 and o32) or (o22 and o42) or (o32 and o42) or (o13 and o23) or (o13 and o33) or (o13 and o43) or (o23 and o33) or (o23 and o43) or (o33 and o43))"
	msg, err = proveForTest(sequent, "10m", 10)
	assert.NoError(err)
	assert.Contains(msg, "The proof tree is too large to output: OutOfMemoryError.")

	sequent = "P to ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~P"
	msg, err = proveForTest(sequent, "10m", 10)
	assert.NoError(err)
	assert.Equal("Proof Failed: StackOverflowError.", msg)

	sequent = "((((a⇔b)⇔c)⇔d)⇔(a⇔(b⇔(c⇔d))))"
	msg, err = proveForTest(sequent, "2g", 10)
	assert.NoError(err)
	assert.Contains(msg, "The proof tree is too large to output: Dimension too large.")

	sequent = "P to ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~P"
	msg, err = proveForTest(sequent, "2g", 10)
	assert.NoError(err)
	assert.Contains(msg, "The proof tree is too large to output: DVI stack overflow.")
}
