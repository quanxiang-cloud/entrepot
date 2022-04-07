package code

const (
	// NoError NoError
	NoError ErrorType = 1200100
	// ErrorSingleToMultiple ErrorSingleToMultiple
	ErrorSingleToMultiple ErrorType = 1200101
	// ErrorFieldIncorrect ErrorFieldIncorrect
	ErrorFieldIncorrect ErrorType = 1200102
	// ErrorDateTime ErrorDateTime
	ErrorDateTime ErrorType = 1200103
	// ErrorDepMultiple ErrorDepMultiple
	ErrorDepMultiple ErrorType = 1200104
	// ErrInter ErrInter
	ErrInter ErrorType = 1200105
	// ErrNumber  ErrNumber
	ErrNumber ErrorType = 1200106
	// ErrDataNotFind ErrDataNotFind
	ErrDataNotFind ErrorType = 1200107
	//ErrNotBlank ErrNotBlank
	ErrNotBlank ErrorType = 1200108
)

// ErrorType ErrorType
type ErrorType int

// ErrorTypeTable ErrorTypeTable
var ErrorTypeTable = map[ErrorType]string{
	NoError:               "没有错误",
	ErrorSingleToMultiple: "单选，但是给了多个值",
	ErrorFieldIncorrect:   "内部解析错误",
	ErrorDateTime:         "时间格式转换错误",
	ErrorDepMultiple:      "查询到多个部门",
	ErrInter:              "内部服务错误",
	ErrNumber:             "数字组件转换数字错误，不是给的数字",
	ErrDataNotFind:        "人员组件或者部门组件，数据未找到，或者找到多条数据",
	ErrNotBlank:           "该字段是非空字段，必须要给值",
}
