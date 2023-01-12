package mysql

import (
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"gorm.io/gorm"
)

type taskRepo struct {
}

func (t *taskRepo) GetByCondition(db *gorm.DB, task *models.Task) (*models.Task, error) {
	tasks := new(models.Task)
	ql := db.Table(t.TableName()).Where("status = 1")
	if task.FileAddr != "" {
		ql = ql.Where("file_addr = ? ", task.FileAddr)
	}
	if task.FileOpt != "" {
		ql = ql.Where("file_opt = ?", task.FileAddr)
	}
	if task.FileSize != 0 {
		ql = ql.Where("file_size = ?", task.FileAddr)

	}
	if task.Command != "" {
		ql = ql.Where("command = ?", task.Command)
	}
	if task.Value != nil {
		ql = ql.Where("value = ?", task.Value)
	}
	if task.CreatorID != "" {
		ql = ql.Where("creator_id = ?", task.CreatorID)
	}
	err := ql.Find(tasks).Error

	if err != nil {
		return nil, err
	}
	if tasks.ID == "" {
		return nil, nil
	}
	return tasks, nil
}

func (t *taskRepo) DeleteByID(db *gorm.DB, id string) error {
	return db.Table(t.TableName()).Where("id = ?", id).
		Delete(&models.Task{}).
		Error
}

func (t *taskRepo) GetProcessing(db *gorm.DB, userID string, types string) (int64, error) {
	ql := db.Table(t.TableName())
	ql = ql.Where("creator_id = ? and types = ? and  status = 1", userID, types)
	var total int64
	ql.Count(&total)

	return total, nil
}

func (t *taskRepo) GetByID(db *gorm.DB, taskID string) (*models.Task, error) {
	task := new(models.Task)
	err := db.Table(t.TableName()).
		Where("id = ?", taskID).
		Find(task).Error
	if err != nil {
		return nil, err
	}
	if task.ID == "" {
		return nil, nil
	}
	return task, nil
}

func (t *taskRepo) ChangeRatio(db *gorm.DB, task *models.Task) error {
	return db.Table(t.TableName()).Where("id = ?", task.ID).Updates(
		map[string]interface{}{
			"ratio": task.Ratio,
		}).Error
}

func (t *taskRepo) Create(db *gorm.DB, task *models.Task) error {
	return db.Table(t.TableName()).Create(task).Error
}

func (t *taskRepo) List(db *gorm.DB, creatorID string, types string, status, page, limit int) ([]*models.Task, int64, error) {
	ql := db.Table(t.TableName())
	if creatorID != "" {
		ql = ql.Where("creator_id = ? ", creatorID)
	}
	if types != "" {
		ql = ql.Where("types = ? ", types)
	}
	if status != 0 {
		ql = ql.Where("status = ? ", status)
	}
	var total int64
	ql.Count(&total)
	ql = ql.Limit(limit).Offset((page - 1) * limit)
	ql = ql.Order("created_at DESC")
	taskList := make([]*models.Task, 0)
	err := ql.Find(&taskList).Error
	return taskList, total, err
}

func (t *taskRepo) ListProcessing(db *gorm.DB) ([]*models.Task, error) {
	ql := db.Table(t.TableName()).Where("status = 1")
	taskList := make([]*models.Task, 0)
	err := ql.Find(&taskList).Error
	return taskList, err
}

func (t *taskRepo) FinishTask(db *gorm.DB, task *models.Task) error {
	return db.Table(t.TableName()).Where("id = ?", task.ID).Updates(
		map[string]interface{}{
			"status":    task.Status,
			"result":    task.Result,
			"finish_at": task.FinishAt,
			"ratio":     100,
		}).Error
}

// TableName TableName
func (t *taskRepo) TableName() string {
	return "task"
}

// NewTaskRepo NewTaskRepo
func NewTaskRepo() models.TaskRepo {
	return &taskRepo{}
}
