package main

import (
	"net/url"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type config struct {
	mattermostUserName string
	mattermostTeamName string
	mattermostToken    string
	mattermostServer   *url.URL
	tarantoolUser      string
	tarantoolPass      string
}

func loadConfig() config {
	var settings config

	settings.mattermostTeamName = os.Getenv("MM_TEAM")
	settings.mattermostUserName = os.Getenv("MM_USERNAME")
	settings.mattermostToken = os.Getenv("MM_TOKEN")
	settings.mattermostServer, _ = url.Parse(os.Getenv("MM_SERVER"))
	settings.tarantoolUser = os.Getenv("TT_USER")
	settings.tarantoolPass = os.Getenv("TT_PASS")

	return settings
}
