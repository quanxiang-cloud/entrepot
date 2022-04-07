package factor

import (
	"bytes"
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
	"github.com/xuri/excelize/v2"
)

const (
	importNumber = 999
)

// FormImport FormImport
type FormImport struct {
	tableAPI client.TableAPI
	formAPI  client.FormAPI
	guide    *guide.Guide

	orgAPI     client.OrgAPI
	comElement *element.Component

	orgMap map[string]string
}

// SetValue SetValue
func (d *FormImport) SetValue(ctx context.Context, conf *config.Config) error {
	d.tableAPI = client.NewTableAPI(conf)
	d.formAPI = client.NewFormAPI(conf)
	guide, err := guide.NewGuide()
	if err != nil {
		return err
	}
	d.guide = guide
	d.orgAPI = client.NewOrgAPI(conf)
	d.comElement = element.NewComponent()
	d.setDepartmentMap(ctx)
	return nil
}

func (d *FormImport) setDepartmentMap(ctx context.Context) {
	treeResp, err := d.orgAPI.DepTree(ctx)
	if err != nil {
		logger.Logger.Errorw("tree client is error ")
	}
	treeMap := make(map[string]string)
	root := fmt.Sprintf("/%s", treeResp.DepartmentName)
	treeMap[root] = treeResp.ID
	for _, value := range treeResp.Child {
		depTraverse(root, "", value, treeMap)
	}
	d.orgMap = treeMap
}

func depTraverse(noRootPath, rootPath string, tree client.TreeResp, dep map[string]string) {
	noRootPath = fmt.Sprintf("%s/%s", noRootPath, tree.DepartmentName)
	rootPath = fmt.Sprintf("%s/%s", rootPath, tree.DepartmentName)
	dep[noRootPath] = tree.ID
	dep[rootPath] = tree.ID
	for _, value := range tree.Child {
		depTraverse(noRootPath, rootPath, value, dep)
	}
}

// NewFormImport NewFormImport
func NewFormImport() *FormImport {
	return &FormImport{}
}

func init() {
	SetMold(logic.CFormImport, NewFormImport())
}

type initImport struct {
	cSchema    *basal.ConvertSchema
	importData [][]string
	fieldKey   []string
	fieldName  []string
	failNumber int
	appID      string
	tableID    string
	resultURL  []logic.M
	errData    [][]string
	txtBuffer  *bytes.Buffer
	menuTitle  string
	result     models.M
}

// SetTaskTitle  SetTaskTitle
func (d *FormImport) SetTaskTitle(ctx context.Context, task *models.Task) error {
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
	task.Title = fmt.Sprintf("【%s】表单数据导入", fileName)
	return nil
}

// BatchHandle  BatchHandle
func (d *FormImport) BatchHandle(ctx context.Context, task *models.Task, callBack chan *basal.CallBackData) {
	var (
		err     error
		mImport = new(initImport)
	)

	defer func(mImport *initImport) {
		d.close()
		if err != nil {
			failCallBack(task, callBack, mImport.result)
			return
		}
		successCallBack(task, callBack, mImport.result)
	}(mImport)

	ctx = ctxCopy(ctx, getGetMapTask(task))
	err = d.pre(ctx, task, mImport)
	if err != nil {
		logger.Logger.Errorw("pre is err", err.Error())
		mImport.result = models.M{
			"title": err.Error(),
		}
		return
	}
	data := mImport.importData[2:]
	number := len(data) % importNumber
	len1 := len(data) / importNumber
	len2 := len1
	if number != 0 {
		len2 = len2 + 1
	}
	for i := 0; i < len1; i++ {
		splice := data[i*importNumber : i*importNumber+importNumber]
		d.importData(ctx, splice, mImport, i*importNumber+2)
		ratioCallBack(task, callBack, basal.Division(i+1, len1))
	}

	if number > 0 {
		splice := data[len1*importNumber:]
		d.importData(ctx, splice, mImport, len1*importNumber+2)
		ratioCallBack(task, callBack, 100)
	}
	resultTitle := fmt.Sprintf("导入数据成功，共处理%d条,成功了%d条,失败了%d条",
		len(mImport.importData)-2, len(mImport.importData)-2-mImport.failNumber, mImport.failNumber)

	if mImport.failNumber != 0 {
		file1 := fmt.Sprintf("%s导入错误数据.xlsx", mImport.menuTitle)
		path1 := fmt.Sprintf("import/%s/%s/%d/%s", mImport.appID, mImport.tableID, time2.NowUnix(), file1)
		toBuffer, err := writerDataExcel(mImport.errData, mImport.menuTitle)
		if err != nil {
			logger.Logger.Errorw("UploadFile1 is err", err.Error())
			return
		}
		err = d.guide.UploadFile(ctx, path1, toBuffer, int64(len(toBuffer.Bytes())))
		if err != nil {
			logger.Logger.Errorw("UploadFile1 is err", err.Error())
		}
		file2 := fmt.Sprintf("%s错误日志报告.txt", mImport.menuTitle)
		path2 := fmt.Sprintf("import/%s/%s/%d/%s", mImport.appID, mImport.tableID, time2.NowUnix(), file2)

		err = d.guide.UploadFile(ctx, path2, mImport.txtBuffer, int64(len(mImport.txtBuffer.Bytes())))
		if err != nil {
			logger.Logger.Errorw("UploadFile1 is err", err.Error())
		}
		mImport.resultURL = append(mImport.resultURL, logic.M{
			"url":      path1,
			"fileName": file1,
		}, logic.M{
			"url":      path2,
			"fileName": file2,
		})

	}
	mImport.result = models.M{
		"title": resultTitle,
		"path":  mImport.resultURL,
	}
}

