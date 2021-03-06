package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/whiteblock/genesis-cli/pkg/config"
	"github.com/whiteblock/genesis-cli/pkg/oauth2-noserver"
	"github.com/whiteblock/genesis-cli/pkg/util"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	conf         = config.NewConfig()
	globalClient *oauth2ns.AuthorizedClient

	mux     = &sync.Mutex{}
	oldFlag = false
)

func GetToken() *oauth2.Token {
	token := new(oauth2.Token)

	if len(conf.GenesisCredentials) != 0 {
		err := json.Unmarshal([]byte(conf.GenesisCredentials), token)
		if err == nil {
			return token
		}
	}
	if !conf.UserDir.Exists(conf.TokenFile) {
		log.Trace("the token file does not exist")
		return nil
	}
	data, err := conf.UserDir.ReadFile(conf.TokenFile)
	if err != nil {
		log.WithField("error", err).Debug("couldn't read token file")
		return nil
	}

	err = json.Unmarshal(data, token)
	if err != nil {
		log.WithField("error", err).Debug("couldn't parse token file")
		return nil
	}
	return token
}

func getClientFromLocalToken(authConf *oauth2.Config) *oauth2ns.AuthorizedClient {
	token := GetToken()
	if token == nil {
		return nil
	}

	return &oauth2ns.AuthorizedClient{
		Client: authConf.Client(context.Background(), token),
		Token:  token,
	}

}

func storeToken(client *oauth2ns.AuthorizedClient) error {

	data, err := json.Marshal(client.Token)
	if err != nil {
		return err
	}
	return conf.UserDir.WriteFile(conf.TokenFile, data)
}

func Login() (*oauth2ns.AuthorizedClient, error) {
	mux.Lock()
	oldFlag = true
	mux.Unlock()
	client, err := oauth2ns.AuthenticateUser(getAuthConf())
	if err != nil {
		return nil, err
	}
	return client, storeToken(client)
}

func GetClient() (*oauth2ns.AuthorizedClient, error) {
	mux.Lock()
	defer mux.Unlock()
	if globalClient != nil && !oldFlag {
		return globalClient, nil
	}
	oldFlag = false
	authConf := getAuthConf()

	var err error

	globalClient = getClientFromLocalToken(authConf)
	if globalClient != nil {
		return globalClient, nil
	}

	globalClient, err = oauth2ns.AuthenticateUser(authConf)
	if err != nil {
		return nil, err
	}
	err = storeToken(globalClient)
	if err != nil {
		util.Errorf("couldn't store token: %v", err)
	}

	return globalClient, nil

}

func getAuthConf() *oauth2.Config {
	return &oauth2.Config{
		ClientID: "cli",
		Scopes:   []string{"offline_access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s%s", conf.AuthEndpoint, conf.AuthPath),
			TokenURL: fmt.Sprintf("https://%s%s", conf.AuthEndpoint, conf.TokenPath),
		},
	}
}
