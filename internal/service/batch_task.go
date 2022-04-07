package service

import (
	"context"
	error2 "github.com/quanxiang-cloud/cabin/error"
	id2 "github.com/quanxiang-cloud/cabin/id"
	"github.com/quanxiang-cloud/cabin/logger"
	mysql2 "github.com/quanxiang-cloud/cabin/tailormade/db/mysql"
	redis2 "github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	"github.com/quanxiang-cloud/entrepot/internal/comet"
	factor2 "github.com/quanxiang-cloud/entrepot/internal/comet/factor"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models/mysql"
	"github.com/quanxiang-cloud/entrepot/internal/models/redis"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"gorm.io/gorm"
	"time"

	"github.com/quanxiang-cloud/entrepot/internal/models"
)

const (
	homeTypes    = "home"
	managerTypes = "manager"
)

var (
	ttl = time.Duration(15) * time.Minute // 15 minute
)

// BatchTask BatchTask
type BatchTask interface {
	CreateTask(ctx context.Context, req *CreateTaskReq) (*CreateTaskResp, error)
	GetList(ctx context.Context, req *GetListReq) (*GetListResp, error)
	GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error)
	Subscribe(ctx context.Context, req *SubscribeReq) (*SubscribeResp, error)
	GetProcessing(ctx context.Context, req *GetProcessingReq) (*GetProcessingResp, error)
	Delete(ctx context.Context, req *DeleteReq) (*DeleteResp, error)
}

type task struct {
	factor chan *comet.FactorData
	task   models.TaskRepo
	db     *gorm.DB
	mold   *factor2.Molders
	conf   *config.Config
	ps     models.PubSubRepo
}

//NewBatchTask NewBatchTask
func NewBatchTask(conf *config.Config, factor chan *comet.FactorData) (BatchTask, error) {
	db, err := mysql2.New(conf.Mysql, logger.Logger)
	if err != nil {
		return nil, err
	}
	redisClient, err := redis2.NewClient(conf.Redis)
	if err != nil {
		return nil, err
	}
	return &task{
		conf:   conf,
		factor: factor,
		db:     db,
		mold:   factor2.GetMolds(),
		task:   mysql.NewTaskRepo(),
		ps:     redis.NewPubSub(redisClient),
	}, nil
}

// CreateTaskReq CreateTaskReq
type CreateTaskReq struct {
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
		ID:          id2.StringUUID(),
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
	tasks, err := t.task.GetByCondition(t.db, task1)
	if err != nil {
		return nil, err
	}
	if tasks != nil {
		return nil, error2.New(code.ErrNotRepeat)
	}
	task1.Types = managerTypes
	_, ok := logic.HomeMap[logic.Command(req.Command)]
	if ok {
		task1.Types = homeTypes
	}
	molders, ok := t.mold.GetMolders(ctx, logic.Command(req.Command), t.conf)
	if !ok {
		return nil, error2.New(code.ErrInternalError)
	}
	err = molders.SetTaskTitle(ctx, task1)
	if err != nil {
		return nil, err
	}

	err = t.task.Create(t.db, task1)
	if err != nil {
		return nil, err
	}
	factorData := &comet.FactorData{
		Command: req.Command,
		Task:    task1,
		Ctx:     ctx,
	}
	// send to channel
	t.factor <- factorData

	return &CreateTaskResp{
		TaskID: task1.ID,
	}, nil

}

func getValue(m logic.M) models.M {
	resp := make(models.M)
	for key, value := range m {
		resp[key] = value
	}
	return resp
}

// GetListReq GetListReq
type GetListReq struct {
	UserID string            `json:"userID"`
	Status models.TaskStatus `json:"status"`
	Page   int               `json:"page"`
	Limit  int               `json:"limit"`
	Types  string            `json:"types"`
}

// GetListResp GetListResp
type GetListResp struct {
	List  []*TaskListVo `json:"list"`
	Total int64         `json:"total"`
}

// TaskListVo TaskListVo
type TaskListVo struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Status    models.TaskStatus `json:"status"`
	Result    logic.M           `json:"result"`
	Command   string            `json:"command"`
	Ratio     float64           `json:"ratio"`
	Value     logic.M           `json:"value"`
	CreatedAt int64             `json:"createdAt"`
	FinishAt  int64             `json:"finishAt"`
}

