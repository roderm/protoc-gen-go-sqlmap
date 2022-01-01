{{ if .Config.Delete }}
func (m *{{ .MsgName  }}) Delete(s *{{ .StoreName }}, ctx context.Context) (error) {

	stmt, err := s.conn.PrepareContext(ctx, `
	DELETE FROM "{{ .Name }}"
	WHERE "{{ GetPKCol . }}" = $1
	RETURNING "{{ getColumnNames . "\", \"" }}"
		`)
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, m.{{ GetPKName . }})
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
