package factor

import (
	"context"
	"fmt"
	"git.internal.yunify.com/qxp/fileserver/pkg/guide"
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
)

const (
	pageSize     = 999
	csvMaxNumber = 100000
)

// FormExport FormExport
type FormExport struct {
	tableAPI   client.TableAPI
	formAPI    client.FormAPI
	orgAPI     client.OrgAPI
	comElement *element.Component
	guide      *guide.Guide
}

// NewFormExport  NewFormExport
func NewFormExport() *FormExport {
	return &FormExport{}
}

// SetValue SetValue
func (d *FormExport) SetValue(ctx context.Context, conf *config.Config) error {
	d.tableAPI = client.NewTableAPI(conf)
	d.formAPI = client.NewFormAPI(conf)
	guide, err := guide.NewGuide()
	if err != nil {
		return err
	}
	d.guide = guide
	d.orgAPI = client.NewOrgAPI(conf)
	d.comElement = element.NewComponent()
	return nil
}

func init() {
	SetMold(logic.CFormExport, NewFormExport())
}

type initExport struct {
	schema *basal.ConvertSchema
	total  int64
	//buffer     *bytes.Buffer
	//writer     *csv.Writer
	entity     []map[string]interface{}
	url        []logic.M
	csvNumber  int
	totalFile  int
	title      string
	tableID    string
	appID      string
	filterKey  []string
	filterName []string
	query      interface{}
	success    int64
	exportData [][]string
	result     models.M
}

// BatchHandle BatchHandle
func (d *FormExport) BatchHandle(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	var (
		err     error
		mExport = new(initExport)
	)
	defer func(export *initExport) {
		d.close()
		if err != nil {
			failCallBack(task, handleData, mExport.result)
			return
		}
		successCallBack(task, handleData, mExport.result)
	}(mExport)
	ctx = ctxCopy(ctx, getGetMapTask(task))
	err = d.exportPre(ctx, task, mExport)
	if err != nil {
		logger.Logger.Errorw("err is ", task.ID, err.Error())
		mExport.result = models.M{
			"title": err.Error(),
		}
		return
	}
	number := int(mExport.total) % pageSize
	len1 := int(mExport.total) / pageSize

	var length = len1
	if number != 0 {
		length = length + 1
	}
	for i := 0; i < length; i++ {
		findOpt := client.FindOptions{
			Page: int64(i + 1),
			Size: pageSize,
			Sort: []string{"-created_at"},
		}
		searchResp, err := d.formAPI.Search(ctx, findOpt, mExport.query, mExport.appID, mExport.tableID)
		if err != nil {
			logger.Logger.Errorw("search is err ", err.Error())
			continue
		}
		mExport.entity = searchResp.Entities
		mExport.csvNumber = mExport.csvNumber + len(searchResp.Entities)
		err = d.exportEntity(ctx, mExport, i == length-1)
		if err != nil {
			logger.Logger.Errorw(" exportEntity is err ", err.Error())
		}
		ratioCallBack(task, handleData, basal.Division(i+1, len1))
		mExport.success = mExport.success + int64(len(searchResp.Entities))
	}
	mExport.result = models.M{
		"title": fmt.Sprintf("导出数据成功,共导出%d条数据,成功%d条", mExport.total, mExport.success),
		"path":  mExport.url,
	}
}

func (d *FormExport) exportEntity(ctx context.Context, mExport *initExport, tail bool) error {
	var second bool
	var preEntity []map[string]interface{}
	var superEntity []map[string]interface{}
	csvNumber := mExport.csvNumber
	entity := mExport.entity
	if csvNumber > csvMaxNumber {
		surplus := csvNumber - csvMaxNumber
		preEntity = mExport.entity[0 : len(entity)-surplus]
		second = true
		superEntity = entity[len(entity)-surplus:]
	} else if csvNumber <= csvMaxNumber {
		preEntity = entity
	}

	err := d.exportFilerServer(ctx, preEntity, mExport)
	if err != nil {
		logger.Logger.Errorw("exportFilerServer err is", err.Error())
		return err
	}
	if csvNumber >= csvMaxNumber || tail {
		//mExport.writer.Flush()
		url, fileName := d.upLoadFile(ctx, mExport)
		if url != "" {
			mExport.url = append(mExport.url, logic.M{
				"url":      url,
				"fileName": fileName,
			})
		}
	}
	if second {
		mExport.exportData = make([][]string, 0)
		mExport.exportData = append(mExport.exportData, mExport.filterKey)
		if err != nil {
			logger.Logger.Errorw("set title is err", err.Error())
			return err
		}
		err = d.exportFilerServer(ctx, superEntity, mExport)
		if err != nil {
			logger.Logger.Errorw("exportFilerServer is err", err.Error())
			return err
		}
		mExport.csvNumber = len(superEntity)
		if tail {
			//mExport.writer.Flush()
			url, fileName := d.upLoadFile(ctx, mExport)
			if url != "" {
				mExport.url = append(mExport.url, logic.M{
					"url":      url,
					"fileName": fileName,
				})
			}
		}
	}
	return nil
}

