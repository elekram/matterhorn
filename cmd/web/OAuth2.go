package main

import (
	appcfg "github.com/elekram/matterhorn/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func newOAuthConfig(c *appcfg.ConfigProperties) *oauth2.Config {
	// Credentials need to be created using Google's dev console
	// Developer Console: (https://console.developers.google.com)
	host := c.Host
	domain := c.Domain

	conf := oauth2.Config{
		ClientID:     c.OAuthClientId,
		ClientSecret: c.OAuthSecret,
		RedirectURL:  "https://" + host + "." + domain + "/oauth2/redirect/google",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	return &conf
}
