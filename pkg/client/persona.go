package client

import (
	"context"
	"github.com/quanxiang-cloud/cabin/tailormade/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"net/http"
)

const (
	exportPersonaPath = "/api/v1/persona/app/export"
	importPersonaPath = "/api/v1/persona/app/import"
)

// Persona Persona
type Persona interface {
	Import(ctx context.Context, appData []*KV) (*ImportPersonaResp, error)
	Export(ctx context.Context, appID string) (*ExportPersonaResp, error)
}
type persona struct {
	client http.Client
	conf   *config.Config
}

// NewPersona return a persona client
func NewPersona(conf *config.Config) Persona {
	return &persona{
		client: client.New(conf.InternalNet),
		conf:   conf,
	}
}

// KV persona format
type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExportPersonaResp ExportPersonaResp
type ExportPersonaResp struct {
	AppData []*KV `json:"AppData"`
}

// Export export persona data.
func (p *persona) Export(ctx context.Context, appID string) (*ExportPersonaResp, error) {
	req := struct {
		AppID string `json:"appId"`
	}{
		AppID: appID,
	}
	url := p.conf.PersonaHost + exportPersonaPath
	resp := &ExportPersonaResp{}
	err := client.POST(ctx, &p.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// ImportPersonaResp ImportPersonaResp
type ImportPersonaResp struct {
}

func (p *persona) Import(ctx context.Context, appData []*KV) (*ImportPersonaResp, error) {
	req := struct {
		AppData []*KV `json:"appData"`
	}{
		AppData: appData,
	}
	url := p.conf.PersonaHost + importPersonaPath
	resp := &ImportPersonaResp{}
	err := client.POST(ctx, &p.client, url, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
