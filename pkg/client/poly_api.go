package client

import (
	"context"
	"fmt"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"net/http"
)

const (
	exportAPI = "/api/v1/polyapi/inner/exportApp/%s"
	importAPI = "/api/v1/polyapi/inner/import"
)

// NewPolyAPI NewPolyAPI
func NewPolyAPI(conf *config.Config) PolyAPI {
	return &polyapi{
		client: client.New(conf.InternalNet),
		conf:   conf,
	}
}

type polyapi struct {
	client http.Client
	conf   *config.Config
}

// ExportAPIResp ExportAPIResp
type ExportAPIResp struct {
	Path string `json:"path"`
	Data string `json:"data"`
}

// ExportAPI ExportAPI
func (p *polyapi) ExportAPI(ctx context.Context, appID string) (*ExportAPIResp, error) {
	reqURL := p.conf.PolyAPIHost + fmt.Sprintf(exportAPI, appID)
	resp := &ExportAPIResp{}
	err := client.POST(ctx, &p.client, reqURL, nil, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ImportAPIResp ImportAPIResp
type ImportAPIResp struct {
}

// ImportAPI ImportAPI
func (p *polyapi) ImportAPI(ctx context.Context, oldAppID, newAppID, data string) (*ImportAPIResp, error) {
	req := struct {
		OldID string `json:"oldID"`
		NewID string `json:"newID"`
		Data  string `json:"data"`
	}{
		OldID: oldAppID,
		NewID: newAppID,
		Data:  data,
	}
	resp := &ImportAPIResp{}
	err := client.POST(ctx, &p.client, p.conf.PolyAPIHost+importAPI, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *polyapi) Close() {
	p.client.CloseIdleConnections()
}

// PolyAPI PolyAPI
type PolyAPI interface {
	Close()
	ExportAPI(ctx context.Context, appID string) (*ExportAPIResp, error)
	ImportAPI(ctx context.Context, oldAppID, newAppID, data string) (*ImportAPIResp, error)
}
