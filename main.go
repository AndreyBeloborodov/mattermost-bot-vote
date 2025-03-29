package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/rs/zerolog"
)

func main() {

	app := &application{
		logger: zerolog.New(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC822,
			},
		).With().Timestamp().Logger(),
	}

	app.config = loadConfig()
	app.logger.Info().Str("config", fmt.Sprint(app.config)).Msg("")

	setupGracefulShutdown(app)

	// Create a new mattermost client.
	app.mattermostClient = model.NewAPIv4Client(app.config.mattermostServer.String())

	// Login.
	app.mattermostClient.SetToken(app.config.mattermostToken)

	if user, resp, err := app.mattermostClient.GetUser("me", ""); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not log in")
	} else {
		app.logger.Debug().Interface("user", user).Interface("resp", resp).Msg("")
		app.logger.Info().Msg("Logged in to mattermost")
		app.mattermostUser = user
	}

	// Find and save the bot's team to app struct.
	if team, resp, err := app.mattermostClient.GetTeamByName(app.config.mattermostTeamName, ""); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not find team. Is this bot a member ?")
	} else {
		app.logger.Debug().Interface("team", team).Interface("resp", resp).Msg("")
		app.mattermostTeam = team
	}

	var err error
	if app.voteRepo, err = NewVoteRepository(); err != nil {
		app.logger.Fatal().Err(err).Msg("Connection Tarantool error")
	}
	app.logger.Info().Msg("Tarantool ready!")

	// Listen to live events coming in via websocket.
	listenToEvents(app)
}

func setupGracefulShutdown(app *application) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			if app.mattermostWebSocketClient != nil {
				app.logger.Info().Msg("Closing websocket connection")
				app.mattermostWebSocketClient.Close()
			}
			app.logger.Info().Msg("Shutting down")
			os.Exit(0)
		}
	}()
}

func sendMsgToChannel(app *application, channelId string, msg string, replyToId string) {
	// Note that replyToId should be empty for a new post.
	// All replies in a thread should reply to root.

	post := &model.Post{}
	post.ChannelId = channelId
	post.Message = msg

	post.RootId = replyToId

	if _, _, err := app.mattermostClient.CreatePost(post); err != nil {
		app.logger.Error().Err(err).Str("RootID", replyToId).Msg("Failed to create post")
	}
}

func listenToEvents(app *application) {
	var err error
	failCount := 0
	maxBackoff := 60 * time.Second // Максимальное время задержки между попытками

	for {
		app.mattermostWebSocketClient, err = model.NewWebSocketClient4(
			fmt.Sprintf("ws://%s", app.config.mattermostServer.Host+app.config.mattermostServer.Path),
			app.mattermostClient.AuthToken,
		)
		if err != nil {
			app.logger.Warn().Err(err).Msg("Mattermost websocket disconnected, retrying")
			failCount++

			// Экспоненциальный бэкофф с максимальным ограничением
			backoff := time.Duration(1<<failCount) * time.Second
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			app.logger.Info().Dur("retry_in", backoff).Msg("Retrying websocket connection")
			time.Sleep(backoff)
			continue
		}

		app.logger.Info().Msg("Mattermost websocket connected")
		failCount = 0
		app.mattermostWebSocketClient.Listen()

		for event := range app.mattermostWebSocketClient.EventChannel {
			go handleWebSocketEvent(app, event)
		}
	}
}

func handleWebSocketEvent(app *application, event *model.WebSocketEvent) {

	// Ignore other types of events.
	if event.EventType() != model.WebsocketEventPosted {
		return
	}

	// Since this event is a post, unmarshal it to (*model.Post)
	post := &model.Post{}
	err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
	if err != nil {
		app.logger.Error().Err(err).Msg("Could not cast event to *model.Post")
	}

	// Ignore messages sent by this bot itself.
	if post.UserId == app.mattermostUser.Id {
		return
	}

	// Handle however you want.
	handlePost(app, post)
}

func handlePost(app *application, post *model.Post) {
	app.logger.Debug().Str("message", post.Message).Msg("")
	app.logger.Debug().Interface("post", post).Msg("")

	// Проверяем, начинается ли сообщение с "/vote"
	if !regexp.MustCompile(`^/vote`).MatchString(post.Message) {
		return
	}

	// Обрабатываем команду "/vote create"
	if regexp.MustCompile(`^/vote create`).MatchString(post.Message) {
		handleVoteCreate(app, post)
		return
	}

	// Обрабатываем команду "/vote"
	if regexp.MustCompile(`^/vote send`).MatchString(post.Message) {
		handleVote(app, post)
		return
	}

	// Обрабатываем команду "/vote result"
	if regexp.MustCompile(`^/vote result`).MatchString(post.Message) {
		handleVoteResult(app, post)
		return
	}

	// Обрабатываем команду "/vote close"
	if regexp.MustCompile(`^/vote close`).MatchString(post.Message) {
		handleVoteClose(app, post)
		return
	}

	// Обрабатываем команду "/vote delete"
	if regexp.MustCompile(`^/vote delete`).MatchString(post.Message) {
		handleVoteDelete(app, post)
		return
	}

	// Обрабатываем команду "/vote help"
	if regexp.MustCompile(`^/vote help`).MatchString(post.Message) {
		handleVoteHelp(app, post)
		return
	}
}
