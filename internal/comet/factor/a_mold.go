package factor

import (
	"context"
	"github.com/quanxiang-cloud/entrepot/internal/logic"

	"sync"

	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
)

//Mold Mold
type Mold interface {
	BatchHandle(context.Context, *models.Task, chan *basal.CallBackData)
	SetValue(ctx context.Context, conf *config.Config) error
	SetTaskTitle(context.Context, *models.Task) error
}

func init() {
	molds = &Molders{
		Molder: make(map[logic.Command]Mold),
	}
}

var molds *Molders

//Molders  Molders
type Molders struct {
	sync.RWMutex
	Molder map[logic.Command]Mold
}

//GetMolds GetMolds
func GetMolds() *Molders {
	return molds
}

// SetMold SetMold
func SetMold(name logic.Command, m Mold) {
	molds.Lock()
	molds.Molder[name] = m
	molds.Unlock()
}

// GetMolders GetMolders
func (m *Molders) GetMolders(ctx context.Context, _c logic.Command, conf *config.Config) (Mold, bool) {
	mold, ok := m.Molder[_c]
	if !ok {
		return nil, false
	}
	err := mold.SetValue(ctx, conf)
	if err != nil {
		return nil, false
	}
	return mold, true
}
