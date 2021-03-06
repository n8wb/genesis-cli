package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/whiteblock/genesis-cli/pkg/auth"

	log "github.com/sirupsen/logrus"
	"github.com/whiteblock/utility/common"
)

func GetStatus(testID string) (common.Status, error) {
	client, err := auth.GetClient()
	if err != nil {
		return common.Status{}, err
	}

	dest := conf.APIEndpoint() + fmt.Sprintf(conf.StatusURI, testID)
	log.WithField("url", dest).Trace("getting url")
	resp, err := client.Get(dest)
	if err != nil {
		return common.Status{}, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.Status{}, nil
	}
	var status common.Status
	err = json.Unmarshal(data, &status)
	if err != nil {
		return common.Status{}, fmt.Errorf(string(data))
	}
	return status, nil
}
