package factor

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"git.internal.yunify.com/qxp/fileserver/pkg/guide"
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"github.com/quanxiang-cloud/entrepot/pkg/zip2"
	"strings"
)

const (
	successImportApp      = "导入应用成功"
	successExportApp      = "导出应用成功"
	successCreateTemplate = "创建模板成功"
	successCreateApp      = "模板创建应用成功"
)
const (
	fileNameTemp = "%s.zip"
	filePathTemp = "appExport/%d/%s/%s"
)

const (
	fileNameKey = "fileName"
	ulrKey      = "url"
	pathKey     = "path"
	titleKey    = "title"
)

// AppData deal with app information
type AppData interface {
	ExportAppData(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData)
	ImportAppData(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData)
	UseTemplate(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData)
	CreateTemplate(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData)
}

type appData struct {
	polyAPI   client.PolyAPI
	flow      client.Flow
	structor  client.Structor
	fileSDK   *guide.Guide
	appCenter client.AppCenter
	persona   client.Persona
}

// NewAppData NewAppData
func NewAppData(conf *config.Config) AppData {
	fileSDK, _ := guide.NewGuide()
	return &appData{
		polyAPI:   client.NewPolyAPI(conf),
		flow:      client.NewFlow(conf),
		fileSDK:   fileSDK,
		structor:  client.NewStructor(conf),
		appCenter: client.NewAppCenter(conf),
		persona:   client.NewPersona(conf),
	}
}

func (a *appData) ExportAppData(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	defer a.closeConnect()
	result := models.M{}
	appID, err := a.getAppID(ctx, task)
	if err != nil {
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		return
	}
	path, fileName, err := a.packAppData(ctx, appID, task, handleData)
	if err != nil {
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		return
	}
	a.sendRate(task, handleData, 100)

	pathResult := make([]logic.M, 0)
	pathResult = append(pathResult, logic.M{
		fileNameKey: fileName,
		ulrKey:      path,
	})
	result[titleKey] = successExportApp
	result[pathKey] = pathResult
	a.sendMessage(nil, result, task, handleData)
}

func (a *appData) ImportAppData(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	defer a.closeConnect()
	result := models.M{}
	appID, err := a.getAppID(ctx, task)
	if err != nil {
		result[titleKey] = err.Error()
		_ = a.appCenter.FailImport(ctx, appID)
		a.sendMessage(err, result, task, handleData)

		return
	}
	err = a.insertAppData(ctx, appID, task, handleData)
	if err != nil {
		result[titleKey] = err.Error()
		_ = a.appCenter.FailImport(ctx, appID)
		a.sendMessage(err, result, task, handleData)

		return
	}
	err = a.appCenter.SuccessImport(ctx, appID)
	if err != nil {
		err = error2.New(code.ErrInternalError)
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)

		return
	}
	a.sendRate(task, handleData, 100)
	result[titleKey] = successImportApp
	a.sendMessage(nil, result, task, handleData)

}

func (a *appData) UseTemplate(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	defer a.closeConnect()
	result := models.M{}
	// get the request params
	req := &struct {
		AppID      string `json:"appID"`
		TemplateID string `json:"templateID"`
	}{}
	err := task.Value.Unmarshal(req)
	if err != nil {
		err = error2.New(code.ErrParameter)
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		return
	}
	// get the template information
	template, err := a.appCenter.GetTemplateByID(ctx, req.TemplateID)
	if err != nil {
		err = error2.New(code.ErrTemplateNotExist)
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		_ = a.appCenter.FailImport(ctx, req.AppID)
		return
	}
	task.FileAddr = template.Path
	err = a.insertAppData(ctx, req.AppID, task, handleData)
	if err != nil {
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		_ = a.appCenter.FailImport(ctx, req.AppID)
		return
	}
	err = a.appCenter.SuccessImport(ctx, req.AppID)
	if err != nil {
		err = error2.New(code.ErrInternalError)
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		_ = a.appCenter.FailImport(ctx, req.AppID)
		return
	}
	a.sendRate(task, handleData, 100)
	result[titleKey] = successCreateApp
	a.sendMessage(nil, result, task, handleData)
}

func (a *appData) CreateTemplate(ctx context.Context, task *models.Task, handleData chan *basal.CallBackData) {
	defer a.closeConnect()
	result := models.M{}
	// get the template information.
	createReq := &client.CreateTemplateReq{}
	err := task.Value.Unmarshal(createReq)
	if err != nil {
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		return
	}
	// create the template
	createResp, err := a.appCenter.CreateTemplate(ctx, createReq)
	if err != nil {
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		return
	}
	id := createResp.ID
	// export AppInformation
	path, _, err := a.packAppData(ctx, createReq.AppID, task, handleData)
	if err != nil {
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		a.appCenter.DeleteTemplate(ctx, id)
		return
	}
	_, err = a.appCenter.FinishTemplate(ctx, id, path)
	if err != nil {
		err = error2.New(code.ErrInternalError)
		result[titleKey] = err.Error()
		a.sendMessage(err, result, task, handleData)
		a.appCenter.DeleteTemplate(ctx, id)
	}
	result[titleKey] = successCreateTemplate
	a.sendRate(task, handleData, 100)
	a.sendMessage(nil, result, task, handleData)
}

