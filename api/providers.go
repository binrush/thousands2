package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const (
	VKAPIEndpoint = "https://api.vk.com/method/users.get"
)

type provider interface {
	GetSrcId() int
	GetConfig() *oauth2.Config
	GetUserId(token *oauth2.Token) (string, error)
	Register(token *oauth2.Token, db *Database, ctx context.Context) (int64, error)
}

type AuthProviders map[string]provider

type VKProvider struct {
	config *oauth2.Config
}

func (provider *VKProvider) GetConfig() *oauth2.Config {
	return provider.config
}

func (provider *VKProvider) GetSrcId() int {
	return 1
}

func (provider *VKProvider) GetUserId(token *oauth2.Token) (string, error) {
	userId := token.Extra("user_id")
	if userId == nil {
		return "", fmt.Errorf("Failed to get VK user Id")
	}
	return userId.(string), nil
}

func (provider *VKProvider) Register(token *oauth2.Token, db *Database, ctx context.Context) (int64, error) {
	oauthClient := provider.GetConfig().Client(ctx, token)
	req, err := http.NewRequest("GET", VKAPIEndpoint, nil)
	if err != nil {
		return 0, err
	}
	query := req.URL.Query()
	query.Add("v", "5.131")
	query.Add("lang", "ru")
	query.Add("fields", "photo_50, photo_200_orig, has_photo")
	req.URL.RawQuery = query.Encode()
	resp, err := oauthClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var vkResponse VKUserGetResponse
	err = json.Unmarshal(content, &vkResponse)
	if err != nil {
		return 0, err
	}
	userData := vkResponse.Response
	var userId int64
	userId, err = CreateUser(
		db,
		fmt.Sprintf("%s %s", userData.FirstName, userData.LastName),
		strconv.Itoa(userData.Id), provider.GetSrcId())
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func NewVKProvider(baseUrl, client_id, client_secret string) provider {
	var vkProvider VKProvider
	vkProvider.config = &oauth2.Config{
		RedirectURL:  baseUrl + "/auth/authorized/vk",
		ClientID:     client_id,
		ClientSecret: client_secret,
		Endpoint:     vk.Endpoint,
	}
	return &vkProvider
}

func GetAuthProviders(baseUrl string) AuthProviders {
	providers := make(AuthProviders)
	providers["vk"] = NewVKProvider(baseUrl, os.Getenv("VK_CLIENT_ID"), os.Getenv("VK_CLIENT_SECRET"))
	return providers
}
