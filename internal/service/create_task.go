package service

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/comet"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
)

// CreateTaskReq CreateTaskReq
type CreateTaskReq struct {
	ID       string  `json:"taskID"`
	UserID   string  `json:"userID"`
	UserName string  `json:"userName"`
	DepID    string  `json:"depID"`
	Value    logic.M `json:"value" binding:"required"`
	Addr     string  `json:"addr"`
	Opt      string  `json:"opt"`
	Size     int     `json:"size"`
	Command  string  `json:"command" binding:"required"`
	Title    string  `json:"title"`
}

// CreateTaskResp CreateTaskResp
type CreateTaskResp struct {
	TaskID string `json:"taskID"`
}

func (t *task) CreateTask(ctx context.Context, req *CreateTaskReq) (*CreateTaskResp, error) {
	// create task
	task1 := &models.Task{
		ID:          req.ID,
		FileAddr:    req.Addr,
		FileOpt:     req.Opt,
		FileSize:    req.Size,
		CreatorID:   req.UserID,
		CreatorName: req.UserName,
		Command:     req.Command,
		Title:       req.Title,
		Value:       getValue(req.Value),
		Status:      models.TaskDoing,
		DepID:       req.DepID,
	}
	task1.Types = managerTypes
	_, ok := logic.HomeMap[logic.Command(req.Command)]
	if ok {
		task1.Types = homeTypes
	}
	err := t.task.Create(t.db, task1)
	if err != nil {
		return nil, err
	}
	factorData := &comet.FactorData{
		Command: req.Command,
		Task:    task1,
		Ctx:     ctx,
	}
	// 发到channel
	t.factor <- factorData
	return &CreateTaskResp{
		//TaskID: task1.ID,
	}, nil

}
