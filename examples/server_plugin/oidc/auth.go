package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

// Authenticator is used to authenticate our users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

// New instantiates the *Authenticator.
func NewAuthenticator(conf config) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+conf.AUTH0_DOMAIN+"/",
	)
	if err != nil {
		return nil, err
	}

	oauth2Conf := oauth2.Config{
		ClientID:     conf.AUTH0_CLIENT_ID,
		ClientSecret: conf.AUTH0_CLIENT_SECRET,
		RedirectURL:  conf.AUTH0_CALLBACK_URL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   oauth2Conf,
	}, nil
}

// VerifyIDToken verifies that an *oauth2.Token is a valid *oidc.IDToken.
func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (a *Authenticator) DoAuth(ctx *gin.Context) error {
	state, err := generateRandomState()
	if err != nil {
		return err
	}

	// Save the state inside the session using helper functions
	sessionSet(ctx, "state", state)
	if err := sessionSave(ctx); err != nil {
		return err
	}

	ctx.Redirect(http.StatusTemporaryRedirect, a.AuthCodeURL(state))
	return nil
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	state := base64.StdEncoding.EncodeToString(b)

	return state, nil
}
