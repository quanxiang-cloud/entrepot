package models

import (
	"database/sql/driver"
	"encoding/json"

	"gorm.io/gorm"
)

// M m
type M map[string]interface{}

// TaskStatus TaskStatus
type TaskStatus int

const (
	// TaskSuccess TaskSuccess
	TaskSuccess TaskStatus = 2
	// TaskFail TaskFail
	TaskFail TaskStatus = 3
	//TaskDoing TaskDoing
	TaskDoing TaskStatus = 1
)

// Task task list
type Task struct {
	ID          string
	CreatedAt   int64
	FinishAt    int64
	CreatorID   string
	CreatorName string
	Title       string
	Types       string
	Command     string
	FileAddr    string
	FileSize    int
	FileOpt     string
	Value       M
	Result      M
	Status      TaskStatus
	Ratio       float64
	DepID       string `gorm:"-"`
}

// Value Value
func (p M) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan Scan
func (p *M) Scan(data interface{}) error {
	return json.Unmarshal(data.([]byte), &p)
}

// Unmarshal Unmarshal M to struct pointer
func (p *M) Unmarshal(pointer interface{}) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, pointer)
}

// TaskRepo TaskRepo
type TaskRepo interface {
	Create(*gorm.DB, *Task) error
	List(*gorm.DB, string, string, int, int, int) ([]*Task, int64, error)
	FinishTask(*gorm.DB, *Task) error
	ChangeRatio(db *gorm.DB, task *Task) error
	GetByID(db *gorm.DB, taskID string) (*Task, error)
	GetProcessing(db *gorm.DB, id string, types string) (int64, error)
	DeleteByID(db *gorm.DB, id string) error
	GetByCondition(db *gorm.DB, task *Task) (*Task, error)
}
