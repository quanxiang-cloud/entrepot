package factor

import (
	"context"
	error2 "github.com/quanxiang-cloud/cabin/error"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"reflect"
)

const (
	component      = "x-component"
	required       = "required"
	properties     = "properties"
	userID         = "user-id"
	userName       = "userName"
	depID          = "dep-id"
	editPermission = 2
)

//
// ComMap ComMap
var comMap = map[string]struct{}{
	"OrganizationPicker": struct{}{},
	"UserPicker":         struct{}{},
	"Input":              struct{}{},
	"Textarea":           struct{}{},
	"RadioGroup":         struct{}{},
	"CheckboxGroup":      struct{}{},
	"NumberPicker":       struct{}{},
	"DatePicker":         struct{}{},
	"Select":             struct{}{},
	"MultipleSelect":     struct{}{},
}

func getSchema(schema1 logic.M, st *basal.ConvertSchema) error {

	for key, value := range schema1 {
		isLayout := IsLayoutComponent(value)
		columnValue, err := GetAsMap(value)
		if err != nil {
			return err
		}
		if isLayout {
			if p, ok := columnValue["properties"]; ok {
				pro, err := GetAsMap(p)
				if err != nil {
					break
				}
				err = getSchema(pro, st)
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
			xRequired, err := GetMapToBool(columnValue, required)
			if err != nil {
				return err
			}
			if xRequired {
				return error2.New(code.ErrHighNoBlank)
			}
			continue
		}
		st.Properties[key] = value
		st.Component[key] = xCom
		xRequired, err := GetMapToBool(columnValue, required)
		if err != nil {
			return err
		}
		if xRequired {
			st.Required[key] = struct{}{}
		}
	}
	return nil

}

// GetAsMap GetAsMap
func GetAsMap(v interface{}) (map[string]interface{}, error) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	}
	return nil, error2.New(code.ErrParameter)
}

// GetMapToString GetMapToString
func GetMapToString(schema logic.M, key string) (string, error) {
	value, ok := schema[key]
	if !ok {
		return "", error2.New(code.ErrParameter)
	}
	if v, ok := value.(string); ok {
		return v, nil
	}
	return "", error2.New(code.ErrParameter)
}

// GetMapToBool GetMapToBool
func GetMapToBool(schema logic.M, key string) (bool, error) {
	value, ok := schema[key]
	if !ok {
		return false, nil
	}
	if v, ok := value.(bool); ok {
		return v, nil
	}
	return false, error2.New(code.ErrParameter)
}

//GetMapToInt GetMapToInt
func GetMapToInt(schema logic.M, key string) (int, error) {
	value, ok := schema[key]
	if !ok {
		return 0, nil
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Int:
		return value.(int), nil
	case reflect.Int32:
		return int(value.(int32)), nil
	case reflect.Int64:
		return int(value.(int64)), nil

	case reflect.Float64:
		return int(value.(float64)), nil
	case reflect.Float32:
		return int(value.(float32)), nil
	}
	return 0, error2.New(code.ErrParameter)
}

// GetMapToArrStr get map
func GetMapToArrStr(schema logic.M, key string) ([]string, error) {
	value, ok := schema[key]
	if !ok {
		return nil, nil
	}
	if v, ok := value.([]interface{}); ok {
		resp := make([]string, 0)
		for _, value := range v {
			if s, ok1 := value.(string); ok1 {
				resp = append(resp, s)
			}
		}
		return resp, nil

	}
	return nil, error2.New(code.ErrParameter)
}

// IsLayoutComponent IsLayoutComponent
func IsLayoutComponent(value interface{}) bool {
	switch reflect.TypeOf(value).Kind() {
	case reflect.Map:
		v := reflect.ValueOf(value)
		if value := v.MapIndex(reflect.ValueOf("x-internal")); value.IsValid() {
			if value.CanInterface() {
				return IsLayoutComponent(value.Interface())
			}
		}
		if value := v.MapIndex(reflect.ValueOf("isLayoutComponent")); value.IsValid() {
			if _, ok := value.Interface().(bool); ok {
				return value.Interface().(bool)
			}
		}
	default:
		return false
	}
	return false

}

