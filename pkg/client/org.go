package client

import (
	"context"
	"fmt"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"

	"io/ioutil"

	"net/http"
	"net/url"
)

const (
	host    = "http://org/api/v1/org"
	depTree = "/m/dep/tree"
	apiURL  = "http://search/api/v1/search/user"
)

type orgAPI struct {
	client http.Client
	conf   *config.Config
}

func (o *orgAPI) DepTree(ctx context.Context) (*TreeResp, error) {
	resp := &TreeResp{}
	err := Get(ctx, &o.client, host+depTree, resp, nil)
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
	DepTree(ctx context.Context) (*TreeResp, error)
	GetUserList(ctx context.Context, userName string) (*GetUserListResp, error)
	Close()
}

// TreeResp  树形返回结构
type TreeResp struct {
	ID             string     `json:"id,omitempty"`
	DepartmentName string     `json:"name,omitempty"`
	Child          []TreeResp `json:"child"`
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
	Total int64         `json:"total"`
	Users []*listVoResp `json:"users"`
}

//listVoResp
type listVoResp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (o *orgAPI) GetUserList(ctx context.Context, userName string) (*GetUserListResp, error) {
	resp := &GetUserListResp{}

	value := url.Values{}
	query := fmt.Sprintf(`{query(name:"%s"){users{id,name},total}}`, userName)
	value.Set("query", query)
	err := Get(ctx, &o.client, apiURL, resp, value)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

//Get Get
func Get(ctx context.Context, client *http.Client, urls string, entity interface{}, values url.Values) error {
	u, err := url.ParseRequestURI(urls)
	if err != nil {
		logger.Logger.Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		return err
	}
	if values != nil {
		u.RawQuery = values.Encode() // URL encode
		logger.Logger.WithName("get").Infow(u.String(), header.GetRequestIDKV(ctx).Fuzzy()...)
	}
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		logger.Logger.Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		return err
	}
	req.Header.Add(header.GetRequestIDKV(ctx).Wreck())
	req.Header.Add(header.GetTimezone(ctx).Wreck())
	req.Header.Add(header.GetTenantID(ctx).Wreck())
	response, err := client.Do(req)
	if err != nil {
		logger.Logger.Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("expected state value is 200, actually %d", response.StatusCode)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return decomposeBody(body, entity)
}
