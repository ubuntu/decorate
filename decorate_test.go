package decorate_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/ubuntu/decorate"
)

func TestOnErrorWithNoError(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		defer decorate.OnError(&err, "My format with %s as argument", "arg")
		return nil
	}()

	require.NoError(t, err, "No decoration as no error")
}

func TestOnErrorWithError(t *testing.T) {
	t.Parallel()

	err := func() (err error) {
		defer decorate.OnError(&err, "My format with %s as argument", "arg")
		return errors.New("Some error")
	}()

	require.Equal(t, errors.New("My format with arg as argument: Some error").Error(), err.Error(), "Should annotate with error format")
}

func TestLogOnErrorWithNoError(t *testing.T) {
	out := captureLogs(t)

	_ = func() (err error) {
		defer decorate.LogOnError(&err)
		return nil
	}()

	require.Empty(t, out(), "No error  no log")
}

func TestLogOnErrorWithError(t *testing.T) {
	out := captureLogs(t)

	err := func() (err error) {
		defer decorate.LogOnError(&err)
		return errors.New("Some error")
	}()

	require.Contains(t, out(), err.Error(), "The error should be logged")
}

func TestLogOnErrorContextWithNoError(t *testing.T) {
	out := captureLogs(t)

	_ = func() (err error) {
		defer decorate.LogOnErrorContext(context.Background(), &err)
		return nil
	}

	require.Empty(t, out(), "No error  no log")
}

func TestLogOnErrorContextWithError(t *testing.T) {
	out := captureLogs(t)

	err := func() (err error) {
		defer decorate.LogOnErrorContext(context.Background(), &err)
		return errors.New("Some error")
	}()

	require.Contains(t, out(), err.Error(), "The error should be logged")
}

func TestLogFuncOnErrorWithNoError(t *testing.T) {
	out := captureLogs(t)

	f := func() error { return nil }
	decorate.LogFuncOnError(f)

	require.Empty(t, out(), "No error no log")
}

func TestLogFuncOnErrorWithError(t *testing.T) {
	out := captureLogs(t)

	err := errors.New("Some error")
	f := func() error { return err }
	decorate.LogFuncOnError(f)

	require.Contains(t, out(), err.Error(), "The error should be logged")
}

func TestLogFuncOnErrorContextNoError(t *testing.T) {
	out := captureLogs(t)

	f := func() error { return nil }
	decorate.LogFuncOnErrorContext(context.Background(), f)

	require.Empty(t, out(), "No error  no log")
}

func TestLogFuncOnErrorContextWithError(t *testing.T) {
	out := captureLogs(t)

	err := errors.New("Some error")
	f := func() error { return err }
	decorate.LogFuncOnErrorContext(context.Background(), f)

	require.Contains(t, out(), err.Error(), "The error should be logged")
}

// captureLogs captures current logs.
// It returns a function to read the bufferred log output.
// The logs output will be restored when the test ends.
func captureLogs(t *testing.T) (out func() string) {
	t.Helper()

	localLogger := logrus.StandardLogger()
	orig := localLogger.Out
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal("Setup error: creating pipe:", err)
	}
	localLogger.SetOutput(w)

	t.Cleanup(func() {
		localLogger.SetOutput(orig)
	})
	return func() string {
		w.Close()
		var buf bytes.Buffer
		_, errCopy := io.Copy(&buf, r)
		if errCopy != nil {
			t.Fatal("Setup error: couldn’t get buffer content:", err)
		}
		return buf.String()
	}
}