// GetList GetList
func (t *task) GetList(ctx context.Context, req *GetListReq) (*GetListResp, error) {
	list, total, err := t.task.List(t.db, req.UserID, req.Types, int(req.Status), req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	resp := &GetListResp{
		List:  make([]*TaskListVo, len(list)),
		Total: total,
	}
	for index, value := range list {
		vo := clone(value)
		resp.List[index] = vo
	}
	return resp, nil

}

func clone(task *models.Task) *TaskListVo {
	vo := &TaskListVo{
		ID:        task.ID,
		Title:     task.Title,
		Command:   task.Command,
		Status:    task.Status,
		Ratio:     task.Ratio,
		Result:    getLogicM(task.Result),
		Value:     getLogicM(task.Value),
		CreatedAt: task.CreatedAt,
		FinishAt:  task.FinishAt,
	}
	return vo
}
func getLogicM(m models.M) logic.M {
	resp := make(logic.M)
	for key, value := range m {
		resp[key] = value
	}
	return resp
}

// GetByIDReq GetByIDReq
type GetByIDReq struct {
	TaskID string `json:"-"`
}

// GetByIDResp GetByIDResp
type GetByIDResp struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Status    models.TaskStatus `json:"status"`
	Result    logic.M           `json:"result"`
	Command   string            `json:"command"`
	Ratio     float64           `json:"ratio"`
	Value     logic.M           `json:"value"`
	CreatedAt int64             `json:"createdAt"`
	FinishAt  int64             `json:"finishAt"`
}

func (t *task) GetByID(ctx context.Context, req *GetByIDReq) (*GetByIDResp, error) {
	task, err := t.task.GetByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return &GetByIDResp{}, nil
	}
	return &GetByIDResp{
		ID:       task.ID,
		Title:    task.Title,
		Status:   task.Status,
		Result:   getLogicM(task.Result),
		Ratio:    task.Ratio,
		FinishAt: task.FinishAt,
	}, nil
}

// SubscribeReq SubscribeReq
type SubscribeReq struct {
	UserID string `json:"userID"`
	UUID   string `json:"uuid"`
	Topic  string `json:"topic" binding:"required"`
	Key    string `json:"key" binding:"required"`
}

// SubscribeResp SubscribeResp
type SubscribeResp struct {
	IsFinish bool              `json:"isFinish"`
	Status   models.TaskStatus `json:"status"`
}

func (t *task) Subscribe(ctx context.Context, req *SubscribeReq) (*SubscribeResp, error) {
	tasks, err := t.task.GetByID(t.db, req.Key)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		return &SubscribeResp{}, nil
	}
	if tasks.Status != models.TaskDoing {
		return &SubscribeResp{
			IsFinish: true,
			Status:   tasks.Status,
		}, nil
	}
	err = t.ps.Subscribe(ctx, &models.PubSub{
		UserID: req.UserID,
		Topic:  req.Topic,
		UUID:   req.UUID,
		Key:    req.Key,
	}, ttl)
	return &SubscribeResp{
		IsFinish: false,
	}, err
}

// GetProcessingReq GetProcessingReq
type GetProcessingReq struct {
	Types  string `json:"types"`
	UserID string `json:"user_id"`
}

// GetProcessingResp GetProcessingResp
type GetProcessingResp struct {
	Total int64 `json:"total"`
}

func (t *task) GetProcessing(ctx context.Context, req *GetProcessingReq) (*GetProcessingResp, error) {
	total, err := t.task.GetProcessing(t.db, req.UserID, req.Types)
	if err != nil {
		return nil, err
	}
	return &GetProcessingResp{
		Total: total,
	}, nil

}

// DeleteReq DeleteReq
type DeleteReq struct {
	TaskID string `json:"taskID"`
}

// DeleteResp DeleteResp
type DeleteResp struct {
}

func (t *task) Delete(ctx context.Context, req *DeleteReq) (*DeleteResp, error) {
	task1, err := t.task.GetByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	if task1 == nil {
		return &DeleteResp{}, nil
	}
	if task1.Status == models.TaskDoing {
		return nil, error2.New(code.ErrNotDelete)
	}
	err = t.task.DeleteByID(t.db, req.TaskID)
	if err != nil {
		return nil, err
	}
	return &DeleteResp{}, nil
}
