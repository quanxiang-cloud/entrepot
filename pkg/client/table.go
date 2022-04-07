package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
)

const (
	homeHost = "/api/v1/form/%s/home"
)

// TableAPI table Client interface
type TableAPI interface {
	HomeTableSchema(ctx context.Context, appID, tableID string) (*GetTableSchemaResp, error)
	Close()
}

type tableAPI struct {
	client http.Client
	conf   *config.Config
}

// NewTableAPI return a table client
func NewTableAPI(conf *config.Config) TableAPI {
	return &tableAPI{
		conf:   conf,
		client: client.New(conf.InternalNet),
	}
}

// GetTableSchemaResp GetTableSchemaResp
type GetTableSchemaResp struct {
	ID      string                 `json:"id"`
	TableID string                 `json:"tableID"`
	Schema  map[string]interface{} `json:"schema"`
	Config  map[string]interface{} `json:"config"`
}

// GetTableSchemaReq GetTableSchemaReq
type GetTableSchemaReq struct {
	TableID string `json:"tableID"`
}

func (t *tableAPI) HomeTableSchema(ctx context.Context, appID, tableID string) (*GetTableSchemaResp, error) {
	params := &GetTableSchemaReq{
		TableID: tableID,
	}
	resp := new(GetTableSchemaResp)
	path := fmt.Sprintf(t.conf.FormHost+homeHost+"/schema/%s", appID, tableID)
	err := client.POST(ctx, &t.client, path, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (t *tableAPI) Close() {
	t.client.CloseIdleConnections()
}
