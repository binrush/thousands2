package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/vk"
)

const (
	VKApiBaseUrl = "https://api.vk.com"
	SUApiBaseUrl = "https://www.southural.ru/oauth2/"
	AuthSrcVK    = 1
	AuthSrcSU    = 2
)

type Provider interface {
	GetSrcId() int
	GetConfig() *oauth2.Config
	GetUserId(token *oauth2.Token) (string, error)
	Register(token *oauth2.Token, storage *Storage, ctx context.Context) (int64, error)
}

type AuthProviders map[string]Provider

type VKUser struct {
	Id           int    `json:"id"`
	Photo200Orig string `json:"photo_200_orig"`
	Photo50      string `json:"photo_50"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	HasPhoto     int    `json:"has_photo"`
}

type VKAPiError struct {
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

type VKUserGetResponse struct {
	Response []*VKUser   `json:"response"`
	Error    *VKAPiError `json:"error"`
}

type VKProvider struct {
	config  *oauth2.Config
	BaseUrl string
}

func (provider *VKProvider) GetConfig() *oauth2.Config {
	return provider.config
}

func (provider *VKProvider) GetSrcId() int {
	return 1
}

func (provider *VKProvider) GetUserId(token *oauth2.Token) (string, error) {
	userIdField := token.Extra("user_id")
	if userIdField == nil {
		return "", fmt.Errorf("failed to get VK user Id")
	}
	return strconv.FormatInt(int64(userIdField.(float64)), 10), nil
}

func (provider *VKProvider) Register(token *oauth2.Token, storage *Storage, ctx context.Context) (int64, error) {
	oauthClient := provider.GetConfig().Client(ctx, token)
	req, err := http.NewRequest("GET", provider.BaseUrl+"/method/users.get", nil)
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
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var vkResponse VKUserGetResponse
	err = json.Unmarshal(content, &vkResponse)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response from VK (%s): %v", content, err)
	}
	if vkResponse.Error != nil {
		// Handle error response from VK API
		return 0, fmt.Errorf("VK API error %d: %s", vkResponse.Error.ErrorCode, vkResponse.Error.ErrorMsg)
	}
	userData := vkResponse.Response[0]
	var userId int64
	userId, err = storage.CreateUser(
		fmt.Sprintf("%s %s", userData.FirstName, userData.LastName),
		strconv.Itoa(userData.Id), provider.GetSrcId())
	if err != nil {
		return 0, err
	}
	// load images. If download failed, just log it and proceed
	if userData.HasPhoto > 0 {
		// download main image
		for _, img := range []struct {
			url  string
			size string
		}{
			{userData.Photo50, ImageSmall},
			{userData.Photo200Orig, ImageMedium},
		} {
			_, err := downloadImage(*oauthClient, img.url)
			if err != nil {
				log.Printf("Failed to load image for user %d: %v", userId, err)
			}
			imageKey := fmt.Sprintf("users/%d_%s.jpg", userId, img.size)
			// TODO: upload image to S3
			err = storage.UpdateUserImage(userId, img.size, imageKey)
			if err != nil {
				log.Printf("Failed to store image for user %d: %v", userId, err)
			}
		}
	}
	return userId, nil
}

func downloadImage(client http.Client, url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image %s: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image %s: unexpected status %d", url, resp.StatusCode)
	}
	img, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image %s: %v", url, err)
	}
	return img, nil
	/*
		err = UpdateUserImage(db, userId, size, img)
		if err != nil {
			return fmt.Errorf("failed to store image %s in database: %v", url, err)
		}
		return nil
	*/
}

type SUAuthProvider struct {
	config *oauth2.Config
}

func (p *SUAuthProvider) GetConfig() *oauth2.Config {
	return p.config
}

func (p *SUAuthProvider) GetSrcId() int {
	return AuthSrcSU
}

func (p *SUAuthProvider) fetchUserInfo(token *oauth2.Token, ctx context.Context) (map[string]interface{}, error) {
	oauthClient := p.GetConfig().Client(ctx, token)
	resp, err := oauthClient.Get(SUApiBaseUrl + "UserInfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info from SU: %v", err)
	}
	defer resp.Body.Close()
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode SU user info: %v", err)
	}
	if errVal, ok := data["error"]; ok {
		return nil, fmt.Errorf("error in SU user info: %v", errVal)
	}
	return data, nil
}

func (p *SUAuthProvider) GetUserId(token *oauth2.Token) (string, error) {
	data, err := p.fetchUserInfo(token, context.Background())
	if err != nil {
		return "", err
	}
	sub, ok := data["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing 'sub' in SU user info")
	}
	return sub, nil
}

func (p *SUAuthProvider) Register(token *oauth2.Token, storage *Storage, ctx context.Context) (int64, error) {
	data, err := p.fetchUserInfo(token, ctx)
	if err != nil {
		return 0, err
	}
	sub, ok := data["sub"].(string)
	if !ok {
		return 0, fmt.Errorf("missing 'sub' in SU user info")
	}
	name, _ := data["name"].(string)
	userId, err := storage.CreateUser(name, sub, p.GetSrcId())
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func GetAuthProviders(baseUrl string) AuthProviders {
	providers := make(AuthProviders)
	providers["vk"] = &VKProvider{
		&oauth2.Config{
			RedirectURL:  baseUrl + "/auth/authorized/vk",
			ClientID:     os.Getenv("VK_CLIENT_ID"),
			ClientSecret: os.Getenv("VK_CLIENT_SECRET"),
			Endpoint:     vk.Endpoint,
		},
		VKApiBaseUrl,
	}
	providers["su"] = &SUAuthProvider{
		&oauth2.Config{
			RedirectURL:  baseUrl + "/auth/authorized/su",
			ClientID:     os.Getenv("SU_CLIENT_ID"),
			ClientSecret: os.Getenv("SU_CLIENT_SECRET"),
			Endpoint: oauth2.Endpoint{
				AuthURL:  SUApiBaseUrl + "authorize",
				TokenURL: SUApiBaseUrl + "token",
			},
			Scopes: []string{"openid", "profile"},
		},
	}
	return providers
}
