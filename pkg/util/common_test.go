package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const validCommand = "sh"

func TestCommandsAvailableSuccessValidCommand(t *testing.T) {
	commands := []string{validCommand}
	res := CommandsAvailable(commands)
	require.True(t, res)
}

func TestCommandsAvailableSuccessEmptyCommands(t *testing.T) {
	commands := []string{}
	res := CommandsAvailable(commands)
	require.True(t, res)
}

func TestCommandsAvailableFailure(t *testing.T) {
	commands := []string{validCommand, "this-command-should-not-exist"}
	res := CommandsAvailable(commands)
	require.False(t, res)
}
