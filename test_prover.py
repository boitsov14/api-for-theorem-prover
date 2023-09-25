import os
from tempfile import TemporaryDirectory

from prover import make_proof_tree as make_proof_tree_base


def make_proof_tree(sequent: str, memory: str = '2g', timeout: int = 10) -> str:
    with TemporaryDirectory(dir=os.getcwd()) as work:
        try:
            os.chdir(work)
            os.symlink('../prover', 'prover')
            msg = make_proof_tree_base(sequent, memory, timeout)
            assert 'An unexpected error has occurred' not in msg
            return msg
        finally:
            os.chdir('..')


def test_make_proof_tree():
    seq = 'P or not P'
    assert 'Provable' in make_proof_tree(seq)

    seq = 'P'
    assert 'Unprovable' in make_proof_tree(seq)

    seq = '(((((((((a⇔b)⇔c)⇔d)⇔e)⇔f)⇔g)⇔h)⇔i)⇔(a⇔(b⇔(c⇔(d⇔(e⇔(f⇔(g⇔(h⇔i)))))))))'
    assert 'Proof Failed: Timeout.' == make_proof_tree(seq, timeout=1)
    assert 'The proof tree is too large to output: Timeout.' in make_proof_tree(seq, timeout=5)
    assert 'Proof Failed: OutOfMemoryError.' == make_proof_tree(seq, '10m')

    seq = ('((o11 ∨ o12 ∨ o13) ∧ (o21 ∨ o22 ∨ o23) ∧ (o31 ∨ o32 ∨ o33) ∧ (o41 ∨ o42 ∨ o43)) → ((o11 ∧ o21) ∨ (o11 ∧ '
           'o31) ∨ (o11 ∧ o41) ∨ (o21 ∧ o31) ∨ (o21 ∧ o41) ∨ (o31 ∧ o41) ∨ (o12 ∧ o22) ∨ (o12 ∧ o32) ∨ (o12 ∧ o42) ∨ '
           '(o22 ∧ o32) ∨ (o22 ∧ o42) ∨ (o32 ∧ o42) ∨ (o13 ∧ o23) ∨ (o13 ∧ o33) ∨ (o13 ∧ o43) ∨ (o23 ∧ o33) ∨ (o23 ∧ '
           'o43) ∨ (o33 ∧ o43))')
    assert 'The proof tree is too large to output: OutOfMemoryError.' in make_proof_tree(seq, '10m')

    seq = ('P to ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~'
           '~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~'
           '~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~'
           '~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~'
           '~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~P')
    assert 'Proof Failed: StackOverflowError.' == make_proof_tree(seq, '10m')

    seq = '((((a⇔b)⇔c)⇔d)⇔(a⇔(b⇔(c⇔d))))'
    assert 'The proof tree is too large to output: Dimension too large.' in make_proof_tree(seq)

    seq = 'P to ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~P'
    assert 'The proof tree is too large to output: DVI stack overflow.' in make_proof_tree(seq)