func (d *FormExport) exportFilerServer(ctx context.Context, entities []map[string]interface{}, mExport *initExport) error {
	filters := make(map[string]struct{})
	for _, value := range mExport.filterKey {
		filters[value] = struct{}{}
	}
	for _, entity1 := range entities {
		row := make([]string, len(filters))
		for columnIndex, value := range mExport.filterKey {
			fieldValue, ok := entity1[value]
			if ok {
				mold, err := d.comElement.GetMold(mExport.schema.Component[value], value, mExport.schema, d.orgAPI, nil)
				if err != nil {
					continue
				}

				row[columnIndex] = mold.ExportValue(ctx, fieldValue)
			}

		}
		mExport.exportData = append(mExport.exportData, row)
	}
	return nil
}

func (d *FormExport) upLoadFile(ctx context.Context, mExport *initExport) (string, string) {

	fileName := fmt.Sprintf("%s%d.xlsx", mExport.title, mExport.totalFile)

	toBuffer, err := writerDataExcel(mExport.exportData, mExport.title)
	if err != nil {
		logger.Logger.Errorw("UploadFile is err ", err.Error())
		return "", ""
	}

	path := fmt.Sprintf("import/%s/%s/%d/%s", mExport.appID, mExport.tableID, time2.NowUnix(), fileName)
	err = d.guide.UploadFile(ctx, path, toBuffer, int64(len(toBuffer.Bytes())))
	if err != nil {
		logger.Logger.Errorw("UploadFile is err ", err.Error())
		return "", ""
	}
	mExport.totalFile = mExport.totalFile + 1
	return path, fileName
}

func (d *FormExport) exportPre(ctx context.Context, task *models.Task, mExport *initExport) error {
	if mExport == nil {
		mExport = new(initExport)
	}
	mExport.url = make([]logic.M, 0)
	value := convert(task)
	appID, err := GetMapToString(value, "appID")
	if err != nil {
		return err
	}
	tableID, err := GetMapToString(value, "tableID")
	if err != nil {
		return err
	}
	filterKey, err := GetMapToArrStr(value, "filterKey")
	if err != nil {
		return err
	}
	filterName, err := GetMapToArrStr(value, "filterName")
	if err != nil {
		return err
	}
	mExport.tableID = tableID
	mExport.appID = appID
	mExport.filterKey = filterKey
	mExport.filterName = filterName
	mExport.query = value["query"]
	table, err := d.tableAPI.HomeTableSchema(ctx, appID, tableID)
	if err != nil {
		return err
	}
	fileName, err := GetMapToString(table.Schema, "title")
	if err != nil {
		return err
	}
	mExport.title = fileName
	schema, err := createImportData(table.Schema)
	if err != nil {
		return err
	}
	mExport.schema = schema
	isPass := preCheckField(mExport.filterKey, schema)
	if !isPass {
		return error2.New(code.ErrNoAuth)
	}
	options := client.FindOptions{
		Page: 1,
		Size: 1,
	}
	searchResp, err := d.formAPI.Search(ctx, options, mExport.query, appID, tableID)
	if err != nil {
		return error2.New(code.ErrInternalError)
	}
	if searchResp.Total == 0 {
		return error2.New(code.ErrHighNoData)
	}
	mExport.total = searchResp.Total
	mExport.exportData = make([][]string, 0)
	mExport.exportData = append(mExport.exportData, mExport.filterName)
	return nil
}

// SetTaskTitle SetTaskTitle
func (d *FormExport) SetTaskTitle(ctx context.Context, task *models.Task) error {
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

	table, err := d.tableAPI.HomeTableSchema(ctx, appID, tableID)
	if err != nil {
		return err
	}
	fileName, err := GetMapToString(table.Schema, "title")
	if err != nil {
		return err
	}
	task.Title = fmt.Sprintf("【%s】表单数据导出", fileName)
	return nil
}

func (d *FormExport) close() {
	d.tableAPI.Close()
	d.formAPI.Close()
	d.orgAPI.Close()
}
