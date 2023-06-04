import os
import subprocess
from subprocess import CalledProcessError, TimeoutExpired

from notify import notify_line


def make_proof_tree(sequent: str, size: str, timeout: float) -> str:
    return prove(sequent, size, timeout) + make_dvi() + make_img()


def prove(sequent: str, size: str, timeout: float) -> str:
    try:
        cmd = ['./prover', '-Xmx' + size, 'out', sequent]
        result = subprocess.run(cmd, capture_output=True, check=True, text=True, timeout=timeout)
        return result.stdout
    except CalledProcessError as e:
        # OutOfMemoryError
        if 'OutOfMemoryError' in e.stderr:
            if not e.stdout:
                return 'Proof Failed: OutOfMemoryError.'
            else:
                return e.stdout + ' The proof tree is too large to output: OutOfMemoryError.'
        # StackOverflowError
        if 'StackOverflowError' in e.stderr:
            return 'Proof Failed: StackOverflowError.'
        # Other error
        notify_line(f'Error: {e}')
        notify_line(f'stdout: {e.stdout}')
        notify_line(f'stderr: {e.stderr}')
        return 'An unexpected error has occurred: Binary Execution Error. The bug report was sent to the developer.'
    except TimeoutExpired as e:
        # Timeout
        if not e.stdout:
            return 'Proof Failed: Timeout.'
        else:
            if isinstance(e.stdout, bytes):
                e.stdout = e.stdout.decode()
            return e.stdout + ' The proof tree is too large to output: Timeout.'


def make_dvi() -> str:
    if not os.path.exists('out.tex'):
        return ''
    try:
        cmd = ['latex', '-halt-on-error', '-interaction=nonstopmode', 'out.tex']
        result = subprocess.run(cmd, capture_output=True, check=True, text=True)
        # DVIが生成されないとき
        if not os.path.exists('out.dvi'):
            raise CalledProcessError(1, cmd, result.stdout, result.stderr)
        # 正常時
        return ''
    except CalledProcessError as e:
        # Dimension too large
        if 'Dimension too large' in e.stdout:
            return ' The proof tree is too large to output: Dimension too large.'
        # Other error
        notify_line(f'Error: {e}')
        notify_line(f'stdout: {e.stdout}')
        notify_line(f'stderr: {e.stderr}')
        return ' An unexpected error has occurred: Could not compile tex file. ' \
               'The bug report was sent to the developer.'


def make_img() -> str:
    if not os.path.exists('out.dvi'):
        return ''
    try:
        cmd = ['dvipng', 'out.dvi', '-o', 'out.png']
        result = subprocess.run(cmd, capture_output=True, check=True, text=True)
        # PNGが生成されないとき
        if not os.path.exists('out.png'):
            raise CalledProcessError(1, cmd, result.stdout, result.stderr)
        # 正常時
        return ''
    except (CalledProcessError, TimeoutExpired) as e:
        # DVI stack overflow
        if 'DVI stack overflow' in e.stderr:
            return ' The proof tree is too large to output: DVI stack overflow.'
        # Other error
        notify_line(f'Error: {e}')
        notify_line(f'stdout: {e.stdout}')
        notify_line(f'stderr: {e.stderr}')
        return ' An unexpected error has occurred: Could not compile dvi file. ' \
               'The bug report was sent to the developer.'
