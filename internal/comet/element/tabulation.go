package element

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/quanxiang-cloud/cabin/logger"
	header2 "github.com/quanxiang-cloud/cabin/tailormade/header"
	time2 "github.com/quanxiang-cloud/cabin/time"
	"github.com/quanxiang-cloud/entrepot/internal/comet/basal"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/pkg/client"
	"github.com/quanxiang-cloud/entrepot/pkg/misc/code"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	splitSign = "|"
)

// Matrix Matrix
type Matrix interface {
	GetTag() string
	SetValue(fieldKey string, schema *basal.ConvertSchema, orgAPI client.OrgAPI, orgMap map[string]string)
	GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType)
	ExportValue(ctx context.Context, value interface{}) string
}

var (
	// ErrNoMatrix no matrix
	ErrNoMatrix = errors.New("no ErrNoMatrix like this")
)

var matrix = []Matrix{
	&organizationPicker{},
	&userPicker{},

	&input{},         // Single-line text  Type:string
	&textarea{},      // Multiline text   Type:string
	&radioGroup{},    // Radio buttons   Type:string
	&checkboxGroup{}, // Check box    Type:array

	&numberPicker{},   // Number    Type:number
	&datePicker{},     // Date   Type:standard time format
	&selects{},        // A drop-down list box is displayed   Type:string
	&multipleSelect{}, // Drop-down box     Type:string
}

// Component Component
type Component struct {
	numerator map[string]Matrix
}

// NewComponent NewComponent
func NewComponent() *Component {
	m := &Component{
		numerator: make(map[string]Matrix, len(matrix)),
	}
	for _, n := range matrix {
		m.numerator[n.GetTag()] = n
	}
	return m
}

// GetMold mold a component
func (c *Component) GetMold(tag, fieldName string, schema *basal.ConvertSchema, orgAPI client.OrgAPI, orgMap map[string]string) (Matrix, error) {
	numerator, ok := c.numerator[tag] // get the component
	if !ok {
		return nil, nil
	}
	numerator.SetValue(fieldName, schema, orgAPI, orgMap)
	return numerator, nil
}

type base struct {
	schema   *basal.ConvertSchema
	fieldKey string
	orgAPI   client.OrgAPI
	orgMap   map[string]string
}

func (b *base) SetValue(fieldKey string, schema *basal.ConvertSchema, orgAPI client.OrgAPI, orgMap map[string]string) {
	b.fieldKey = fieldKey
	b.schema = schema
	b.orgAPI = orgAPI
	b.orgMap = orgMap

}

func (b *base) ExportValue(ctx context.Context, value interface{}) string {
	v1, ok := value.(string)
	if ok {
		return v1
	}
	return ""
}

func (b *base) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	return nil, 0
}

type organizationPicker struct {
	base
}

// GetValueByString return the error type
func (o *organizationPicker) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	orgArr := strings.Split(value, splitSign)
	respArr := make([]logic.M, 0)
	for _, value1 := range orgArr {
		id, ok := o.orgMap[value1]
		if !ok {
			return respArr, code.ErrDataNotFind
		}
		v := strings.Split(value1, "/")
		v1 := v[len(v)-1]
		entity1 := logic.M{
			"label": v1,
			"value": id,
		}
		respArr = append(respArr, entity1)
	}

	return respArr, code.NoError

}

func (o *organizationPicker) GetTag() string {
	return "OrganizationPicker"
}

func cExportValue(value interface{}) string {
	arr, ok := value.([]interface{})
	if !ok {
		return ""
	}
	s1 := make([]string, 0)
	var bt bytes.Buffer
	for _, v1 := range arr {
		// Check whether the interface type is Map
		switch reflect.TypeOf(v1).Kind() {
		case reflect.Map:
			v := reflect.ValueOf(v1)
			if v2 := v.MapIndex(reflect.ValueOf("label")); v2.IsValid() {
				if reflect.TypeOf(v2.Interface()).Kind() == reflect.String {
					s := v2.Interface().(string)
					s1 = append(s1, s)
				}

			}
		}

	}
	for index, v := range s1 {
		if index == len(s1)-1 {
			bt.WriteString(v)
			continue
		}
		bt.WriteString(fmt.Sprintf("%s|", v))

	}
	return bt.String()
}

func (o *organizationPicker) ExportValue(ctx context.Context, value interface{}) string {
	return cExportValue(value)
}

type userPicker struct {
	base
}

