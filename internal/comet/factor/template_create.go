package factor

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
)

// TemplateCreate TemplateCreate
type TemplateCreate struct {
	appData AppData
}

func init() {
	SetMold(logic.CCreateTemplate, NewTemplateCreate())
}

// NewTemplateCreate NewTemplateCreate
func NewTemplateCreate() *TemplateCreate {
	return &TemplateCreate{}
}

// SetTaskTitle SetTaskTitle
func (t *TemplateCreate) SetTaskTitle(ctx context.Context, task *models.Task) error {

	return nil
}

// SetValue SetValue
func (t *TemplateCreate) SetValue(ctx context.Context, conf *config.Config) error {
	t.appData = NewAppData(conf)
	return nil
}

// BatchHandle BatchHandle
func (t *TemplateCreate) BatchHandle(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	ctx = ctxCopy(ctx, getGetMapTask(task))
	t.appData.CreateTemplate(ctx, task, handleData)
}