func (a *appData) getAppID(ctx context.Context, task *models.Task) (string, error) {
	req := convert(task)
	appIDInterface, ok := req[appIDKey]
	if !ok {
		return "", error2.New(code.ErrParameter)
	}
	appID, ok := appIDInterface.(string)
	if !ok {

		return "", error2.New(code.ErrParameter)
	}
	return appID, nil
}

func (a *appData) packAppData(ctx context.Context, appID string, task *models.Task, handleData chan *basal.CallBackData) (string, string, error) {

	md5Info := map[string]string{}
	zipInfo := make([]zip2.FileInfo, 0)

	// export app information
	appInfoResp, err := a.appCenter.ExportAppInfo(ctx, appID)
	if err != nil {
		return "", "", err
	}
	appName := appInfoResp.AppName
	appInfo, err := json.Marshal(appInfoResp)
	if err != nil {
		return "", "", error2.New(code.ErrInternalError)
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: appFileName,
		Body: []byte(appInfo),
	})
	md5Info[appFileName] = getMd5(appInfo)
	a.sendRate(task, handleData, 10)

	// export form
	tableInfo, err := a.structor.ExportTable(ctx, appID)
	if err != nil {
		return "", "", err
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: tableFileName,
		Body: []byte(tableInfo.JSONs),
	})
	md5Info[tableFileName] = getMd5([]byte(tableInfo.JSONs))
	a.sendRate(task, handleData, 20)

	// save custom page and custom page's md5
	for k, file := range tableInfo.PageFiles {
		zipInfo = append(zipInfo, zip2.FileInfo{
			Name: k,
			Body: file,
		})
		md5Info[k] = getMd5(file)
	}
	a.sendRate(task, handleData, 30)

	// export page engine
	pageEngines, err := a.persona.Export(ctx, appID)
	if err != nil {
		return "", "", err
	}
	pages, err := json.Marshal(pageEngines.AppData)
	if err != nil {
		return "", "", err
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: pageEngineFileName,
		Body: pages,
	})
	md5Info[pageEngineFileName] = getMd5(pages)
	a.sendRate(task, handleData, 35)

	// export permission
	perInfo, err := a.structor.ExportPermission(ctx, appID)
	if err != nil {
		return "", "", err
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: perFileName,
		Body: []byte(perInfo.JSONs),
	})
	md5Info[perFileName] = getMd5([]byte(perInfo.JSONs))
	a.sendRate(task, handleData, 40)

	// export api
	apiInfo, err := a.polyAPI.ExportAPI(ctx, appID)
	if err != nil {

		return "", "", err
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: apiFileName,
		Body: []byte(apiInfo.Data),
	})
	md5Info[apiFileName] = getMd5([]byte(apiInfo.Data))
	a.sendRate(task, handleData, 50)

	// export flow
	flowInfo, err := a.flow.ExportFlow(ctx, appID)
	if err != nil {

		return "", "", err
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: flowFileName,
		Body: []byte(flowInfo.Jsons),
	})
	md5Info[flowFileName] = getMd5([]byte(flowInfo.Jsons))
	a.sendRate(task, handleData, 60)
	// serialize md5 information
	md5InfoBytes, err := json.Marshal(md5Info)
	if err != nil {
		return "", "", error2.New(code.ErrInternalError)
	}
	// base64 encoding the md5 information
	md5Str := base64.StdEncoding.EncodeToString(md5InfoBytes)
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: fileInfo,
		Body: []byte(md5Str),
	})

	// package the all information
	buf, err := zip2.PackagingToBuffer(ctx, zipInfo)
	if err != nil {
		return "", "", error2.New(code.ErrInternalError)
	}
	a.sendRate(task, handleData, 70)
	fileName := fmt.Sprintf(fileNameTemp, appName)
	path := fmt.Sprintf(filePathTemp, time2.NowUnix(), appID, fileName)

	err = a.fileSDK.UploadFile(ctx, path, buf, int64(buf.Len()))
	if err != nil {
		return "", "", error2.New(code.ErrInternalError)
	}
	a.sendRate(task, handleData, 90)
	return path, fileName, nil

}

