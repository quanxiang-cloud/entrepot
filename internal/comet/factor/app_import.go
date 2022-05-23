package factor

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
)

const (
	appFileName        = "appInfo"
	tableFileName      = "tableInfo"
	fileInfo           = "fileInfo"
	flowFileName       = "flowInfo"
	apiFileName        = "apiInfo"
	pageEngineFileName = "pageEngine"
)

const (
	appIDKey = "appID"
)

// AppImport AppImport
type AppImport struct {
	appData AppData
}

func init() {
	SetMold(logic.CAppImport, NewAppImport())
}

// NewAppImport NewAppImport
func NewAppImport() *AppImport {
	return &AppImport{}
}

// SetTaskTitle SetTaskTitle
func (a *AppImport) SetTaskTitle(ctx context.Context, task *models.Task) error {
	return nil
}

// SetValue SetValue
func (a *AppImport) SetValue(ctx context.Context, conf *config.Config) error {
	a.appData = NewAppData(conf)
	return nil
}

// BatchHandle BatchHandle
func (a *AppImport) BatchHandle(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	ctx = ctxCopy(ctx, getGetMapTask(task))
	a.appData.ImportAppData(ctx, task, handleData)
}
