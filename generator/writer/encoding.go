package writer

import (
	"text/template"

	"github.com/roderm/protoc-gen-go-sqlmap/generator/types"
)

var structEncoding = `
type {{ .MsgName }}List map[interface{}]*{{ .MsgName }}

{{- if .Config.JSONB }}
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
{{- else }}
func (m *{{ .MsgName }}) GetSqlmapPK() interface{} {
	pk := map[string]interface{}{
		{{- range $i, $f := GetPrimaries . }}
		"{{ $f.ColName }}": m.{{ $f.MsgName }},
		{{- end }}
	}
	return pk
}
func (m *{{ .MsgName }}) Scan(value interface{}) error {
	buff, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("Failed %+v", value)
	}
	err := json.Unmarshal(buff, m)
	if err != nil {
		return fmt.Errorf("Unmarshal '%s' => '{{ .MsgName }}' failed: %s", string(buff), err)
	}
	return nil
}
{{- end}}
`

/*

func (m *{{ .MsgName }}) Value() (driver.Value, error) {
	pk := struct {
		{{- range $i, $f := GetPrimaries . }}
		{{- if eq $f.Type "message" }}
		{{ $f.MsgName }} *{{ $f.TypeMessage }} ` + "`" + `json:"{{ $f.ColName }}"` + "`" + `
		{{- else }}
		{{ $f.MsgName }} {{ $f.Type }} ` + "`" + `json:"{{ $f.ColName }}"` + "`" + `
		{{- end }}
		{{- end }}
	}{
		{{- range $i, $f := GetPrimaries . }}
		m.{{ $f.MsgName }},
		{{- end }}
	}
	return json.Marshal(pk)
} */

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
