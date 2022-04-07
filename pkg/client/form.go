package client

import (
	"context"
	"fmt"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"net/http"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
)

type formAPI struct {
	client http.Client
	conf   *config.Config
}

func (f *formAPI) Close() {
	f.client.CloseIdleConnections()
}

// NewFormAPI NewFormAPI
func NewFormAPI(conf *config.Config) FormAPI {
	return &formAPI{
		client: client.New(conf.InternalNet),
		conf:   conf,
	}
}

// FormAPI FormAPI
type FormAPI interface {
	Search(ctx context.Context, options FindOptions, query interface{}, appID, tableID string) (*SearchResp, error)
	CreateBatch(ctx context.Context, appID, tableID string, req []*FormReq) (*CreateResp, error)
	Close()
}

// FindOptions page options
type FindOptions struct {
	Page int64    `json:"page"`
	Size int64    `json:"size"`
	Sort []string `json:"sort"`
}

// FormReq FormReq
type FormReq struct {
	FindOptions
	Query  interface{} `json:"query"`
	Entity interface{} `json:"entity"`
}

// SearchResp SearchResp
type SearchResp struct {
	Entities     []map[string]interface{} `json:"entities"`
	Total        int64                    `json:"total"`
	Aggregations interface{}              `json:"aggregations"`
}

/*
Search : use to search data
*/
func (f *formAPI) Search(ctx context.Context, options FindOptions, query interface{}, appID, tableID string) (*SearchResp, error) {
	path := fmt.Sprintf(f.conf.FormHost+homeHost+"/form/%s/search", appID, tableID)
	params := &FormReq{
		Query: query,
		FindOptions: FindOptions{
			Page: options.Page,
			Size: options.Size,
			Sort: options.Sort,
		},
	}
	resp := &SearchResp{}
	err := client.POST(ctx, &f.client, path, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// CreateResp CreateResp
type CreateResp struct {
	Entity     interface{} `json:"entity"`
	ErrorCount int64       `json:"errorCount"`
}

func (f *formAPI) CreateBatch(ctx context.Context, appID, tableID string, req []*FormReq) (*CreateResp, error) {
	path := fmt.Sprintf(f.conf.FormHost+homeHost+"/form/%s/create/batch", appID, tableID)
	resp := &CreateResp{}
	err := client.POST(ctx, &f.client, path, req, resp)

	if err != nil {
		return nil, err
	}
	return resp, nil
}
