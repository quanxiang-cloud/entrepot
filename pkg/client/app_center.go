package client

import (
	"context"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"net/http"
)

const (
	exportAppURL      = "/api/v1/app-center/exportApp"
	checkVersionURL   = "/api/v1/app-center/checkVersion"
	failImportURL     = "/api/v1/app-center/failImport"
	successImportURL  = "/api/v1/app-center/successImport"
	createTemplateURL = "/api/v1/app-center/template/create"
	getTemplateURL    = "/api/v1/app-center/template/getOne"
	finishTemplateURL = "/api/v1/app-center/template/finish"
	deleteTemplateURL = "/api/v1/app-center/template/delete"
)

// NewAppCenter AppCenterClient
func NewAppCenter(conf *config.Config) AppCenter {
	return &appCenter{
		client: client.New(conf.InternalNet),
		conf:   conf,
	}
}

// AppCenter AppCenter
type AppCenter interface {
	ExportAppInfo(ctx context.Context, appID string) (*ExportAppInfoResp, error)
	CheckVersion(ctx context.Context, version string) error
	SuccessImport(ctx context.Context, appID string) error
	FailImport(ctx context.Context, appID string) error
	CreateTemplate(ctx context.Context, req *CreateTemplateReq) (*CreateTemplateResp, error)
	GetTemplateByID(ctx context.Context, id string) (*GetTemplateByIDResp, error)
	FinishTemplate(ctx context.Context, id, path string) (*FinishTemplateResp, error)
	DeleteTemplate(ctx context.Context, id string) (*DeleteTemplateResp, error)
	Close()
}
type appCenter struct {
	client http.Client
	conf   *config.Config
}

// ExportAppInfoResp ExportAppInfoResp
type ExportAppInfoResp struct {
	AppID   string `json:"appID"`
	AppName string `json:"appName"`
	Version string `json:"version"`
}

func (a *appCenter) ExportAppInfo(ctx context.Context, appID string) (*ExportAppInfoResp, error) {
	req := struct {
		AppID string `json:"appID"`
	}{
		AppID: appID,
	}
	url := a.conf.AppCenterHost + exportAppURL
	resp := &ExportAppInfoResp{}
	err := client.POST(ctx, &a.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *appCenter) CheckVersion(ctx context.Context, version string) error {
	req := struct {
		Version string `json:"version"`
	}{
		Version: version,
	}
	url := a.conf.AppCenterHost + checkVersionURL
	resp := map[string]interface{}{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return err
	}
	return nil
}

func (a *appCenter) FailImport(ctx context.Context, appID string) error {
	req := struct {
		AppID string `json:"appID"`
	}{
		AppID: appID,
	}
	url := a.conf.AppCenterHost + failImportURL
	resp := map[string]interface{}{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return err
	}
	return nil
}

func (a *appCenter) SuccessImport(ctx context.Context, appID string) error {
	req := struct {
		AppID string `json:"appID"`
	}{
		AppID: appID,
	}
	url := a.conf.AppCenterHost + successImportURL
	resp := map[string]interface{}{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return err
	}
	return nil
}

// CreateTemplateReq CreateTemplateReq
type CreateTemplateReq struct {
	Name    string `json:"name"`
	AppIcon string `json:"appIcon"`
	AppID   string `json:"appID"`
	Version string `json:"version"`
	GroupID string `json:"groupID"`
	Path    string `json:"path"`
}

// CreateTemplateResp CreateTemplateResp
type CreateTemplateResp struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	AppIcon string `json:"appIcon"`
	AppID   string `json:"appID"`
	AppName string `json:"appName"`
	GroupID string `json:"groupID"`
	Status  int    `json:"status"`
}

func (a *appCenter) CreateTemplate(ctx context.Context, req *CreateTemplateReq) (*CreateTemplateResp, error) {
	url := a.conf.AppCenterHost + createTemplateURL
	resp := &CreateTemplateResp{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetTemplateByIDResp GetTemplateByIDResp
type GetTemplateByIDResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppIcon     string `json:"appIcon"`
	Path        string `json:"path"`
	AppID       string `json:"appID"`
	AppName     string `json:"appName"`
	GroupID     string `json:"groupID"`
	CreatedBy   string `json:"createdBy"`
	CreatedName string `json:"createdName"`
	CreatedTime int64  `json:"createdTime"`
	UpdatedBy   string `json:"updatedBy"`
	UpdatedName string `json:"updatedName"`
	UpdatedTime int64  `json:"updatedTime"`
	Status      int    `json:"status"`
}

func (a *appCenter) GetTemplateByID(ctx context.Context, id string) (*GetTemplateByIDResp, error) {
	req := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	url := a.conf.AppCenterHost + getTemplateURL
	resp := &GetTemplateByIDResp{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FinishTemplateResp FinishTemplateResp
type FinishTemplateResp struct {
}

func (a *appCenter) FinishTemplate(ctx context.Context, id, path string) (*FinishTemplateResp, error) {
	req := struct {
		ID   string `json:"id"`
		Path string `json:"path"`
	}{
		ID:   id,
		Path: path,
	}
	url := a.conf.AppCenterHost + finishTemplateURL
	resp := &FinishTemplateResp{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// DeleteTemplateResp DeleteTemplateResp
type DeleteTemplateResp struct {
}

func (a *appCenter) DeleteTemplate(ctx context.Context, id string) (*DeleteTemplateResp, error) {
	req := struct {
		ID string `json:"id"`
	}{
		ID: id,
	}
	url := a.conf.AppCenterHost + deleteTemplateURL
	resp := &DeleteTemplateResp{}
	err := client.POST(ctx, &a.client, url, req, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *appCenter) Close() {
	a.client.CloseIdleConnections()
}
