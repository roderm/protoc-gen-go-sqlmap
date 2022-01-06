{{ if .Config.Read }}
func (s *{{ .StoreName }}) {{ .MsgName  }}(ctx context.Context, opts ...{{ .MsgName  }}Option) (map[interface{}]*{{ .MsgName  }}, error) {
	config := &query{{ .MsgName  }}Config{
		Store: s,
		filter: pg.NONE(),
		limit: 1000,
		rows: make(map[interface{}]*{{ .MsgName  }}),
	}
	for _, o := range opts {
		o(config)
	}
	err := s.select{{ .MsgName  }}(ctx, config, func(row *{{ .MsgName  }}) {
		config.rows[row.{{ GetPKName . }}] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}
	{{ range $i, $f := SubQueries . }}
	if config.load{{ $f.MsgName  }} {
		// {{ $f.FK.ChildOf.Table.GoPackageImport }}
		{{ if eq $f.Table.StoreName $f.FK.ChildOf.Table.StoreName }}
	 	_, err = s.{{ $f.FK.ChildOf.Table.MsgName  }}(ctx, config.opts{{ $f.MsgName }}...)
		{{ else }}
		store := {{ PackagePrefix $f.Table $f.FK.ChildOf.Table}}New{{ $f.FK.ChildOf.Table.StoreName }}(s.conn)
		_, err = store.{{ $f.FK.ChildOf.Table.MsgName  }}(ctx, config.opts{{ $f.MsgName }}...)
		{{ end }}
	}
	if err != nil {
	 	return config.rows, err
	}
	{{ end }}
	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *{{ .StoreName }}) Get{{ .MsgName }}SelectSqlString(filter pg.Where, limit int, start int) (string, []interface{}) {
	base := 0
	where, vals := pg.GetWhereClause(filter, &base)
	tpl := fmt.Sprintf(`
		SELECT "{{ getColumnNames .  "\", \"" }}"
		FROM "{{ .Name }}"
		%s`, where)

	if limit > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nLIMIT $%d", base)
		vals = append(vals, limit)
	}
	if start > 0 {
		base++
		tpl = tpl + fmt.Sprintf("\nOFFSET $%d", base)
		vals = append(vals, start)
	}
	return tpl, vals
}

func (s *{{ .StoreName }}) select{{ .MsgName }}(ctx context.Context, config *query{{ .MsgName  }}Config, withRow func(*{{ .MsgName  }})) error {
    query, vals := s.Get{{ .MsgName }}SelectSqlString(config.filter, config.limit, config.start)
	stmt, err := s.conn.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed preparing '%s' query in 'select{{ .MsgName }}': %s", query, err)
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return fmt.Errorf("failed executing query '%s' in 'select{{ .MsgName }}' (with %+v) : %s", query, vals, err)
	}
	defer cursor.Close()
	for cursor.Next() {
		row := &{{ .MsgName }}{
			{{- range $i, $join := $.Joins }}
			{{- if $join.IsRepeated }}
			{{ $join.TargetFieldName }}: []*{{ $join.SourcePackagePrefix}}{{ $join.SourceMessageName }}{},
			{{- else if $join.TargetIsOneOf }}
			{{- else }}
			{{ $join.TargetFieldName }}: new({{ $join.SourcePackagePrefix}}{{ $join.SourceMessageName }}),
			{{- end }}
			{{- end }}
		}
		err := cursor.Scan( &row.{{ getFieldNames . ", &row." }} )
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
{{ end }}
