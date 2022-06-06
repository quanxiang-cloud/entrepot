package logic

const (
	// CFormExport FormExport
	CFormExport Command = "formExport"
	// CFormImport 	 FormImport
	CFormImport Command = "formImport"
	// CFormTemplate   FormTemplate
	CFormTemplate Command = "formTemplate"
	// CAppImport CAppImport
	CAppImport Command = "appImport"
	// CAppExport CAppExport
	CAppExport Command = "appExport"
	// CCreateTemplate CCreateTemplate
	CCreateTemplate Command = "createTemplate"
	// CUseTemplate CUseTemplate
	CUseTemplate Command = "useTemplate"
)

// M map
type M map[string]interface{}

// Command  Command
type Command string

// HomeMap HomeMap
var HomeMap = map[Command]struct{}{
	CFormExport:   struct{}{},
	CFormImport:   struct{}{},
	CFormTemplate: struct{}{},
}
