{{ if .Config.Create }}
func (m *{{ .MsgName  }}) Insert(s *{{ .StoreName }}, ctx context.Context) (error) {
	ins := pg.NewInsert()
	ins.Add({{ GetInsertFieldNames . "m" "," }})

	stmt, err := s.conn.PrepareContext(ctx, `
		INSERT INTO "{{ .Name }}" ( "{{ GetInsertColNames .  "\", \"" }}" )
		VALUES ` + ins.String(nil) + `
		RETURNING "{{ getColumnNames . "\", \"" }}"
		`)
	
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, ins.Values()...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		err := cursor.Scan( &m.{{ getFieldNames . ", &m." }} )
		if err != nil {
			return err
		}
	}
	return nil
}
{{end}}
