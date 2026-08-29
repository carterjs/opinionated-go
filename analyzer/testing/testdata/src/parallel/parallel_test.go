package parallel

import "testing"

func TestParse(t *testing.T) {
	t.Parallel()
	t.Run("returns error when input is empty", func(t *testing.T) {
		t.Parallel()
	})
}

func TestFormat(t *testing.T) { // want "does not call t.Parallel"
	t.Run("returns the formatted document", func(t *testing.T) { // want "subtest does not call t.Parallel"
		_ = Format(nil)
	})
}

// TestConfigFromEnv is exempt: t.Setenv panics if combined with t.Parallel,
// so it must not be required here.
func TestConfigFromEnv(t *testing.T) {
	t.Setenv("APP_CONFIG_PATH", "/tmp/config.json")
	_ = Format(nil)
}

// TestWorkingDirectory is exempt for the same reason, via t.Chdir.
func TestWorkingDirectory(t *testing.T) {
	t.Chdir("/tmp")
	_ = Format(nil)
}

// TestWithEnvSubtest: the outer test is exempt because a subtest inside it
// calls t.Setenv, which panics if any ancestor is parallel.
func TestWithEnvSubtest(t *testing.T) {
	t.Run("reads the configured path", func(t *testing.T) {
		t.Setenv("APP_CONFIG_PATH", "/tmp/config.json")
		_ = Format(nil)
	})
}
