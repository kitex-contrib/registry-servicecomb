package servicecomb

import (
	"github.com/go-chassis/sc-client"
	"strconv"
)

func NewDefaultSCClient() (*sc.Client, error) {
	ep := SCAddr() + ":" + strconv.FormatInt(SCPort(), 10)
	client, err := sc.NewClient(sc.Options{
		Endpoints: []string{ep},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}
