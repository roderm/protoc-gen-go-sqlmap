{{- range $i, $j := $.Joins }}
{{ if $j.IsRepeated }}
func {{ $j.TargetMessageName }}With{{ $j.TargetFieldName }}(opts ...{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}Option) {{ $j.TargetMessageName }}Option {
	return func(config *query{{ $j.TargetMessageName }}Config) {
		config.load{{ $j.TargetFieldName }} = true
		parent := make(map[interface{}]*{{ $j.TargetMessageName }})
		config.cb = append(config.cb, func(row *{{ $j.TargetMessageName }}) {
			child_key := row.{{ $j.TargetSourceKeyField }}
			parent[child_key] = row
		})
		config.opts{{ $j.TargetFieldName }} = append(opts,
			{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}Filter(
				pg.INCallabel("{{ $j.SourceColumnName }}", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}OnRow(func(row *{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}) {
					parent_id := row.{{ $j.SourceTargetKeyField}}
					if _, ok := parent[parent_id]; ok {
						parent[parent_id].{{ $j.TargetFieldName }} = append(parent[parent_id].{{ $j.TargetFieldName }}, row)
					}
				}),
		)
	}
}

{{- else }}
func {{ $j.TargetMessageName }}With{{ $j.TargetFieldName }}(opts ...{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}Option) {{ $j.TargetMessageName }}Option {
	return func(config *query{{ $j.TargetMessageName }}Config) {
		config.load{{ $j.TargetFieldName }} = true
		parent := make(map[interface{}][]*{{ $j.TargetMessageName }})
		config.cb = append(config.cb, func(row *{{ $j.TargetMessageName }}) {
			child_key := row.{{ $j.TargetSourceKeyField }}
			parent[child_key] = append(parent[child_key], row)
		})
		config.opts{{ $j.TargetFieldName }} = append(opts,
			{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}Filter(
				pg.INCallabel("{{ $j.SourceColumnName }}", func() []interface{} {
					ids := []interface{}{}
					for p := range parent {
						ids = append(ids, p)
					}
					return ids
				}),
			),
			{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}OnRow(func(row *{{ $j.SourcePackagePrefix}}{{ $j.SourceMessageName }}) {
					children := parent[row.{{ $j.SourceTargetKeyField}}]
					for _, c := range children {
						{{- if $j.TargetIsOneOf }}
						c.{{ $j.TargetOneOfField }} = &{{ $j.TargetMessageName }}_{{ $j.SourceMessageName }} { {{$j.SourceMessageName}}: row }
						{{- else}}
						c.{{ $j.TargetFieldName }} = row
						{{- end }}
					}
				}),
		)
	}
}

{{- end }}
{{- end }}
