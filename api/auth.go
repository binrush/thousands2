package main

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
	"os"
)

type AuthProviders map[string]*oauth2.Config


func GetAuthProviders() AuthProviders {
    providers := make(map[string]*oauth2.Config)
    providers["vk"] = &oauth2.Config{
	    RedirectURL:  "https://thousands.su/api/auth/authorized",
	    ClientID:     os.Getenv("VK_CLIENT_ID"),
	    ClientSecret: os.Getenv("VK_CLIENT_SECRET"),
	    Scopes:       []string{},
	    Endpoint:     vk.Endpoint,
    }
    return providers
}
