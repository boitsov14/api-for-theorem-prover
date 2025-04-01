package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProve(t *testing.T) {
	assert := assert.New(t)

	sequent := "P or not P"
	re, err := prove(sequent, "2g", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "Provable.")
	assert.NotNil(re.Img)
	assert.NotNil(re.Tex)

	sequent = "P"
	re, err = prove(sequent, "2g", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "Unprovable.")

	sequent = "あ"
	re, err = prove(sequent, "2g", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "Illegal Argument")

	sequent = ""
	re, err = prove(sequent, "2g", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "Parse Error")

	sequent = "(((((((((a⇔b)⇔c)⇔d)⇔e)⇔f)⇔g)⇔h)⇔i)⇔(a⇔(b⇔(c⇔(d⇔(e⇔(f⇔(g⇔(h⇔i)))))))))"

	re, err = prove(sequent, "2g", 1, false)
	require.NoError(t, err)
	assert.Equal("Proof Failed: Timeout.", re.Msg)

	re, err = prove(sequent, "2g", 5, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "The proof tree is too large to output: Timeout.")

	re, err = prove(sequent, "10m", 20, false)
	require.NoError(t, err)
	assert.Equal("Proof Failed: OutOfMemoryError.", re.Msg)

	sequent = "((o11 or o12 or o13) and (o21 or o22 or o23) and (o31 or o32 or o33) and (o41 or o42 or o43)) -> ((o11 and o21) or (o11 and o31) or (o11 and o41) or (o21 and o31) or (o21 and o41) or (o31 and o41) or (o12 and o22) or (o12 and o32) or (o12 and o42) or (o22 and o32) or (o22 and o42) or (o32 and o42) or (o13 and o23) or (o13 and o33) or (o13 and o43) or (o23 and o33) or (o23 and o43) or (o33 and o43))"
	re, err = prove(sequent, "10m", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "The proof tree is too large to output: OutOfMemoryError.")

	sequent = "P to ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~P"
	re, err = prove(sequent, "10m", 10, false)
	require.NoError(t, err)
	assert.Equal("Proof Failed: StackOverflowError.", re.Msg)

	sequent = "((((a⇔b)⇔c)⇔d)⇔(a⇔(b⇔(c⇔d))))"
	re, err = prove(sequent, "2g", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "The proof tree is too large to output: Dimension too large.")

	sequent = "P to ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~P"
	re, err = prove(sequent, "2g", 10, false)
	require.NoError(t, err)
	assert.Contains(re.Msg, "The proof tree is too large to output: DVI stack overflow.")
}
