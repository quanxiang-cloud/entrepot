package comet

import (
	"context"
	"encoding/json"
	"github.com/quanxiang-cloud/cabin/logger"
	mysql2 "github.com/quanxiang-cloud/cabin/tailormade/db/mysql"
	redis2 "github.com/quanxiang-cloud/cabin/tailormade/db/redis"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	factor2 "github.com/quanxiang-cloud/entrepot/internal/comet/factor"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/internal/models/mysql"
	"github.com/quanxiang-cloud/entrepot/internal/models/redis"
	"github.com/quanxiang-cloud/entrepot/pkg/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"gorm.io/gorm"
)

const (
	entrepotType = "entrepot-task"
)

// FactorData FactorData
type FactorData struct {
	Command string
	Task    *models.Task
	Ctx     context.Context
}

// FactorManager FactorManager
type FactorManager struct {
	conf *config.Config

	Factor chan *FactorData

	handleData chan *basal.CallBackData //

	mold *factor2.Molders

	message client.MessageAPI

	taskRepo models.TaskRepo

	db *gorm.DB

	ps models.PubSubRepo
}

// NewFactor NewFactor
func NewFactor(ctx context.Context, conf *config.Config) (*FactorManager, error) {
	db, err := mysql2.New(conf.Mysql, logger.Logger)
	if err != nil {
		return nil, err
	}
	redisClient, err := redis2.NewClient(conf.Redis)
	if err != nil {
		return nil, err
	}

	manager := &FactorManager{
		Factor:     make(chan *FactorData),
		handleData: make(chan *basal.CallBackData),
		mold:       factor2.GetMolds(),
		conf:       conf,
		db:         db,
		taskRepo:   mysql.NewTaskRepo(),
		message:    client.NewMessage(conf.InternalNet),
		ps:         redis.NewPubSub(redisClient),
	}

	go manager.process(ctx, conf)

	return manager, nil
}

// Open multiple coroutines
func (f *FactorManager) process(ctx context.Context, conf *config.Config) {
	for i := 0; i < f.conf.ProcessorNum; i++ {
		go f.consumer(ctx, conf)
	}

}

func (f *FactorManager) consumer(ctx context.Context, conf *config.Config) {
	for {
		select {
		case <-ctx.Done():
			return
		case factor := <-f.Factor:
			//
			molders, ok := f.mold.GetMolders(factor.Ctx, logic.Command(factor.Command), conf)
			if !ok {
				factor.Task.Status = models.TaskFail
				f.handleResult(ctx, &basal.CallBackData{
					Task:  factor.Task,
					Types: basal.CallResult,
					Message: &basal.MesContent{
						TaskID: factor.Task.ID,
						Types:  basal.CallResult,
						Value:  models.TaskFail,
					},
				})
				break
			}
			molders.BatchHandle(factor.Ctx, factor.Task, f.handleData)
		case handle := <-f.handleData:
			f.handleResult(ctx, handle)
		}

	}
}

func (f *FactorManager) handleResult(ctx context.Context, handleResult *basal.CallBackData) {
	switch handleResult.Types {
	case basal.CallResult:
		// write the result
		err := f.taskRepo.FinishTask(f.db, handleResult.Task)
		if err != nil {
			logger.Logger.Errorw("err is ", err.Error(), handleResult.Task.ID)
		}

	case basal.CallRatio:
		// write the ratio
		err := f.taskRepo.ChangeRatio(f.db, handleResult.Task)
		if err != nil {
			logger.Logger.Errorw("err is ", err.Error(), handleResult.Task.ID)
		}
	}
	handleResult.Message.Command = handleResult.Task.Command
	f.sendMs(ctx, handleResult.Task.CreatorID, handleResult.Message)

}

//SendMs SendMs
func (f *FactorManager) sendMs(ctx context.Context, userID string, message *basal.MesContent) {

	consumers, err := f.ps.Get(ctx, entrepotType, message.TaskID)
	if err != nil {
		logger.Logger.Errorw("send ms ")
	}
	logger.Logger.Info("send ms", len(consumers))

	for _, consumer := range consumers {
		params := struct {
			Types   string            `json:"type"`
			Content *basal.MesContent `json:"content"`
		}{
			Types:   entrepotType,
			Content: message,
		}
		contentByte, err := json.Marshal(params)
		if err != nil {
			logger.Logger.Errorw("err is ")
		}
		_, err = f.message.SendMs(ctx, consumer.UserID, consumer.UUID, contentByte)
		if err != nil {
			logger.Logger.Errorw("err is", err.Error(), userID)
		}
	}

}
