package restful

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/entrepot/internal/comet"
	"github.com/quanxiang-cloud/entrepot/internal/service"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"io/ioutil"
	"net/http"
	"regexp"
)

// BatchTask BatchTask
type BatchTask struct {
	batchTask service.BatchTask
}

const (
	_userID       = "User-Id"
	_userName     = "User-Name"
	_departmentID = "Department-Id"
)

// NewBatchTask NewBatchTask
func NewBatchTask(conf *config.Config, factor chan *comet.FactorData) (*BatchTask, error) {
	batchTask, err := service.NewBatchTask(conf, factor)
	if err != nil {
		return nil, err
	}
	return &BatchTask{
		batchTask: batchTask,
	}, nil
}

// CreatTask CreatTask
func (batch *BatchTask) CreatTask(c *gin.Context) {
	batchReq := &service.CreateTaskReq{
		UserID:   c.GetHeader(_userID),
		UserName: c.GetHeader(_userName),
		Command:  c.Param("command"),
	}
	ctx := header.MutateContext(c)
	if err := c.ShouldBind(batchReq); err != nil {
		logger.Logger.WithName("batchReq").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	value, err := json.Marshal(batchReq.Value)
	if err != nil {
		resp.Format(nil, error2.New(code.ErrParamFormat)).Context(c)
		return
	}
	if !checkName(batchReq.Title) || !checkName(string(value)) {
		resp.Format(nil, error2.New(code.ErrParamFormat)).Context(c)
		return
	}
	resp.Format(batch.batchTask.CreateTask(ctx, batchReq)).Context(c)

}

// GetList GetList
func (batch *BatchTask) GetList(c *gin.Context) {
	req := &service.GetListReq{}
	req.UserID = c.GetHeader(_userID)
	ctx := header.MutateContext(c)
	if err := c.ShouldBind(req); err != nil {
		logger.Logger.WithName("batchReq").Errorw(err.Error(), header.GetRequestIDKV(ctx).Fuzzy()...)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp.Format(batch.batchTask.GetList(ctx, req)).Context(c)

}

// GetByID GetByID
func (batch *BatchTask) GetByID(c *gin.Context) {
	req := &service.GetByIDReq{
		TaskID: c.Param("taskID"),
	}
	ctx := header.MutateContext(c)
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp.Format(batch.batchTask.GetByID(ctx, req)).Context(c)
}

// Subscribe Subscribe
func (batch *BatchTask) Subscribe(c *gin.Context) {
	req := new(service.SubscribeReq)
	ctx := header.MutateContext(c)
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	req.UserID = c.GetHeader(_userID)
	resp.Format(batch.batchTask.Subscribe(ctx, req)).Context(c)

}

// GetProcessing GetProcessing
func (batch *BatchTask) GetProcessing(c *gin.Context) {
	req := &service.GetProcessingReq{
		UserID: c.GetHeader(_userID),
	}
	ctx := header.MutateContext(c)
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp.Format(batch.batchTask.GetProcessing(ctx, req)).Context(c)

}

// Delete Delete
func (batch *BatchTask) Delete(c *gin.Context) {
	req := &service.DeleteReq{
		TaskID: c.Param("taskID"),
	}
	ctx := header.MutateContext(c)
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(batch.batchTask.Delete(ctx, req)).Context(c)
}

// Send Send
func (batch *BatchTask) Send(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		errHandle(c, err)
		return
	}
	event := new(service.DaprEvent)
	err = json.Unmarshal(body, event)
	if err != nil {
		errHandle(c, err)
		return
	}
	taskReq := event.Data
	logger.Logger.Infow("task Req", "data is", taskReq)
	ctx := header.MutateContext(c)
	_, err = batch.batchTask.CreateTask(ctx, taskReq)
	errHandle(c, err)
}

func errHandle(c *gin.Context, err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	c.JSON(http.StatusOK, nil)
}

func checkName(name string) bool {
	regexStr := "^[\u4e00-\u9fa5a-zA-Z0-9,./;'\\[\\]\\\\<>?:\"{}|`~!@#$%^&*()_+-=\\s\\n，。；‘’【】、《》？：“”{}·~！￥…（）]*$"
	result, err := regexp.MatchString(regexStr, name)
	if err != nil {
		return false
	}
	return result
}

func transformCTX(ctx context.Context, c *gin.Context) context.Context {
	var (
		_userID   interface{} = "User-Id"
		_userName interface{} = "User-Name"
		userID    string      = "User-Id"
		userName  string      = "User-Name"
	)
	ctx = context.WithValue(ctx, _userID, c.GetHeader(userID))
	ctx = context.WithValue(ctx, _userName, c.GetHeader(userName))
	return ctx
}
