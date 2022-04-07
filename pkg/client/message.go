package client

import (
	"context"
	"net/http"

	"github.com/quanxiang-cloud/cabin/tailormade/client"
)

const (
	messageURL = "http://message/api/v1/message/manager/create"
)

// MessageAPI MessageAPI
type MessageAPI interface {
	SendMs(ctx context.Context, userID, uuID string, contents []byte) (*SendMsResp, error)
	Close()
}

type message struct {
	client http.Client
}

func (m *message) Close() {
	m.client.CloseIdleConnections()
}

// NewMessage NewMessage
func NewMessage(config client.Config) MessageAPI {
	return &message{client: client.New(config)}
}

// CreateReq CreateReq
type CreateReq struct {
	data `json:",omitempty"`
}

type data struct {
	Letter *letter `json:"letter"`
}

type letter struct {
	ID      string   `json:"id,omitempty"`
	UUID    []string `json:"uuid,omitempty"`
	Content []byte   `json:"contents"`
}

// SendMsResp SendMsResp
type SendMsResp struct {
}

func (m *message) SendMs(ctx context.Context, userID, uuID string, contents []byte) (*SendMsResp, error) {
	req := new(CreateReq)
	req.Letter = &letter{
		ID:      userID,
		UUID:    []string{uuID},
		Content: contents,
	}
	resp := new(SendMsResp)
	err := client.POST(ctx, &m.client, messageURL, req, resp)
	return resp, err
}