func (d *FormImport) importData(ctx context.Context, data [][]string, mImport *initImport, start int) {
	handleData := d.handleData(ctx, data, mImport, start)
	err := d.createData(ctx, handleData.entity, mImport)
	if err != nil {
		mImport.failNumber = mImport.failNumber + len(data)
		return
	}
	mImport.failNumber = mImport.failNumber + len(handleData.errorResp)
	for _, row := range handleData.errorResp {
		for _, column := range row.Column {
			mImport.txtBuffer.WriteString(getErrorMessageTxt(row.Raw, column.Index, column.FieldsName, column.ErrorType))
		}
	}
}

func getErrorMessageTxt(row, column int, fieldName string, errorType code.ErrorType) string {
	message := code.ErrorTypeTable[errorType]
	return fmt.Sprintf("第%d行%d列【%s】导入失败:%s\n", row, column, fieldName, message)
}

type handleResp struct {
	entity    []*client.FormReq
	errorResp []*ImportError
}

// handleData handleData
func (d *FormImport) handleData(ctx context.Context, data [][]string, mImport *initImport, start int) *handleResp {
	entity, errorResp := d.getEntity(ctx, data, start, mImport)
	return &handleResp{
		entity:    entity,
		errorResp: errorResp,
	}
}

// post Data warehousing
func (d *FormImport) createData(ctx context.Context, entity []*client.FormReq, initImport *initImport) error {
	if entity == nil {
		return nil
	}
	_, err := d.formAPI.CreateBatch(ctx, initImport.appID, initImport.tableID, entity)
	if err != nil {
		return err
	}
	return nil
}

// ImportError ImportError
type ImportError struct {
	Raw    int
	Column []*ColumnError
}

// ColumnError ColumnError
type ColumnError struct {
	Index      int
	FieldsName string
	ErrorType  code.ErrorType
}

func (d *FormImport) getEntity(ctx context.Context, data [][]string, startNumber int, mImport *initImport) ([]*client.FormReq, []*ImportError) {
	entities := make([]*client.FormReq, 0)
	importErrors := make([]*ImportError, 0)
	for rawIndex, elem := range data {
		importError := new(ImportError)
		importError.Raw = rawIndex + startNumber + 1
		entity1 := make(map[string]interface{})
		for column, value := range elem {
			// read the column
			fieldsName := mImport.fieldKey[column]
			_, ok := mImport.cSchema.Component[fieldsName]
			if !ok {
				// Wrong number of rows
				c := &ColumnError{
					Index:      column + 1,
					ErrorType:  code.ErrorFieldIncorrect,
					FieldsName: mImport.fieldName[column],
				}
				importError.Column = append(importError.Column, c)
				continue
			}
			if value == "" {
				_, ok := mImport.cSchema.Required[fieldsName]
				if ok {
					c := &ColumnError{
						Index:      column + 1,
						ErrorType:  code.ErrNotBlank,
						FieldsName: mImport.fieldName[column],
					}
					importError.Column = append(importError.Column, c)
				}
				continue
			}
			c := mImport.cSchema.Component[fieldsName]
			m, err := d.comElement.GetMold(c, fieldsName, mImport.cSchema, d.orgAPI, d.orgMap)
			if err != nil {
				continue
			}
			handleValue, errType := m.GetValueByString(ctx, value)
			if errType != code.NoError {
				c := &ColumnError{
					Index:      column + 1,
					ErrorType:  errType,
					FieldsName: mImport.fieldName[column],
				}
				importError.Column = append(importError.Column, c)
			}
			entity1[fieldsName] = handleValue
		}
		if len(importError.Column) > 0 {
			importErrors = append(importErrors, importError)
			mImport.errData = append(mImport.errData, elem)

		} else {
			e := &client.FormReq{
				Entity: entity1,
			}
			entities = append(entities, e)
		}
	}
	if len(entities) > 0 {
		return entities, importErrors
	}
	return nil, importErrors
}

