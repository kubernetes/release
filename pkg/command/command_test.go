package command

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuccess(t *testing.T) {
	res, err := New("echo hi").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Zero(t, res.ExitCode())
}

func TestSuccessWithWorkingDir(t *testing.T) {
	res, err := NewWithWorkDir("/", "ls -1").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
	require.Zero(t, res.ExitCode())
}

func TestFailureWithWrongWorkingDir(t *testing.T) {
	_, err := NewWithWorkDir("/should/not/exist", "ls -1").Run()
	require.NotNil(t, err)
}

func TestSuccessSilent(t *testing.T) {
	res, err := New("echo hi").RunSilent()
	require.Nil(t, err)
	require.True(t, res.Success())
}

func TestSuccessSepareted(t *testing.T) {
	res, err := New("echo", "hi").RunSilent()
	require.Nil(t, err)
	require.True(t, res.Success())
}

func TestSuccessSingleArgument(t *testing.T) {
	res, err := New("echo").Run()
	require.Nil(t, err)
	require.True(t, res.Success())
}

func TestSuccessNoArgument(t *testing.T) {
	res, err := New("").Run()
	require.NotNil(t, err)
	require.Nil(t, res)
}

func TestSuccessOutput(t *testing.T) {
	res, err := New("echo -n 'hello world'").Run()
	require.Nil(t, err)
	require.Equal(t, "'hello world'", res.Output())
}

func TestSuccessOutputSeparated(t *testing.T) {
	res, err := New("echo", "-n", "hello").Run()
	require.Nil(t, err)
	require.Equal(t, "hello", res.Output())
}

func TestFailureStdErr(t *testing.T) {
	res, err := New("cat /not/valid").Run()
	require.Nil(t, err)
	require.False(t, res.Success())
	require.Equal(t, res.ExitCode(), 1)
}

func TestFailureNotExisting(t *testing.T) {
	res, err := New("/not/valid").Run()
	require.NotNil(t, err)
	require.Nil(t, res)
}

func TestSuccessExecute(t *testing.T) {
	err := Execute("echo -n hi", "ho")
	require.Nil(t, err)
}

func TestFailureExecute(t *testing.T) {
	err := Execute("cat", "/not/invalid")
	require.NotNil(t, err)
}

func TestAvailableSuccessValidCommand(t *testing.T) {
	res := Available("echo")
	require.True(t, res)
}

func TestAvailableSuccessEmptyCommands(t *testing.T) {
	res := Available()
	require.True(t, res)
}

func TestAvailableFailure(t *testing.T) {
	res := Available("echo", "this-command-should-not-exist")
	require.False(t, res)
}

func TestSuccessRunSuccess(t *testing.T) {
	require.Nil(t, New("echo hi").RunSuccess())
}

func TestFailureRunSuccess(t *testing.T) {
	require.NotNil(t, New("cat /not/available").RunSuccess())
}

func TestSuccessRunSilentSuccess(t *testing.T) {
	require.Nil(t, New("echo hi").RunSilentSuccess())
}

func TestFailureRunSuccessSilent(t *testing.T) {
	require.NotNil(t, New("cat /not/available").RunSilentSuccess())
}
