package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	e "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"io/ioutil"
	"net/http"
	"reflect"
)

func doPOST(ctx context.Context, client *http.Client, uri string, params interface{}, entity interface{}) error {
	if reflect.ValueOf(entity).Kind() != reflect.Ptr {
		return errors.New("the entity type must be a pointer")
	}

	paramByte, err := json.Marshal(params)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(paramByte)
	req, err := http.NewRequest("POST", uri, reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add(header.GetRequestIDKV(ctx).Wreck())
	req.Header.Add(header.GetTimezone(ctx).Wreck())
	req.Header.Add(header.GetTenantID(ctx).Wreck())
	req.Header.Add(getUserID(ctx).Wreck())
	req.Header.Add(getUserName(ctx).Wreck())

	response, err := client.Do(req)
	if err != nil {
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

func decomposeBody(body []byte, entity interface{}) error {
	r := new(resp.Resp)
	r.Data = entity

	err := json.Unmarshal(body, r)
	if err != nil {
		return err
	}

	if r.Code != e.Success {
		return r.Error
	}

	return nil
}

func getUserID(ctx context.Context) header.KV {
	_userID := "User-Id"
	i := ctx.Value(_userID)
	uid, ok := i.(string)
	if ok {
		return header.KV{_userID, uid}
	}
	return header.KV{_userID, "unexpected type"}
}

func getUserName(ctx context.Context) header.KV {
	_userName := "User-Name"
	i := ctx.Value(_userName)
	uName, ok := i.(string)
	if ok {
		return header.KV{_userName, uName}
	}
	return header.KV{_userName, "unexpected type"}
}
