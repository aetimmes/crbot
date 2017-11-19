package main

import (
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jakevoytko/crbot/api"
	"github.com/jakevoytko/crbot/app"
	"github.com/jakevoytko/crbot/feature"
	"github.com/jakevoytko/crbot/feature/help"
	"github.com/jakevoytko/crbot/feature/list"
	"github.com/jakevoytko/crbot/feature/moderation"
	"github.com/jakevoytko/crbot/feature/vote"
	"github.com/jakevoytko/crbot/log"
	"github.com/jakevoytko/crbot/model"
)

func InitializeRegistry(commandMap model.StringMap, voteMap model.StringMap, gist api.Gist, config *app.Config, clock model.UTCClock, timer model.UTCTimer, commandChannel chan<- *model.Command) *feature.Registry {
	// Initializing builtin features.
	// TODO(jvoytko): investigate the circularity that emerged to see if there's
	// a better pattern here.
	featureRegistry := feature.NewRegistry()
	allFeatures := []feature.Feature{
		help.NewFeature(featureRegistry),
		NewLearnFeature(featureRegistry, commandMap),
		list.NewFeature(featureRegistry, commandMap, gist),
		moderation.NewFeature(featureRegistry, config),
		vote.NewFeature(featureRegistry, voteMap, clock, timer, commandChannel),
	}

	for _, f := range allFeatures {
		err := featureRegistry.Register(f)
		if err != nil {
			panic(err)
		}
	}
	return featureRegistry
}

///////////////////////////////////////////////////////////////////////////////
// Controller methods
///////////////////////////////////////////////////////////////////////////////

func handleCommands(featureRegistry *feature.Registry, s api.DiscordSession, commandChannel <-chan *model.Command) {
	for command := range commandChannel {
		var err error // so I don't have to use := in the intercept() call
		for _, interceptor := range featureRegistry.CommandInterceptors() {
			command, err = interceptor.Intercept(command, s)
			if err != nil {
				panic("Ran into error intercepting commands")
			}
		}

		executor := featureRegistry.GetExecutorByType(command.Type)
		if executor != nil {
			if err != nil {
				log.Fatal("Error parsing snowflake", err)
			}
			executor.Execute(s, command.ChannelID, command)
		}
	}
}

// getHandleMessage returns the main handler for incoming messages.
func getHandleMessage(commandMap model.StringMap, featureRegistry *feature.Registry, commandChannel chan<- *model.Command) func(api.DiscordSession, *discordgo.MessageCreate) {
	return func(s api.DiscordSession, m *discordgo.MessageCreate) {
		// Never reply to a bot.
		if m.Author.Bot {
			return
		}

		command, err := parseCommand(commandMap, featureRegistry, m.Content)
		if err != nil {
			log.Info("Error parsing command", err)
			return
		}
		command.Author = m.Author
		channelID, err := model.ParseSnowflake(m.ChannelID)
		if err != nil {
			log.Info("Error parsing channel ID", err)
			return
		}
		command.ChannelID = channelID

		commandChannel <- command
	}
}

// Parses the raw text string from the user. Returns an executable command.
func parseCommand(commandMap model.StringMap, registry *feature.Registry, content string) (*model.Command, error) {
	if !strings.HasPrefix(content, "?") {
		return &model.Command{
			Type: model.Type_None,
		}, nil
	}
	splitContent := strings.Split(content, " ")

	// Parse builtins.
	if parser := registry.GetParserByName(splitContent[0]); parser != nil {
		return parser.Parse(splitContent)
	}

	// See if it's a custom command.
	has, err := commandMap.Has(splitContent[0][1:])
	if err != nil {
		log.Info("Error doing custom parsing", err)
		return nil, err
	}
	if has {
		return registry.FallbackParser.Parse(splitContent)
	}

	// No such command!
	return &model.Command{
		Type: model.Type_Unrecognized,
	}, nil
}
