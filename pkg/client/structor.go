package client

import (
	"context"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"net/http"
)

const (
	exportTableAPI      = "/api/v1/structor/appid/internal/appbase/exportTable"
	exportPermissionAPI = "/api/v1/structor/appid/internal/appbase/exportPermission"
	importTableAPI      = "/api/v1/structor/appid/internal/appbase/importTable"
	importPermissionAPI = "/api/v1/structor/appid/internal/appbase/importPermission"

	removeTable = "/api/v1/structor/%s/base/recycle/app/removeTable"

	removePer = "/api/v1/structor/%s/base/recycle/app/removePer"
)

// NewStructor NewStructor
func NewStructor(conf *config.Config) Structor {
	return &structor{
		client: client.New(conf.InternalNet),
		conf:   conf,
	}
}

// Structor Structor
type Structor interface {
	ExportTable(ctx context.Context, appID string) (*ExportTableResp, error)
	ExportPermission(ctx context.Context, appID string) (*ExportPermissionResp, error)
	ImportTable(ctx context.Context, appID, oldAppID, jsons string, pageFiles map[string][]byte) (*ImportTableResp, error)
	ImportPermission(ctx context.Context, appID, jsons string, tableIDs map[string]string) (*ImportPermissionResp, error)
	Close()
}
type structor struct {
	client http.Client
	conf   *config.Config
}

// ExportTableResp ExportTableResp
type ExportTableResp struct {
	JSONs     string            `json:"exportJSON"`
	PageFiles map[string][]byte `json:"pageFiles"`
}

func (s *structor) ExportTable(ctx context.Context, appID string) (*ExportTableResp, error) {
	url := s.conf.StructorHost + exportTableAPI
	req := struct {
		AppID string `json:"appID"`
	}{
		AppID: appID,
	}
	resp := &ExportTableResp{}
	err := client.POST(ctx, &s.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ExportPermissionResp ExportPermissionResp
type ExportPermissionResp struct {
	JSONs string `json:"jsons"`
}

func (s *structor) ExportPermission(ctx context.Context, appID string) (*ExportPermissionResp, error) {
	url := s.conf.StructorHost + exportPermissionAPI
	req := struct {
		AppID string `json:"appID"`
	}{
		AppID: appID,
	}
	resp := &ExportPermissionResp{}
	err := client.POST(ctx, &s.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ImportTableResp ImportTableResp
type ImportTableResp struct {
	TableIDs map[string]string `json:"tableIDs"`
}

func (s *structor) ImportTable(ctx context.Context, appID, oldAppID, jsons string, pageFiles map[string][]byte) (*ImportTableResp, error) {
	url := s.conf.StructorHost + importTableAPI

	req := struct {
		JSONs     string            `json:"jsons"`
		AppID     string            `json:"appID"`
		PreAppID  string            `json:"preAppID"`
		PageFiles map[string][]byte `json:"pageFiles"`
	}{
		JSONs:     jsons,
		AppID:     appID,
		PageFiles: pageFiles,
		PreAppID:  oldAppID,
	}
	resp := &ImportTableResp{}
	err := client.POST(ctx, &s.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ImportPermissionResp ImportPermissionResp
type ImportPermissionResp struct {
}

func (s *structor) ImportPermission(ctx context.Context, appID, jsons string, tableIDs map[string]string) (*ImportPermissionResp, error) {
	url := s.conf.StructorHost + importPermissionAPI
	req := struct {
		JSONs    string            `json:"jsons"`
		AppID    string            `json:"appID"`
		TableIDs map[string]string `json:"tableIDs"`
	}{
		JSONs:    jsons,
		AppID:    appID,
		TableIDs: tableIDs,
	}
	resp := &ImportPermissionResp{}
	err := client.POST(ctx, &s.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *structor) Close() {
	s.client.CloseIdleConnections()
}
