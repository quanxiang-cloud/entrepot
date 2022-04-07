package basal

import (
	"fmt"
	"github.com/quanxiang-cloud/entrepot/internal/logic"
	"github.com/quanxiang-cloud/entrepot/internal/models"
	"strconv"
)

const (
	// CallResult CallResult
	CallResult = "result"
	// CallRatio CallRatio
	CallRatio = "ratio"
)

// CallBackData CallBackData
type CallBackData struct {
	Types   string       // result ,ratio
	Task    *models.Task // task
	Message *MesContent  // send to message service
}

// MesContent MesContent
type MesContent struct {
	TaskID  string      `json:"taskID"`
	Command string      `json:"command"`
	Types   string      `json:"types"`
	Value   interface{} `json:"value"`
}

// ConvertSchema ConvertSchema
type ConvertSchema struct {
	Component      map[string]string
	Required       map[string]struct{}
	Properties     logic.M
	IndexComponent []SchemeIndex
}

//SchemeIndex SchemeIndex
type SchemeIndex struct {
	FieldKey  string
	FieldName string
	Index     int
}

// StrVal StrVal
func StrVal(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = fmt.Sprintf("%.2f", ft)
	case float32:
		ft := value.(float32)
		key = fmt.Sprintf("%.2f", ft)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	}
	return key
}

// DecimalFloat decimalFloat
func DecimalFloat(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

// Division division
func Division(num1 int, num2 int) float64 {
	result := float64(num1) / float64(num2) * 100
	return DecimalFloat(result)
}