func (d *FormImport) pre(ctx context.Context, task *models.Task, mImport *initImport) error {
	if mImport == nil {
		mImport = new(initImport)
	}
	value := convert(task)
	appID, err := GetMapToString(value, "appID")
	if err != nil {
		return err
	}
	mImport.appID = appID
	taleID, err := GetMapToString(value, "tableID")
	if err != nil {
		return err
	}
	mImport.tableID = taleID
	table, err := d.tableAPI.HomeTableSchema(ctx, appID, taleID)
	if err != nil {
		return err
	}
	cSchema, err := createImportData(table.Schema)
	if err != nil {
		return err
	}
	mImport.cSchema = cSchema

	fileName, err := GetMapToString(table.Schema, "title")
	if err != nil {
		return err
	}
	mImport.menuTitle = fileName

	downBytes := new(bytes.Buffer)
	err = d.guide.DownloadFile(ctx, task.FileAddr, downBytes)
	if err != nil {
		return err
	}

	reader, err := excelize.OpenReader(downBytes)
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}

	data, err := reader.GetRows(reader.GetSheetName(0))
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}

	if len(data) < 3 {
		return error2.New(code.ErrFile)
	}
	mImport.importData = data
	fieldKey := make([]string, len(data[0]))
	for c, value1 := range data[0] {
		fieldKey[c] = value1
	}
	fieldName := make([]string, len(data[0]))
	for c, value1 := range data[1] {
		fieldName[c] = value1
	}
	mImport.fieldKey = fieldKey
	mImport.fieldName = fieldName
	isPass := preImportField(fieldKey, cSchema)

	if !isPass {
		return error2.New(code.ErrNoAuth)
	}
	mImport.errData = make([][]string, 0)
	mImport.errData = append(mImport.errData, fieldKey)
	mImport.errData = append(mImport.errData, fieldName)
	mImport.txtBuffer = new(bytes.Buffer)

	return nil
}

func createImportData(schema logic.M) (*basal.ConvertSchema, error) {
	properties := schema[properties]
	pr, err := GetAsMap(properties)
	if err != nil {
		return nil, err
	}
	s := &basal.ConvertSchema{
		Component:  make(map[string]string),
		Properties: make(logic.M),
		Required:   make(map[string]struct{}),
	}
	err = getSchema(pr, s)
	if err != nil {
		return nil, err
	}
	return s, nil

}

func (d *FormImport) close() {
	d.tableAPI.Close()
	d.formAPI.Close()
	d.orgAPI.Close()
}

func writerDataExcel(data [][]string, sheet string) (*bytes.Buffer, error) {
	fs := excelize.NewFile()
	fs.SetSheetName("Sheet1", sheet)
	streamWriter, err := fs.NewStreamWriter(sheet)
	if err != nil {
		return nil, err
	}
	for rowID := 1; rowID <= len(data); rowID++ {
		colLen := len(data[rowID-1])

		row := make([]interface{}, colLen)

		for colID := 0; colID < colLen; colID++ {

			row[colID] = data[rowID-1][colID]
		}
		cell, err := excelize.CoordinatesToCellName(1, rowID)
		if err != nil {
			return nil, err
		}
		if err := streamWriter.SetRow(cell, row); err != nil {
			return nil, err
		}
	}
	err = streamWriter.Flush()
	if err != nil {
		return nil, err
	}
	toBuffer, err := fs.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return toBuffer, nil
}
