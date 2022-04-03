package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type config struct {
	SSHPort     int    `yaml:"ssh_port"`
	ProfilePort int    `yaml:"profile_port"`
	DataDir     string `yaml:"data_dir"`
	KeyFile     string `yaml:"key_file"`

	IntegrationConfig string `yaml:"integration_config"`
}

// integrations stores information needed by integrations.
// Code that uses this should check if fields are nil.
type integrations struct {
	// Twitter stores the information needed for the Twitter integration.
	// Check if it is enabled by checking if Twitter is nil.
	Twitter *twitterInfo `yaml:"twitter"`
	// Slack stores the information needed for the Slack integration.
	// Check if it is enabled by checking if Slack is nil.
	Slack *slackInfo `yaml:"slack"`
}

type twitterInfo struct {
	ConsumerKey       string `yaml:"consumer_key"`
	ConsumerSecret    string `yaml:"consumer_secret"`
	AccessToken       string `yaml:"access_token"`
	AccessTokenSecret string `yaml:"access_token_secret"`
}

type slackInfo struct {
	// Token is the Slack API token
	Token string `yaml:"token"`
	// Channel is the Slack channel to post to
	ChannelID string `yaml:"channel_id"`
	// Prefix is the prefix to prepend to messages from slack when rendered for SSH users
	Prefix string `yaml:"prefix"`
}

var (
	// TODO: use this config!!

	Config = config{ // first stores default config
		SSHPort:     2221,
		ProfilePort: 5555,
		DataDir:     "./devzat-data",
		KeyFile:     "./devzat-sshkey",

		IntegrationConfig: "",
	}

	Integrations = integrations{
		Twitter: nil,
		Slack:   nil,
	}
)

func init() {
	cfgFile := os.Getenv("DEVZAT_CONFIG")
	if cfgFile == "" {
		cfgFile = "devzat-config.yml"
	}

	errCheck := func(err error) {
		if err != nil {
			fmt.Println("err: " + err.Error())
			os.Exit(0) // match `return` behavior
		}
	}

	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Config file not found, so writing the default one to " + cfgFile)

			d, err := yaml.Marshal(Config)
			errCheck(err)
			err = os.WriteFile(cfgFile, d, 0644)
			errCheck(err)
			return
		}
		errCheck(err)
	}
	d, err := ioutil.ReadFile(cfgFile)
	errCheck(err)
	err = yaml.Unmarshal(d, &Config)
	errCheck(err)
	fmt.Println("Config loaded from " + cfgFile)

	if Config.IntegrationConfig != "" {
		d, err = ioutil.ReadFile(Config.IntegrationConfig)
		errCheck(err)
		err = yaml.Unmarshal(d, &Integrations)
		errCheck(err)

		if Integrations.Slack.Prefix == "" {
			Integrations.Slack.Prefix = "Slack"
		}
		if sl := Integrations.Slack; sl.Token == "" || sl.ChannelID == "" {
			fmt.Println("error: Slack token or Channel ID is missing")
			os.Exit(0)
		}
		if tw := Integrations.Twitter; tw.AccessToken == "" ||
			tw.AccessTokenSecret == "" ||
			tw.ConsumerKey == "" ||
			tw.ConsumerSecret == "" {
			fmt.Println("error: Twitter credentials are incomplete")
			os.Exit(0)
		}

		fmt.Println("Integration config loaded from " + Config.IntegrationConfig)

		if os.Getenv("DEVZAT_OFFLINE_SLACK") == "" {
			Integrations.Slack = nil
		}
		if os.Getenv("DEVZAT_OFFLINE_TWITTER") == "" {
			Integrations.Twitter = nil
		}
		// Check for global offline for backwards compatibility
		if os.Getenv("DEVZAT_OFFLINE") == "" {
			Integrations.Slack = nil
			Integrations.Twitter = nil
		}
	}
	slackInit()
	twitterInit()
}
