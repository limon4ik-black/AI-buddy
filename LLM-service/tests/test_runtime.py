import pytest

from src.runtime import single_instance


def test_second_instance_is_rejected_and_lock_is_released(tmp_path, monkeypatch):
    monkeypatch.setattr("src.runtime.ROOT", tmp_path)
    with single_instance():
        with pytest.raises(RuntimeError, match="уже запущен"):
            with single_instance():
                pytest.fail("Second instance acquired the lock")
    with single_instance():
        pass
