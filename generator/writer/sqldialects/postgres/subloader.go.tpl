{{- range $i, $f := SubQueries . }}
func {{ $.MsgName }}With{{ $f.MsgName }}(opts ...{{ PackagePrefix $f.Table $f.FK.Remote.Table }}{{ $f.FK.Remote.Table.MsgName }}Option) {{$.MsgName }}Option {
	return func(config *query{{ $.MsgName }}Config) {
		config.load{{ $f.MsgName }} = true
		config.opts{{ $f.MsgName }} = opts
		{{- if $f.IsRepeated }}
		// Repeated
		parent := make(map[interface{}]*{{ $.MsgName }})
		config.cb = append(config.cb, func(row *{{ $.MsgName }}) {
			parent[row.{{ RepeatedFKFieldGetter $ $f }}] = row
		})
		config.opts{{ $f.MsgName }} = append(config.opts{{ $f.MsgName }}, 
			{{ PackagePrefix $f.Table $f.FK.Remote.Table }}{{ $f.FK.Remote.Table.MsgName }}OnRow(func(row *{{ PackagePrefix $f.Table $f.FK.Remote.Table }}{{ $f.FK.Remote.Table.MsgName }}) {
				if config.rows[row.{{ getFullFieldName $f.FK.Remote }}] != nil {
					config.rows[row.{{ getFullFieldName $f.FK.Remote }}].{{ $f.MsgName }} = append(config.rows[row.{{ getFullFieldName $f.FK.Remote }}].{{ $f.MsgName }}, row)
				}
			}),
		)
		{{- else if $f.IsMessage}}
		// Message
		parent := make(map[interface{}][]*{{ $.MsgName }})
		config.cb = append(config.cb, func(row *{{ $.MsgName }}) {
			_, ok := parent[row.{{ RepeatedFKFieldGetter $ $f }}]
			if !ok {
				parent[row.{{ RepeatedFKFieldGetter $ $f }}] = []*{{ $.MsgName }}{}
			}
			parent[row.{{ RepeatedFKFieldGetter $ $f }}] = append(parent[row.{{ RepeatedFKFieldGetter $ $f }}], row)
		})
		config.opts{{ $f.MsgName }} = append(config.opts{{ $f.MsgName }}, 
			{{ PackagePrefix $f.Table $f.FK.Remote.Table }}{{ $f.FK.Remote.Table.MsgName }}OnRow(
				func(row *{{ PackagePrefix $f.Table $f.FK.Remote.Table }}{{ $f.FK.Remote.Table.MsgName }}) {
				items, ok := parent[row.{{ MessageFKField $ $f }}]
				if ok && items != nil {
					for _, i := range items {
					{{- if ne $f.Oneof "" }}
						config.rows[i.{{ MessageFKItemField $ $f }}].{{ Title $f.Oneof }} = &{{ $.MsgName }}_{{ $f.MsgName }}{ {{$f.MsgName}}: row }
					{{- else }}
						config.rows[i.{{ MessageFKItemField $ $f }}].{{ $f.MsgName }} = row
					{{- end }}
					}
				}
			}),
		)
		{{- else}}
				// Unhandled
		{{- end }}

		config.opts{{ $f.MsgName }} = append(config.opts{{ $f.MsgName }}, 
			{{ PackagePrefix $f.Table $f.FK.Remote.Table }}{{ $f.FK.Remote.Table.MsgName }}Filter(
				pg.INCallabel("{{ $f.DbfkField }}", func() []interface{} {
					ids := []interface{}{}
					for id := range parent {
						ids = append(ids, id)
					}
					return ids
				},
				),
			),
		)
	}
}
{{- end }}
