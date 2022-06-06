package factor

import (
	"context"
	"fmt"
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/comet/element"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"github.com/quanxiang-cloud/fileserver/pkg/guide"
	"sort"
)

const (
	title = "title"
)

const (
	_id           = "_id"
	_createdAt    = "created_at"
	_creatorID    = "creator_id"
	_creatorName  = "creator_name"
	_updatedAt    = "updated_at"
	_modifierID   = "modifier_id"
	_modifierName = "modifier_name"
)

// FormTemplate FormTemplate
type FormTemplate struct {
	tableAPI   client.TableAPI
	formAPI    client.FormAPI
	guide      *guide.Guide
	orgAPI     client.OrgAPI
	comElement *element.Component
}

// SetTaskTitle SetTaskTitle
func (f *FormTemplate) SetTaskTitle(ctx context.Context, task *models.Task) error {
	if task.Title != "" {
		return nil
	}
	value := convert(task)
	tableID, err := GetMapToString(value, "tableID")
	if err != nil {
		return err
	}
	appID, err := GetMapToString(value, "appID")
	if err != nil {
		return err
	}

	table, err := f.tableAPI.HomeTableSchema(ctx, appID, tableID)
	if err != nil {
		return err
	}
	fileName, err := GetMapToString(table.Schema, "title")
	if err != nil {
		return err
	}
	task.Title = fmt.Sprintf("【%s】表单模板导出", fileName)
	return nil
}

// SetValue SetValue
func (f *FormTemplate) SetValue(ctx context.Context, conf *config.Config) error {
	f.tableAPI = client.NewTableAPI(conf)
	f.formAPI = client.NewFormAPI(conf)
	guide, err := guide.NewGuide()
	if err != nil {
		return err
	}
	f.guide = guide
	f.orgAPI = client.NewOrgAPI(conf)
	f.comElement = element.NewComponent()
	return nil
}

type initDown struct {
	appID       string
	tableID     string
	schema      logic.M
	schemaIndex []schemeIndex
	result      models.M
}

// BatchHandle BatchHandle  export form template
func (f *FormTemplate) BatchHandle(ctx context.Context, task *models.Task, data chan *basal.CallBackData) {
	var (
		err  error
		init = new(initDown)
	)

	defer func(init *initDown) {
		f.close()
		if err != nil {
			failCallBack(task, data, init.result)
			return
		}
		successCallBack(task, data, init.result)
	}(init)
	ctx = ctxCopy(ctx, getGetMapTask(task))
	err = f.downloadPre(ctx, task, init)
	if err != nil {
		init.result = models.M{
			"title": err.Error(),
		}
		return
	}
	url, fileName, err := f.downloadTemplate(ctx, init)
	if err != nil {
		logger.Logger.Errorw("taskID", err.Error(), task.ID)
	}
	init.result = models.M{
		"title": "模板导出成功",
		"path": []logic.M{{
			"url":      url,
			"fileName": fileName,
		}},
	}
}

func (f *FormTemplate) downloadPre(ctx context.Context, task *models.Task, init *initDown) error {
	// 1、get the schema
	value := convert(task)
	appID, err := GetMapToString(value, "appID")
	if err != nil {
		return err
	}
	init.appID = appID
	taleID, err := GetMapToString(value, "tableID")
	if err != nil {
		return err
	}
	init.tableID = taleID
	table, err := f.tableAPI.HomeTableSchema(ctx, appID, taleID)
	if err != nil {
		return error2.New(code.ErrInternalError)
	}
	init.schema = table.Schema
	init.schemaIndex = make([]schemeIndex, 0)
	properties := init.schema[properties]
	pr, err := GetAsMap(properties)
	if err != nil {
		return err
	}
	err = TraverseSchema(pr, init)
	if err != nil {
		return err
	}
	if len(init.schemaIndex) == 0 {
		return error2.New(code.ErrNotField)
	}
	return nil

}

func (f *FormTemplate) downloadTemplate(ctx context.Context, init *initDown) (string, string, error) {
	fileName, err := GetMapToString(init.schema, title)
	if err != nil {
		return "", "", err
	}

	fieldKey := make([]string, len(init.schemaIndex))
	fieldName := make([]string, len(init.schemaIndex))
	sort.SliceStable(init.schemaIndex, func(i, j int) bool {
		return init.schemaIndex[i].Index < init.schemaIndex[j].Index
	})
	for index, value := range init.schemaIndex {
		fieldKey[index] = value.FieldKey
		fieldName[index] = value.FieldName
	}
	data := make([][]string, 2)
	data[0] = fieldKey
	data[1] = fieldName
	toBuffer, err := writerDataExcel(data, fileName)
	if err != nil {
		return "", "", err
	}
	fileName = fmt.Sprintf("%s-%s.xlsx", fileName, "导入模板")
	path := fmt.Sprintf("import/%s/%s/%d/%s", init.appID, init.tableID, time2.NowUnix(), fileName)
	err = f.guide.UploadFile(ctx, path, toBuffer, int64(len(toBuffer.Bytes())))
	if err != nil {
		return "", "", err
	}
	return path, fileName, nil
}

type schemeIndex struct {
	FieldKey  string
	FieldName string
	Index     int
}

// TraverseSchema TraverseSchema
func TraverseSchema(schema logic.M, init *initDown) error {
	for key, value := range schema {
		switch key {
		case _id, _createdAt, _creatorID, _creatorName, _updatedAt, _modifierID, _modifierName:
			continue
		}
		isLayout := IsLayoutComponent(value)
		columnValue, err := GetAsMap(value)
		if err != nil {
			return err
		}
		if isLayout {
			if p, ok := columnValue[properties]; ok {
				pro, err := GetAsMap(p)
				if err != nil {
					break
				}
				err = TraverseSchema(pro, init)
				if err != nil {
					break
				}
			}
		}
		xCom, err := GetMapToString(columnValue, component)
		if err != nil {
			return err
		}
		_, exit := comMap[xCom]
		if !exit {
			continue
		}
		if !getEditPermission(value) {
			continue
		}

		titles1, err := GetMapToString(columnValue, title)

		if err != nil {
			return err
		}
		index, err := GetMapToInt(columnValue, "x-index")
		if err != nil {
			return nil
		}
		s1 := schemeIndex{
			Index:     index,
			FieldName: titles1,
			FieldKey:  key,
		}
		init.schemaIndex = append(init.schemaIndex, s1)
	}
	return nil
}

// NewFormTemplate NewFormTemplate
func NewFormTemplate() *FormTemplate {
	return &FormTemplate{}
}

func init() {
	SetMold(logic.CFormTemplate, NewFormTemplate())
}

func (f *FormTemplate) close() {
	f.tableAPI.Close()
	f.formAPI.Close()
	f.orgAPI.Close()
}