func preCheckField(fieldKey []string, schema *basal.ConvertSchema) bool {
	for _, value := range fieldKey {
		_, ok := schema.Component[value]
		if !ok {
			return false
		}
	}
	return true
}

func preImportField(fieldKey []string, schema *basal.ConvertSchema) bool {
	for _, value := range fieldKey {
		_, ok := schema.Component[value]
		if !ok {
			return false
		}
		// Check read permission
		if !getEditPermission(schema.Properties[value]) {
			return false
		}
	}
	return true
}

// getEditPermission getEditPermission
func getEditPermission(value interface{}) bool {
	if value == nil {
		return false
	}
	switch reflect.TypeOf(value).Kind() {
	case reflect.Map:
		v := reflect.ValueOf(value)
		if value := v.MapIndex(reflect.ValueOf("x-internal")); value.IsValid() {
			if value.CanInterface() {
				return getEditPermission(value.Interface())
			}
		}
		if value := v.MapIndex(reflect.ValueOf("permission")); value.IsValid() {
			switch reflect.TypeOf(value.Interface()).Kind() {
			case reflect.Int:
				return value.Interface().(int)&editPermission != 0

			case reflect.Int32:
				return value.Interface().(int32)&editPermission != 0
			case reflect.Int64:
				return value.Interface().(int64)&editPermission != 0
			case reflect.Float64:
				return int64(value.Interface().(float64))&editPermission != 0
			case reflect.Float32:
				return int64(value.Interface().(float32))&editPermission != 0
			default:
				return false
			}
		}
	default:
		return false
	}
	return false

}

func convert(task *models.Task) logic.M {
	l := make(logic.M)
	for key, value := range task.Value {
		l[key] = value
	}
	return l
}

func failCallBack(task *models.Task, callBack chan *basal.CallBackData, result models.M) {
	task.Status = models.TaskFail
	task.Result = result
	task.FinishAt = time2.NowUnix()
	data := &basal.CallBackData{
		Task:  task,
		Types: basal.CallResult,
		Message: &basal.MesContent{
			TaskID: task.ID,
			Types:  basal.CallResult,
			Value:  models.TaskFail,
		},
	}
	callBack <- data
}

func successCallBack(task *models.Task, callBack chan *basal.CallBackData, result models.M) {
	task.Status = models.TaskSuccess
	task.Result = result
	task.FinishAt = time2.NowUnix()
	data := &basal.CallBackData{
		Task:  task,
		Types: basal.CallResult,
		Message: &basal.MesContent{
			TaskID: task.ID,
			Types:  basal.CallResult,
			Value:  models.TaskSuccess,
		},
	}
	callBack <- data

}

func ratioCallBack(task *models.Task, callBack chan *basal.CallBackData, ratio float64) {
	task.Ratio = ratio
	data := &basal.CallBackData{
		Task:  task,
		Types: basal.CallRatio,
		Message: &basal.MesContent{
			TaskID: task.ID,
			Types:  basal.CallRatio,
			Value:  ratio,
		},
	}
	callBack <- data

}

// ctxCopy ctxCopy
func ctxCopy(c context.Context, data map[string]string) context.Context {
	var (
		_userID       interface{} = "User-Id"
		_userName     interface{} = "User-Name"
		_departmentID interface{} = "Department-Id"
	)
	if c == nil {
		c = context.Background()
	}
	c = context.WithValue(c, _userID, data[userID])
	c = context.WithValue(c, _departmentID, data[depID])
	c = context.WithValue(c, _userName, data[userName])
	return c
}

func getGetMapTask(task *models.Task) map[string]string {
	resp := make(map[string]string)
	resp[userID] = task.CreatorID
	resp[depID] = task.DepID
	resp[userName] = task.CreatorName
	return resp
}