func (a *appData) insertAppData(ctx context.Context, appID string, task *models.Task, handleData chan *basal.CallBackData) error {

	// get the app file
	fileBuffer := new(bytes.Buffer)
	err := a.fileSDK.DownloadFile(ctx, task.FileAddr, fileBuffer)
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}
	file := fileBuffer.Bytes()

	// decompress the file
	fileInfos, err := zip2.DecompressToBytes(ctx, file)
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}

	/**
	Verify file MD5
	*/
	// get the file md5 map
	fileInfoBytes := fileInfos[fileInfo]
	if fileInfoBytes == nil || len(fileInfoBytes) == 0 {
		return error2.New(code.ErrFileFormat)
	}
	delete(fileInfos, fileInfo)
	// decode the md5 information by base 64
	md5Info, err := base64.StdEncoding.DecodeString(string(fileInfoBytes))
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}

	// deserialize the md5 information
	fileMd5 := map[string]string{}
	err = json.Unmarshal(md5Info, &fileMd5)
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}

	var tableInfo, perInfo, apiInfo, flowInfo, pageEngineInfo string
	var appBytes []byte
	// traverse the md5 map to get file information.if file not exist
	for k, v := range fileMd5 {
		fileBytes, ok := fileInfos[k]
		if !ok {
			return error2.New(code.ErrFileFormat)
		}
		if getMd5(fileBytes) != v {
			return error2.New(code.ErrFileFormat)
		}

		// get enum information
		switch k {
		case tableFileName:
			tableInfo = string(fileBytes)
			delete(fileInfos, k)
		case appFileName:
			appBytes = fileBytes
			delete(fileInfos, k)
		case perFileName:
			perInfo = string(fileBytes)
			delete(fileInfos, k)
		case apiFileName:
			apiInfo = string(fileBytes)
			delete(fileInfos, k)
		case flowFileName:
			flowInfo = string(fileBytes)
			delete(fileInfos, k)
		case pageEngineFileName:
			pageEngineInfo = string(fileBytes)
			delete(fileInfos, k)
		}
	}
	if appBytes == nil {
		return error2.New(code.ErrFileFormat)
	}
	appStruct := &struct {
		AppID   string `json:"AppID"`
		Version string `json:"version"`
	}{}
	err = json.Unmarshal(appBytes, appStruct)
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}
	oldAppID := appStruct.AppID

	// check version compatible
	err = a.appCenter.CheckVersion(ctx, appStruct.Version)
	if err != nil {
		return error2.New(code.ErrVersion)
	}

	a.sendRate(task, handleData, 10)
	var tableIDs map[string]string
	if tableInfo != "" {
		tableResp, err := a.structor.ImportTable(ctx, appID, oldAppID, tableInfo, fileInfos)
		if err != nil {
			return err
		}
		tableIDs = tableResp.TableIDs
	}
	a.sendRate(task, handleData, 30)
	logger.Logger.Info("---------------------------------------------")
	logger.Logger.Info(pageEngineInfo)
	logger.Logger.Info("---------------------------------------------")
	if pageEngineInfo != "" {
		pageEngineData := make([]*client.KV, 0)
		err := json.Unmarshal([]byte(pageEngineInfo), &pageEngineData)
		if err != nil {
			return err
		}
		data := replacePageKey(pageEngineData, tableIDs, oldAppID, appID)
		_, err = a.persona.Import(ctx, data)
		if err != nil {
			return err
		}
	}
	if perInfo != "" {
		_, err := a.structor.ImportPermission(ctx, appID, perInfo, tableIDs)
		if err != nil {
			return err
		}
	}
	a.sendRate(task, handleData, 50)
	if apiInfo != "" {
		_, err := a.polyAPI.ImportAPI(ctx, oldAppID, appID, apiInfo, tableIDs)
		if err != nil {
			return err
		}
	}
	a.sendRate(task, handleData, 70)
	if flowInfo != "" {
		_, err := a.flow.ImportFlow(ctx, appID, flowInfo, tableIDs)
		if err != nil {
			return err
		}
	}
	a.sendRate(task, handleData, 90)
	return nil
}

func (a *appData) sendMessage(err error, result models.M, task *models.Task, handleData chan *basal.CallBackData) {
	taskStatus := models.TaskSuccess
	if err != nil {
		taskStatus = models.TaskFail
	}
	task.Status = taskStatus
	task.Result = result
	task.FinishAt = time2.NowUnix()
	data := &basal.CallBackData{
		Task:  task,
		Types: basal.CallResult,
		Message: &basal.MesContent{
			TaskID: task.ID,
			Types:  basal.CallResult,
			Value:  taskStatus,
		},
	}
	handleData <- data
}

func (a *appData) sendRate(task *models.Task, handleData chan *basal.CallBackData, rate float64) {
	task.Ratio = rate
	data := &basal.CallBackData{
		Task:  task,
		Types: basal.CallRatio,
		Message: &basal.MesContent{
			TaskID: task.ID,
			Types:  basal.CallRatio,
			Value:  rate,
		},
	}
	handleData <- data
}

func (a *appData) closeConnect() {
	a.polyAPI.Close()
	a.flow.Close()
	a.structor.Close()
	a.appCenter.Close()
}

func getMd5(info []byte) string {
	md5info := md5.Sum(info)
	return hex.EncodeToString(md5info[:])
}

func replacePageKey(kvs []*client.KV, menus map[string]string, oldAppID, appID string) []*client.KV {
	datas := make([]*client.KV, 0, len(kvs))
	for _, kv := range kvs {
		newK := strings.ReplaceAll(kv.Key, oldAppID, appID)
		for k, v := range menus {
			if strings.Contains(newK, k) {
				newK = strings.ReplaceAll(newK, k, v)
				datas = append(datas, &client.KV{
					Key:   newK,
					Value: kv.Value,
				})
				break
			}
		}
	}
	return datas
}