func (o *userPicker) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	//userArr := strings.Split(value, splitSign)
	//respArr := make([]logic.M, 0)
	//for _, userName := range userArr {
	//	respUser, err := o.orgAPI.GetUserList(ctx, userName)
	//	if err != nil {
	//		return nil, code.ErrDataNotFind
	//	}
	//	if len(respUser) != 1 {
	//		return nil, code.ErrDataNotFind
	//	}
	//	entity1 := logic.M{
	//		"label": userName,
	//		"value": respUser[0].ID,
	//	}
	//	respArr = append(respArr, entity1)
	//}
	//return respArr, code.NoError
	return nil, 0
}

func (o *userPicker) GetTag() string {
	return "UserPicker"
}

func (o *userPicker) ExportValue(ctx context.Context, value interface{}) string {
	return cExportValue(value)
}

type checkboxGroup struct {
	base
}

func (c *checkboxGroup) GetTag() string {
	return "CheckboxGroup"
}

func (c *checkboxGroup) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	split := strings.Split(value, splitSign)
	arr := make([]string, len(split))
	for index, data := range split {
		arr[index] = data
	}
	return arr, code.NoError
}

func arrStringExportValue(value interface{}) string {
	arr, ok := value.([]interface{})
	if !ok {
		return ""
	}
	s1 := make([]string, 0)
	var bt bytes.Buffer
	for _, v1 := range arr {
		s, ok := v1.(string)
		if ok {
			s1 = append(s1, s)
		}
	}
	for index, v := range s1 {
		if index == len(s1)-1 {
			bt.WriteString(v)
			continue
		}
		bt.WriteString(fmt.Sprintf("%s|", v))

	}
	return bt.String()

}

func (c *checkboxGroup) ExportValue(ctx context.Context, value interface{}) string {
	return arrStringExportValue(value)
}

type datePicker struct {
	base
}

func (d *datePicker) ExportValue(ctx context.Context, value interface{}) string {
	v1, ok := value.(string)
	if !ok {
		return ""
	}
	timezone := header2.GetTimezone(ctx)
	_, v := timezone.Wreck()
	tolerant, err := time2.Tolerant(v)
	if err != nil {
		logger.Logger.Errorw("err  is Tolerant")
		return ""
	}
	revise, err := time2.Revise(v1, tolerant)
	if err != nil {
		logger.Logger.Errorw("err  is  time2.Revise")
		return ""
	}
	parse, err := time.Parse(time2.ISO8601, revise)
	if err != nil {
		return ""
	}
	return parse.Format("2006-01-02 15:04:05")
}

func (d *datePicker) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	any, err := dateparse.ParseAny(value)
	if err != nil {
		return nil, code.ErrorDateTime
	}
	timezone := header2.GetTimezone(ctx)
	_, v := timezone.Wreck()
	tolerant, err2 := time2.Tolerant(v)
	if err2 != nil {
		logger.Logger.Errorw("err  is  time2.Tolerant")
		return nil, code.ErrorDateTime
	}
	regular, err := time2.Regular(any.Format(time2.ISO8601), tolerant)
	if err != nil {
		logger.Logger.Errorw("err  is  time2.Regular")
		return nil, code.ErrorDateTime
	}
	return regular, code.NoError
}

func (d *datePicker) GetTag() string {
	return "DatePicker"
}

type input struct {
	base
}

func (i *input) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	return value, code.NoError
}

func (i *input) GetTag() string {
	return "Input"
}

type radioGroup struct {
	base
}

// GetValueByString Radio buttons
func (r *radioGroup) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	return value, code.NoError

}

func (r *radioGroup) GetTag() string {
	return "RadioGroup"
}

type selects struct {
	base
}

func (s *selects) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	return value, code.NoError
}

func (s *selects) GetTag() string {
	return "Select"
}

type textarea struct {
	base
}

func (t *textarea) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	return value, code.NoError
}

func (t *textarea) GetTag() string {
	return "Textarea"
}

type numberPicker struct {
	base
}

func (t *numberPicker) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	value1, err := strconv.Atoi(value)
	if err == nil {
		return value1, code.NoError
	}
	f64, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, code.ErrNumber
	}
	return f64, code.NoError
}

func (t *numberPicker) ExportValue(ctx context.Context, value interface{}) string {
	return basal.StrVal(value)
}

func (t *numberPicker) GetTag() string {
	return "NumberPicker"
}

type multipleSelect struct {
	base
}

func (m *multipleSelect) ExportValue(ctx context.Context, value interface{}) string {
	return arrStringExportValue(value)
}

func (m *multipleSelect) GetTag() string {
	return "MultipleSelect"
}

func (m *multipleSelect) GetValueByString(ctx context.Context, value string) (interface{}, code.ErrorType) {
	split := strings.Split(value, splitSign)
	arr := make([]string, len(split))
	for index, data := range split {
		arr[index] = data
	}
	return arr, code.NoError
}
