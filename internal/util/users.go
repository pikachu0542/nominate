package util

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"time"

	cshAuth "github.com/computersciencehouse/csh-auth"
	"github.com/computersciencehouse/nominate/internal/config"
	"github.com/computersciencehouse/nominate/logging"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OIDCClient struct {
	oidcClientId     string
	oidcClientSecret string

	accessToken  string
	providerBase string
	quit         chan struct{}
}

type OIDCUser struct {
	Uuid     string `json:"id"`
	Username string `json:"username"`
	Gatekeep bool   `json:"result"`
	SlackUID string `json:"slackuid"`
}

var groupCache map[string]string

func (client *OIDCClient) setupOidcClient(oidcClientId, oidcClientSecret string) {
	client.oidcClientId = oidcClientId
	client.oidcClientSecret = oidcClientSecret
	parse, err := url.Parse(cshAuth.ProviderURI)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "setupOidcClient"}).Error(err)
		return
	}
	groupCache = make(map[string]string)
	client.providerBase = parse.Scheme + "://" + parse.Host
	exp := client.getAccessToken()
	ticker := time.NewTicker(time.Duration(exp) * time.Second)
	// this will async get the token
	go func() {
		for {
			select {
			case <-ticker.C:
				exp = client.getAccessToken()
				ticker.Reset(time.Duration(exp) * time.Second)
			case <-client.quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (client *OIDCClient) getAccessToken() int {
	htclient := http.DefaultClient
	//request body
	authData := url.Values{}
	authData.Set("client_id", client.oidcClientId)
	authData.Set("client_secret", client.oidcClientSecret)
	authData.Set("grant_type", "client_credentials")
	resp, err := htclient.PostForm(cshAuth.ProviderURI+"/protocol/openid-connect/token", authData)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "getAccessToken"}).Error(err)
		return 0
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logging.Logger.WithFields(logrus.Fields{"method": "getAccessToken", "statusCode": resp.StatusCode}).Error(resp.Status)
		return 0
	}
	respData := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "getAccessToken"}).Error(err)
		return 0
	}
	if respData["error"] != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "getAccessToken"}).Error(respData)
		return 0
	}
	client.accessToken = respData["access_token"].(string)
	return int(respData["expires_in"].(float64))
}

// Retreives a struct containing info about the CSH member
// Useful for checking LDAP groups
func GetUserData(c *gin.Context) cshAuth.CSHUserInfo {
	cl, _ := c.Get("cshauth")
	user := cl.(cshAuth.CSHClaims).UserInfo
	return user
}

// Determines if the user is an active member, allowing for a dev mode override
func IsActive(cfg *config.Config, user cshAuth.CSHUserInfo) bool {
	return cfg.DevFlags.DisableActiveFilters || slices.Contains(user.Groups, "active")
}

// Determines if the current user is on eboard, allowing for dev mode override
func IsEboard(cfg *config.Config, user cshAuth.CSHUserInfo) bool {
	return cfg.DevFlags.ForceEboard || slices.Contains(user.Groups, "eboard")
}

func IsChair(cfg *config.Config, user cshAuth.CSHUserInfo) bool {
	return cfg.DevFlags.ForceChair || slices.Contains(user.Groups, "eboard-chairman")
}
