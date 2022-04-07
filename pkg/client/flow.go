package client

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"net/http"
)

const (
	removeAPI     = "/api/v1/flow/deleteApp"
	exportFlowAPI = "/api/v1/flow/appReplicationExport"
	importFlowAPI = "/api/v1/flow/appReplicationImport"
)

// NewFlow NewFlow
func NewFlow(conf *config.Config) Flow {
	return &flow{
		client: client.New(conf.InternalNet),
		conf:   conf,
	}
}

// Flow Flow
type Flow interface {
	ExportFlow(ctx context.Context, appID string) (*ExportFlowResp, error)
	ImportFlow(ctx context.Context, appID, jsons string, tableIDs map[string]string) (*ImportFlowResp, error)
	Close()
}

type flow struct {
	client http.Client
	conf   *config.Config
}

// ExportFlowResp ExportFlowResp
type ExportFlowResp struct {
	Jsons string `json:"data"`
}

// ExportFlow ExportFlow
func (f *flow) ExportFlow(ctx context.Context, appID string) (*ExportFlowResp, error) {
	params := struct {
		AppID string `json:"appID"`
	}{
		AppID: appID,
	}

	data := ""
	err := client.POST(ctx, &f.client, f.conf.FlowHost+exportFlowAPI, params, &data)
	if err != nil {
		return nil, err
	}
	return &ExportFlowResp{
		Jsons: data,
	}, nil
}

// ImportFlowResp ImportFlowResp
type ImportFlowResp struct {
}

func (f *flow) ImportFlow(ctx context.Context, appID, jsons string, tableIDs map[string]string) (*ImportFlowResp, error) {
	params := struct {
		AppID    string            `json:"appID"`
		Jsons    string            `json:"flows"`
		TableIDs map[string]string `json:"formID"`
	}{
		AppID:    appID,
		Jsons:    jsons,
		TableIDs: tableIDs,
	}
	resp := &ImportFlowResp{}
	var x bool
	err := client.POST(ctx, &f.client, f.conf.FlowHost+importFlowAPI, params, &x)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (f *flow) Close() {
	f.client.CloseIdleConnections()
}
