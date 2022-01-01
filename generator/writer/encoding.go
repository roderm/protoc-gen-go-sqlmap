package writer

import (
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
)

var structEncoding = `
{{if .JSONB }}
func (m *{{ .MsgName }}) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	return json.Unmarshal(buff, m)
}

func (m *{{ .MsgName }}) Value() (driver.Value, error) {
	return json.Marshal(m)
}
{{else if HasPK .}}
func (m *{{ .MsgName }}) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	m.{{ GetPKName . }} = {{ GetPKConvert . "buff" }}
	return nil
}

func (m *{{ .MsgName }}) Value() (driver.Value, error) {
	return m.{{ GetPKName . }}, nil
}

func (m *{{ .MsgName }}) GetIdentifier() interface{} {
	return m.{{ GetPKName . }}
}
{{end}}
`

func writeEncodings(p Printer) *template.Template {
	funcs := GetTemplateFuns()
	tpl, err := template.New("SQLEncodings").Funcs(funcs).Parse(structEncoding)
	if err != nil {
		panic(err)
	}
	return tpl
}

func WriteEncodings(g Printer, m *types.Table) {
	err := writeEncodings(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
