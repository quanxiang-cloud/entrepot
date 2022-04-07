package client

import (
	"context"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"net/http"
)

const (
	host         = "http://org/api/v1/org"
	adminDepList = "/adminDEPList"
	getUserInfo  = "/otherGetUserList"
	depTree      = "/DEPTree"
)

type orgAPI struct {
	client http.Client
	conf   *config.Config
}

func (o *orgAPI) DepTree(ctx context.Context) (*TreeResp, error) {
	params := struct{}{}
	resp := &TreeResp{}
	err := client.POST(ctx, &o.client, host+depTree, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (o *orgAPI) Close() {
	o.client.CloseIdleConnections()
}

// OrgAPI OrgAPI
type OrgAPI interface {
	GetUserList(ctx context.Context, userName string) ([]GetUserListResp, error)
	GetDepList(ctx context.Context, depName string) (*GetDepListResp, error)
	DepTree(ctx context.Context) (*TreeResp, error)
	Close()
}

// TreeResp  TreeResp
type TreeResp struct {
	ID                 string     `json:"id,omitempty"`
	DepartmentName     string     `json:"departmentName,omitempty"`
	DepartmentLeaderID string     `json:"departmentLeaderID,omitempty"`
	UseStatus          int        `json:"useStatus,omitempty"`
	PID                string     `json:"pid,omitempty"`
	SuperPID           string     `json:"superID,omitempty"`
	CompanyID          string     `json:"companyID,omitempty"`
	Grade              int        `json:"grade,omitempty"`
	Child              []TreeResp `json:"child"`
}

// NewOrgAPI NewOrgAPI
func NewOrgAPI(conf *config.Config) OrgAPI {
	return &orgAPI{
		conf:   conf,
		client: client.New(conf.InternalNet),
	}
}

// GetDepListResp GetDepListResp
type GetDepListResp struct {
	PageSize    int                `json:"-"`
	TotalCount  int64              `json:"total_count"`
	TotalPage   int                `json:"-"`
	CurrentPage int                `json:"-"`
	StartIndex  int                `json:"-"`
	Data        []*AdminDepartment `json:"data"`
}

// AdminDepartment AdminDepartment
type AdminDepartment struct {
	ID             string `json:"id,omitempty"`
	DepartmentName string `json:"departmentName,omitempty"`
}

// GetUserListResp GetUserListResp
type GetUserListResp struct {
	ID       string `json:"id,omitempty"`
	UserName string `json:"userName,omitempty"`
}

func (o *orgAPI) GetDepList(ctx context.Context, depName string) (*GetDepListResp, error) {
	params := struct {
		DepartmentName string `json:"departmentName"`
	}{
		DepartmentName: depName,
	}
	resp := &GetDepListResp{}
	err := client.POST(ctx, &o.client, host+adminDepList, params, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (o *orgAPI) GetUserList(ctx context.Context, userName string) ([]GetUserListResp, error) {
	params := struct {
		UserName string `json:"userName"`
	}{
		UserName: userName,
	}

	var batch []GetUserListResp
	err := client.POST(ctx, &o.client, host+getUserInfo, params, &batch)
	if err != nil {
		return nil, err
	}
	return batch, nil
}
