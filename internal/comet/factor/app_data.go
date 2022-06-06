package factor

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	error2 "github.com/quanxiang-cloud/cabin/error"
	"github.com/quanxiang-cloud/cabin/logger"
	"github.com/quanxiang-cloud/cabin/tailormade/header"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/config"
	"github.com/quanxiang-cloud/entrepot/pkg/zip2"
	"github.com/quanxiang-cloud/fileserver/pkg/guide"
	"github.com/quanxiang-cloud/form/pkg/backup"
	"github.com/quanxiang-cloud/form/pkg/backup/aide"
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

const (
	staticNodeType    = "react-component"
	staticPackageName = "SimpleViewRenders"
	staticExportName  = "StaticViewRender"
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
	form      *backup.Backup
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
		form:      backup.NewBackup(conf.InternalNet),
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
		logger.Logger.Errorw("app-center success import error", err)
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
		logger.Logger.Errorw("app-center success import error", err)
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
		logger.Logger.Errorw("app-center finish template error", err)
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
		logger.Logger.Errorw("json marshal error", err)
		return "", "", error2.New(code.ErrInternalError)
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: appFileName,
		Body: appInfo,
	})
	md5Info[appFileName] = getMd5(appInfo)
	a.sendRate(task, handleData, 10)

	//formInfo := &bytes.Buffer{}

	// export form
	formInfo, err := a.form.Export(ctx, &aide.ExportOption{AppID: appID})
	if err != nil {
		return "", "", err
	}
	formBytes, err := json.Marshal(formInfo)
	if err != nil {
		return "", "", err
	}
	zipInfo = append(zipInfo, zip2.FileInfo{
		Name: tableFileName,
		Body: formBytes,
	})
	md5Info[tableFileName] = getMd5(formBytes)
	a.sendRate(task, handleData, 20)

	// export persona
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

	// export static file
	for _, kv := range pageEngines.AppData {
		kvNode := new(node)
		err = json.Unmarshal([]byte(kv.Value), kvNode)
		if err != nil {
			logger.Logger.Errorw("parsing persona data error", err)
			return "", "", err
		}
		if kvNode.Node != nil {
			path := getFileURL(kvNode)
			if path != "" {
				path = customPageURLReplace(path)
				file := &bytes.Buffer{}
				err = a.fileSDK.DownloadFile(ctx, path, file)
				if err != nil {
					logger.Logger.Errorw("download static ", err)
					continue
				}
				fileName := getFileName(path)
				zipInfo = append(zipInfo, zip2.FileInfo{
					Name: fileName,
					Body: file.Bytes(),
				})
				md5Info[fileName] = getMd5(file.Bytes())
			}
		}
	}

	// export api
	apiInfo, err := a.polyAPI.ExportAPI(ctx, appID)
	if err != nil {
		logger.Logger.Errorw("export api data error", err)
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
		logger.Logger.Errorw("export flow data error", err)
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
		logger.Logger.Errorw("md5 information json marshal error", err)
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
		logger.Logger.Errorw("package to zip error", err)
		return "", "", error2.New(code.ErrInternalError)
	}
	a.sendRate(task, handleData, 70)
	fileName := fmt.Sprintf(fileNameTemp, appName)
	path := fmt.Sprintf(filePathTemp, time2.NowUnix(), appID, fileName)

	err = a.fileSDK.UploadFile(ctx, path, buf, int64(buf.Len()))
	if err != nil {
		logger.Logger.Errorw("file server sdk error", err)
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
		logger.Logger.Errorw("file verify error")
		return error2.New(code.ErrFileFormat)
	}
	delete(fileInfos, fileInfo)
	// decode the md5 information by base 64
	md5Info, err := base64.StdEncoding.DecodeString(string(fileInfoBytes))
	if err != nil {
		logger.Logger.Errorw("base64 decode error")
		return error2.New(code.ErrFileFormat)
	}

	// deserialize the md5 information
	fileMd5 := map[string]string{}
	err = json.Unmarshal(md5Info, &fileMd5)
	if err != nil {
		return error2.New(code.ErrFileFormat)
	}

	var tableInfo, appBytes, apiInfo, flowInfo, pageEngineInfo []byte
	staticFileBytes := make([][]byte, 0)
	// traverse the md5 map to get file information.if file not exist
	for k, v := range fileMd5 {
		fileBytes, ok := fileInfos[k]
		if !ok {
			logger.Logger.Errorf("get file %s error", k)
			return error2.New(code.ErrFileFormat)
		}
		if getMd5(fileBytes) != v {
			logger.Logger.Errorf("get %s md5 error", k)
			return error2.New(code.ErrFileFormat)
		}

		// get enum information
		switch k {
		case tableFileName: // form data
			tableInfo = fileBytes
			delete(fileInfos, k)
		case appFileName: // app data
			appBytes = fileBytes
			delete(fileInfos, k)
		case apiFileName: // api data
			apiInfo = fileBytes
			delete(fileInfos, k)
		case flowFileName: // flow data
			flowInfo = fileBytes
			delete(fileInfos, k)
		case pageEngineFileName: // persona data
			pageEngineInfo = fileBytes
			delete(fileInfos, k)
		default:
			staticFileBytes = append(staticFileBytes, fileInfoBytes)
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
		logger.Logger.Errorw("json unmarshal app information error")
		return error2.New(code.ErrFileFormat)
	}
	oldAppID := appStruct.AppID

	// check version compatible
	err = a.appCenter.CheckVersion(ctx, appStruct.Version)
	if err != nil {
		logger.Logger.Errorw("version verify not pass")
		return error2.New(code.ErrVersion)
	}

	a.sendRate(task, handleData, 10)
	var tableIDs map[string]string
	if tableInfo != nil {
		formBytes := new(backup.Result)
		err := json.Unmarshal(tableInfo, &formBytes)
		if err != nil {
			logger.Logger.Errorw("Unmarshal form error", err)
			return err
		}
		_, err = a.form.Import(ctx, formBytes,
			&aide.ImportOption{
				AppID:    appID,
				UserID:   ctx.Value("User-Id").(string),
				UserName: ctx.Value("User-Name").(string),
			})
		if err != nil {
			logger.Logger.Errorw("import form data error", err)
			return err
		}
	}
	a.sendRate(task, handleData, 30)

	if pageEngineInfo != nil {
		pageEngineData := make([]*client.KV, 0)
		err := json.Unmarshal([]byte(pageEngineInfo), &pageEngineData)
		if err != nil {
			logger.Logger.Errorw("pageEngine json unmarshal error", err)
			return err
		}
		data := replacePageKey(pageEngineData, tableIDs, oldAppID, appID)
		_, err = a.persona.Import(ctx, data)
		if err != nil {
			logger.Logger.Errorw("import persona information error", err)
			return err
		}
	}

	a.sendRate(task, handleData, 50)
	if apiInfo != nil {
		_, err := a.polyAPI.ImportAPI(ctx, oldAppID, appID, string(apiInfo))
		if err != nil {
			logger.Logger.Errorw("import api information error", err)
			return err
		}
	}
	a.sendRate(task, handleData, 70)
	if flowInfo != nil {
		_, err := a.flow.ImportFlow(ctx, appID, string(flowInfo))
		if err != nil {
			logger.Logger.Errorw("import flow information error", err)
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
	a.appCenter.Close()
}

func getMd5(info []byte) string {
	md5info := md5.Sum(info)
	return hex.EncodeToString(md5info[:])
}

func getFileURL(nd *node) string {
	if nd.Children != nil && len(nd.Children) != 0 {
		for _, n := range nd.Children {
			url := getFileURL(n)
			if url != "" {
				return url
			}
		}
		return ""
	}
	if nd.Node != nil {
		n := nd.Node
		if n.PackageName == staticPackageName &&
			n.ExportName == staticExportName &&
			n.Type == staticNodeType {
			return n.Props["fileUrl"].Value.(string)
		}
	}
	return ""
}

func replacePageKey(kvs []*client.KV, menus map[string]string, oldAppID, appID string) []*client.KV {
	datas := make([]*client.KV, 0, len(kvs))
	for _, kv := range kvs {
		newK := strings.ReplaceAll(kv.Key, oldAppID, appID)
		datas = append(datas, &client.KV{
			Key:   newK,
			Value: kv.Value,
		})
	}
	return datas
}

type node struct {
	ID             string        `json:"id"`
	Type           string        `json:"type"`
	Path           string        `json:"path"`
	Name           string        `json:"name"`
	Node           *node         `json:"node"`
	Label          string        `json:"label"`
	PackageName    string        `json:"packageName"`
	PackageVersion string        `json:"packageVersion"`
	ExportName     string        `json:"exportName"`
	Children       []*node       `json:"children"`
	Props          map[string]tv `json:"props"`
}

type tv struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func customPageURLReplace(url string) string {
	url = strings.TrimPrefix(url, "/blob/")
	index := strings.LastIndex(url, ".html")
	if index < 0 {
		return ""
	}
	return url[0:index] + ".zip"
}

func getFileName(path string) string {
	idx := strings.LastIndex(path, "/")
	if idx != -1 {
		path = path[:idx]
	}
	idx = strings.LastIndex(path, "/")
	return path[idx+1:]
}

func getUserID(ctx context.Context) header.KV {
	_userID := "User-Id"
	i := ctx.Value(_userID)
	uid, ok := i.(string)
	if ok {
		return header.KV{_userID, uid}
	}
	return header.KV{_userID, "unexpected type"}
}

func getUserName(ctx context.Context) header.KV {
	_userName := "User-Name"
	i := ctx.Value(_userName)
	uName, ok := i.(string)
	if ok {
		return header.KV{_userName, uName}
	}
	return header.KV{_userName, "unexpected type"}
}
