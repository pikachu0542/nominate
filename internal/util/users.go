package util

import (
	"encoding/json"
	"io"
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
	Uuid      string `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Username  string `json:"username"`
	Gatekeep  bool   `json:"result"`
	SlackUID  string `json:"slackuid"`
}

var groupCache map[string]string

func NewOIDCClient(clientID, clientSecret string) *OIDCClient {
	client := &OIDCClient{quit: make(chan struct{})}
	client.setupOidcClient(clientID, clientSecret)
	return client
}

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

// Retreives the ID of a specific OIDC group
// Will be used to get the members of a group
func (client *OIDCClient) FindOIDCGroupID(name string) string {
	if groupCache[name] != "" {
		return groupCache[name]
	}
	htclient := &http.Client{}
	//active
	req, err := http.NewRequest("GET", client.providerBase+"/auth/admin/realms/csh/groups?exact=true&search="+name, nil)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "FindOIDCGroupID"}).Error(err)
		return ""
	}
	req.Header.Add("Authorization", "Bearer "+client.accessToken)
	resp, err := htclient.Do(req)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "FindOIDCGroupID"}).Error(err)
		return ""
	}
	defer resp.Body.Close()
	ret := make([]map[string]any, 0)
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "FindOIDCGroupID"}).Error(err)
		return ""
	}
	//Example:
	//[{"id":"47dd1a94-853c-426d-b181-6d0714074892","name":"eboard","path":"/eboard","subGroups":[{"id":"66b9578a-2b58-46a6-8040-59388e57e830","name":"eboard-opcomm","path":"/eboard/eboard-opcomm","subGroups":[]}]}]
	//it returns as an array for SOME reason, so we cut to the group we want
	group := ret[0]
	// and now we have SUBgroups, so we do this fucked parse
	// subGroups := group["subGroups"].([]any)
	// fmt.Println(subGroups)
	// subGroup := subGroups[0].(map[string]interface{})
	// fmt.Println(subGroup)
	gid := group["id"].(string)
	groupCache[name] = gid
	return gid

}

// Retreives a list of users in a specific OIDC group
func (client *OIDCClient) GetOIDCGroup(groupID string) []OIDCUser {
	htclient := &http.Client{}
	//active
	req, err := http.NewRequest("GET", client.providerBase+"/auth/admin/realms/csh/groups/"+groupID+"/members", nil)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "GetOIDCGroup"}).Error(err)
		return nil
	}
	req.Header.Add("Authorization", "Bearer "+client.accessToken)
	resp, err := htclient.Do(req)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "GetOIDCGroup"}).Error(err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logging.Logger.WithFields(logrus.Fields{
			"method":     "GetOIDCGroup",
			"statusCode": resp.StatusCode,
			"body":       string(body),
			"groupID":    groupID,
		}).Error("unexpected response from Keycloak")
		return nil
	}

	ret := make([]OIDCUser, 0)
	err = json.NewDecoder(resp.Body).Decode(&ret)
	if err != nil {
		logging.Logger.WithFields(logrus.Fields{"method": "GetOIDCGroup"}).Error(err)
		return nil
	}
	return ret
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
