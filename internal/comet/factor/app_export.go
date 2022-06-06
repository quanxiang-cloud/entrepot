package factor

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
)

// AppExport AppExport
type AppExport struct {
	appData AppData
}

func init() {
	SetMold(logic.CAppExport, NewAppExport())
}

// NewAppExport NewAppExport
func NewAppExport() *AppExport {
	return &AppExport{}
}

// SetTaskTitle SetTaskTitle
func (a *AppExport) SetTaskTitle(ctx context.Context, task *models.Task) error {

	return nil
}

// SetValue SetValue
func (a *AppExport) SetValue(ctx context.Context, conf *config.Config) error {
	a.appData = NewAppData(conf)
	return nil
}

// BatchHandle BatchHandle
func (a *AppExport) BatchHandle(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	ctx = ctxCopy(ctx, getGetMapTask(task))
	a.appData.ExportAppData(ctx, task, handleData)
}
