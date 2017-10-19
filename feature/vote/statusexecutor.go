package vote

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jakevoytko/crbot/api"
	"github.com/jakevoytko/crbot/log"
	"github.com/jakevoytko/crbot/model"
)

type StatusExecutor struct {
	modelHelper *ModelHelper
}

func NewStatusExecutor(modelHelper *ModelHelper) *StatusExecutor {
	return &StatusExecutor{
		modelHelper: modelHelper,
	}
}

// GetType returns the type of this feature.
func (e *StatusExecutor) GetType() int {
	return model.Type_VoteStatus
}

const (
	MsgMinutesRemaining  = "%d minutes remaining"
	MsgNoActiveVote      = "No active vote"
	MsgOneVoteAgainst    = "1 vote against"
	MsgOneVoteFor        = "1 vote for"
	MsgSecondsRemaining  = "%d seconds remaining"
	MsgSpacer            = "-----"
	MsgStatusVoteFailing = "Vote is failing"
	MsgStatusVotePassing = "Vote is passing"
	MsgStatusVotesNeeded = "5 votes for or 5 votes against needed"
	MsgVoteOwner         = "Vote started by %s: "
	MsgVotesAgainst      = "%d votes against"
	MsgVotesFor          = "%d votes for"
)

// Execute prints the status of the current vote.
func (e *StatusExecutor) Execute(s api.DiscordSession, channel string, command *model.Command) {
	ok, err := e.modelHelper.IsVoteActive()
	if err != nil {
		log.Fatal("Error reading vote status", err)
	}
	if !ok {
		if _, err := s.ChannelMessageSend(channel, MsgNoActiveVote); err != nil {
			log.Fatal("Unable to send no-active-vote message to user", err)
		}
		return
	}

	vote, err := e.modelHelper.MostRecentVote()
	if err != nil {
		log.Fatal("Error pulling most recent vote", err)
	}
	if vote == nil {
		log.Fatal("Nil vote found after vote already active", errors.New("Vote should not be null"))
	}

	// The below creates a string like this:
	//
	// Vote started by @SomeoneElse: Votekick @Jake?
	// -----
	// 12 minutes remaining
	// 5 votes needed to pass. 3 votes for, 1 vote against

	messages := []string{}

	// Add the owner string.
	owner, err := s.User(strconv.FormatInt(vote.UserID, 10))
	if err != nil {
		log.Fatal("Error fetching the owner when rendering a vote response", err)
	}
	// Status line and message.
	messages = append(messages, fmt.Sprintf(MsgVoteOwner, owner.Username)+vote.Message)

	// Spacer
	messages = append(messages, MsgSpacer)

	messages = append(messages, StatusLine(vote))

	finalMessage := strings.Join(messages, "\n")
	if _, err := s.ChannelMessageSend(channel, finalMessage); err != nil {
		log.Info("Failed to send vote message", err)
	}
}

// Returns the full status line. To be reused in the ballot executor.
func StatusLine(vote *Vote) string {
	// Add the vote totals.
	statusStr := statusString(vote)
	votesFor := len(vote.VotesFor)
	votesAgainst := len(vote.VotesAgainst)
	votesForStr := MsgOneVoteFor
	if votesFor != 1 {
		votesForStr = fmt.Sprintf(MsgVotesFor, votesFor)
	}
	votesAgainstStr := MsgOneVoteAgainst
	if votesAgainst != 1 {
		votesAgainstStr = fmt.Sprintf(MsgVotesAgainst, votesAgainst)
	}

	return statusStr + ". " + votesForStr + ", " + votesAgainstStr
}

func statusString(vote *Vote) string {
	if vote.HasEnoughVotes() {
		switch vote.CalculateActiveStatus() {
		case VoteOutcomePassed:
			return MsgStatusVotePassing

		default:
			return MsgStatusVoteFailing
		}
	}
	return MsgStatusVotesNeeded
}
