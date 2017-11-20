package learn

import (
	"testing"

	"github.com/jakevoytko/crbot/testutil"
)

func TestLearn_NoResponse(t *testing.T) {
	runner := testutil.NewRunner(t)

	// Commands that should never return a response.
	runner.SendMessageWithoutResponse(testutil.MainChannelID, "?")
	runner.SendMessageWithoutResponse(testutil.MainChannelID, "!")
	runner.SendMessageWithoutResponse(testutil.MainChannelID, ".")
	runner.SendMessageWithoutResponse(testutil.MainChannelID, "")
	runner.SendMessageWithoutResponse(testutil.MainChannelID, "!help")
	runner.SendMessageWithoutResponse(testutil.MainChannelID, "help")
	runner.SendMessageWithoutResponse(testutil.MainChannelID, ".help")
}

func TestLearn_WrongFormat(t *testing.T) {
	runner := testutil.NewRunner(t)

	// Basic learn responses.
	// Wrong call format
	runner.SendMessage(testutil.MainChannelID, "?learn", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn test", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn ?call response", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn !call response", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn /call response", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn ", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn multi\nline\ncall response", MsgHelpLearn)
	// Wrong response format.
	runner.SendMessage(testutil.MainChannelID, "?learn call ?response", MsgHelpLearn)
	runner.SendMessage(testutil.MainChannelID, "?learn call !response", MsgHelpLearn)
}