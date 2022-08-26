package rmfetch

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/juruen/rmapi/config"
	"github.com/juruen/rmapi/model"
	"github.com/juruen/rmapi/transport"
	"time"
)

const (
	defaultDeviceDesc string = "desktop-linux"
)

var (
	ErrMissingCode       = errors.New("missing required one-time code")
	ErrDeviceTokenCreate = errors.New("failed to create device token from on-time code")
	ErrUserTokenCreate   = errors.New("failed to create user token from device token")
)

func authHttpCtx(oneTimeCode *string) (*transport.HttpClientCtx, error) {
	configPath := config.ConfigPath()
	authTokens := config.LoadTokens(configPath)
	httpClientCtx := transport.CreateHttpClientCtx(authTokens)

	if authTokens.DeviceToken == "" {
		if oneTimeCode == nil {
			return nil, ErrMissingCode
		}
		deviceToken, err := newDeviceToken(&httpClientCtx, *oneTimeCode)
		if err != nil {
			return nil, ErrDeviceTokenCreate
		}
		authTokens.DeviceToken = deviceToken
		httpClientCtx.Tokens.DeviceToken = deviceToken
		config.SaveTokens(configPath, authTokens)
	}

	userTokenExpired := false
	if authTokens.UserToken != "" {
		token, _, err := jwt.NewParser().ParseUnverified(authTokens.UserToken, jwt.MapClaims{})
		if err != nil {
			return nil, ErrUserTokenCreate
		}
		claims := token.Claims.(jwt.MapClaims)
		if val, ok := claims["exp"].(float64); ok {
			exp := time.Unix(int64(val), 0).UTC()
			if time.Now().UTC().Add(time.Hour).After(exp) {
				userTokenExpired = true
			}
		} else {
			return nil, ErrUserTokenCreate
		}
	}

	if userTokenExpired || authTokens.UserToken == "" {
		userToken, err := newUserToken(&httpClientCtx)
		if err != nil {
			return nil, ErrUserTokenCreate
		}
		authTokens.UserToken = userToken
		httpClientCtx.Tokens.UserToken = userToken
		config.SaveTokens(configPath, authTokens)
	}

	return &httpClientCtx, nil
}

func newDeviceToken(http *transport.HttpClientCtx, code string) (string, error) {
	id := uuid.NewString()
	req := model.DeviceTokenRequest{Code: code, DeviceDesc: defaultDeviceDesc, DeviceId: id}
	resp := transport.BodyString{}
	err := http.Post(transport.EmptyBearer, config.NewTokenDevice, req, &resp)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

func newUserToken(http *transport.HttpClientCtx) (string, error) {
	resp := transport.BodyString{}
	err := http.Post(transport.DeviceBearer, config.NewUserDevice, nil, &resp)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}
