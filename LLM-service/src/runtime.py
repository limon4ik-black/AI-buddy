import fcntl
from contextlib import contextmanager

from .config import ROOT


@contextmanager
def single_instance():
    # Lock before importing torch or allocating a second model. The OS releases
    # the lock even when the process is killed; keep the file inode in place.
    with (ROOT / ".model.lock").open("a") as handle:
        try:
            fcntl.flock(handle, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            raise RuntimeError(
                "AI Buddy уже запущен. Остановите предыдущий процесс перед запуском "
                "нового. Используйте один worker без --reload."
            ) from None
        try:
            yield
        finally:
            fcntl.flock(handle, fcntl.LOCK_UN)
