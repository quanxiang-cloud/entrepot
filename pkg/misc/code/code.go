package code

import (
	error2 "github.com/quanxiang-cloud/cabin/error"
)

func init() {
	error2.CodeTable = CodeTable
}

const (
	// ErrNotDelete ErrNotDelete
	ErrNotDelete = 130014000001
	// ErrNotField ErrNotField
	ErrNotField = 130014000002
	// ErrNotRepeat ErrNotRepeat
	ErrNotRepeat = 130014000003
	// ErrHighNoBlank ErrHighNoBlank
	ErrHighNoBlank = 130014000004
	// ErrHighNoData ErrHighNoData
	ErrHighNoData = 130014000005
	// ErrParameter ErrParameter
	ErrParameter = 120024000001

	// ErrNoAuth ErrNoAuth
	ErrNoAuth = 120034000003

	// ErrInternalError ErrInternalError
	ErrInternalError = 120044000001
	// ErrFile ErrFile
	ErrFile = 120044000002
	// ErrUploadFile ErrUploadFile
	ErrUploadFile = 120044000003
	// ErrVersion ErrVersion
	ErrVersion = 120044000004
	// ErrSystemNet ErrSystemNet
	ErrSystemNet = 120044000005
	// ErrFileFormat ErrFileFormat
	ErrFileFormat = 120044000006

	// ErrTemplateNotExist ErrTemplateNotExist
	ErrTemplateNotExist = 120054000001

	// ErrParamFormat ErrParamFormat
	ErrParamFormat = 120064000001
)

// CodeTable CodeTable
var CodeTable = map[int64]string{
	ErrNotDelete:        "任务在处理中，不能删除",
	ErrNotField:         "该表单没有可以导入导出的字段",
	ErrNotRepeat:        "已经有相同的任务在进行，等任务完成，再创建同样的任务",
	ErrHighNoBlank:      "导入的表单,除了支持的导入字段,有必填字段",
	ErrHighNoData:       "导出的数据为空",
	ErrParameter:        "参数类型错误",
	ErrInternalError:    "内部服务错误",
	ErrFile:             "上传的文件不符合规则",
	ErrNoAuth:           "权限不足",
	ErrUploadFile:       "上传文件错误",
	ErrSystemNet:        "系统网络错误",
	ErrFileFormat:       "文件格式错误！",
	ErrVersion:          "导入版本不兼容",
	ErrTemplateNotExist: "模板不可用",
	ErrParamFormat:      "参数错误",
}
