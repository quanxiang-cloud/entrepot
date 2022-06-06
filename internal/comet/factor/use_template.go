package factor

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
)

// UseTemplate UseTemplate
type UseTemplate struct {
	appData AppData
}

func init() {
	SetMold(logic.CUseTemplate, NewUseTemplate())
}

// NewUseTemplate NewUseTemplate
func NewUseTemplate() *UseTemplate {
	return &UseTemplate{}
}

// SetTaskTitle SetTaskTitle
func (t *UseTemplate) SetTaskTitle(ctx context.Context, task *models.Task) error {
	return nil
}

// SetValue SetValue
func (t *UseTemplate) SetValue(ctx context.Context, conf *config.Config) error {
	t.appData = NewAppData(conf)
	return nil
}

// BatchHandle BatchHandle
func (t *UseTemplate) BatchHandle(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	ctx = ctxCopy(ctx, getGetMapTask(task))
	t.appData.UseTemplate(ctx, task, handleData)
}
