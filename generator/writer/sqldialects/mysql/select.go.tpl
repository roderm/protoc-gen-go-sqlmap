func (s *{{ .StoreName }}) {{ .MsgName  }}(ctx context.Context, opts ...{{ .MsgName  }}Option) ({{ .MsgName }}List, error) {
	config := &query{{ .MsgName  }}Config{
		Store: s,
		filter: squirrel.And{},
		limit: 1000,
		rows: make({{ .MsgName }}List),
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
	{{- range $i, $f := SubQueries . }}
	if config.load{{ $f.MsgName  }} {
		{{- if eq $f.Table.StoreName $f.FK.ChildOf.Table.StoreName }}
	 	_, err = s.{{ $f.FK.ChildOf.Table.MsgName  }}(ctx, config.opts{{ $f.MsgName }}...)
		{{- else }}
		store := {{ PackagePrefix $f.Table $f.FK.ChildOf.Table}}New{{ $f.FK.ChildOf.Table.StoreName }}(s.conn)
		_, err = store.{{ $f.FK.ChildOf.Table.MsgName  }}(ctx, config.opts{{ $f.MsgName }}...)
		{{- end }}
		if err != nil {
			 return config.rows, err
		}
	}
	{{- end }}
	for _, cb := range config.beforeReturn {
		err = cb(config.rows)
		if err != nil {
			return config.rows, err
		}
	}
	return config.rows, nil
}

func (s *{{ .StoreName }}) Get{{ .MsgName }}SelectSqlString(filter []squirrel.Sqlizer, limit int, start int) squirrel.SelectBuilder {
	q := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(`{{ getColumnNames .  ", " }}`).
		From("\"{{ .Name }}\"").
		Where(append(squirrel.And{}, filter...))
	if limit > 0 {
		q.Limit(uint64(limit))
	}
	if start > 0 {
		q.Offset(uint64(limit))
	}
	return q
}

func (s *{{ .StoreName }}) select{{ .MsgName }}(ctx context.Context, config *query{{ .MsgName  }}Config, withRow func(*{{ .MsgName  }})) error {
    query := s.Get{{ .MsgName }}SelectSqlString(config.filter, config.limit, config.start)
	// cursor, err := query.RunWith(s.conn).QueryContext(ctx)
	sql, params, _ := query.ToSql()
	cursor, err := s.conn.QueryxContext(ctx, sql, params...)
	if err != nil {
		return fmt.Errorf("failed executing query '%+v' in 'select{{ .MsgName }}': %s", query, err)
	}
	defer cursor.Close()
	resultRows := []*{{ .MsgName }}{}
	for cursor.Next() {
		row := new({{ .MsgName }})
		err = cursor.StructScan(row)
		if err == nil {
			withRow(row)
			resultRows = append(resultRows, row)
		} else {
			return fmt.Errorf("sqlx.StructScan failed: %s", err)
		}

	}
	// err = sqlx.StructScan(cursor, &resultRows)
	// if err != nil {
	// 	return fmt.Errorf("sqlx.StructScan failed: %s", err)
	// }
	// for _, row := range resultRows {
	// 	withRow(row)
	// }
	return nil
}
