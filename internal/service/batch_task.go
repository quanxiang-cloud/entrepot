package service

import (
	"context"
	"time"

	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	mysql2 "github.com/quanxiang-cloud/cabin/tailormade/db/mysql"
	redis2 "github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/entrepot/internal/comet"
	factor2 "github.com/quanxiang-cloud/entrepot/internal/comet/factor"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models/mysql"
	"github.com/quanxiang-cloud/entrepot/internal/models/redis"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"gorm.io/gorm"

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
	Timeout(ctx context.Context)
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
		db:     db,
		factor: factor,
		mold:   factor2.GetMolds(),
		task:   mysql.NewTaskRepo(),
		ps:     redis.NewPubSub(redisClient),
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

// DaprEvent DaprEvent
type DaprEvent struct {
	Topic           string         `json:"topic"`
	Pubsubname      string         `json:"pubsubname"`
	Traceid         string         `json:"traceid"`
	ID              string         `json:"id"`
	Datacontenttype string         `json:"datacontenttype"`
	Data            *CreateTaskReq `json:"data"`
	Type            string         `json:"type"`
	Specversion     string         `json:"specversion"`
	Source          string         `json:"source"`
}

const (
	timeout = 60 * 60 * 10
)

func (t *task) Timeout(ctx context.Context) {
	tick := time.NewTicker(time.Minute * 10)
	go func(t *task) {
		for {
			tasks, err := t.task.ListProcessing(t.db)
			if err != nil {
				logger.Logger.Errorw("fail list porcessing task", "err", err.Error())
				continue
			}

			now := time2.NowUnix() / 1000
			for _, task := range tasks {
				if now-task.CreatedAt > timeout {
					task.Status = models.TaskFail
					task.Result = make(models.M)
					task.Result["title"] = "Timeout"
					task.FinishAt = time2.NowUnix()
					t.task.FinishTask(t.db, task)
				}
			}

			select {
			case <-ctx.Done():
				tick.Stop()
				return
			case <-tick.C:
			}
		}
	}(t)
}
