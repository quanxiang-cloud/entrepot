package restful

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	error2 "github.com/quanxiang-cloud/cabin/error"
	header2 "github.com/quanxiang-cloud/cabin/tailormade/header"
	"github.com/quanxiang-cloud/cabin/tailormade/resp"
	"github.com/quanxiang-cloud/entrepot/internal/comet"
	"github.com/quanxiang-cloud/entrepot/internal/service"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
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
	if err := c.ShouldBind(batchReq); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
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
	resp.Format(batch.batchTask.CreateTask(header2.MutateContext(c), batchReq)).Context(c)

}

// GetList GetList
func (batch *BatchTask) GetList(c *gin.Context) {
	req := &service.GetListReq{}
	req.UserID = c.GetHeader(_userID)
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(batch.batchTask.GetList(header2.MutateContext(c), req)).Context(c)

}

// GetByID GetByID
func (batch *BatchTask) GetByID(c *gin.Context) {
	req := &service.GetByIDReq{
		TaskID: c.Param("taskID"),
	}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(batch.batchTask.GetByID(header2.MutateContext(c), req)).Context(c)
}

// Subscribe Subscribe
func (batch *BatchTask) Subscribe(c *gin.Context) {
	req := new(service.SubscribeReq)

	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	req.UserID = c.GetHeader(_userID)
	resp.Format(batch.batchTask.Subscribe(header2.MutateContext(c), req)).Context(c)

}

// GetProcessing GetProcessing
func (batch *BatchTask) GetProcessing(c *gin.Context) {
	req := &service.GetProcessingReq{
		UserID: c.GetHeader(_userID),
	}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	resp.Format(batch.batchTask.GetProcessing(header2.MutateContext(c), req)).Context(c)

}

// Delete Delete
func (batch *BatchTask) Delete(c *gin.Context) {
	req := &service.DeleteReq{
		TaskID: c.Param("taskID"),
	}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	resp.Format(batch.batchTask.Delete(header2.MutateContext(c), req)).Context(c)
}

func checkName(name string) bool {
	regexStr := "^[\u4e00-\u9fa5a-zA-Z0-9,./;'\\[\\]\\\\<>?:\"{}|`~!@#$%^&*()_+-=\\s\\n，。；‘’【】、《》？：“”{}·~！￥…（）]*$"
	result, err := regexp.MatchString(regexStr, name)
	if err != nil {
		return false
	}
	return result
}
