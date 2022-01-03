{{ if .Config.Update }}
func (m *{{ .MsgName  }}) Update(s *{{ .StoreName }}, ctx context.Context, conf *pg.UpdateSQL) (error) {
	base := {{ GetPrimaryBase . }}
	if conf == nil {
		conf = &pg.UpdateSQL{
			ValueMap: make(map[string]interface{}),
		}{{ range $i, $f := GetUpdateFields .}}
		conf.ValueMap["{{ $f.ColName }}"] = m.{{ $f.MsgName }}{{end}}
	}
	stmt, err := s.conn.PrepareContext(ctx, `
	UPDATE {{ .Name }} 
	SET ` + conf.String(&base) + `
	WHERE {{ GetPrimaryCols . }}
	RETURNING {{ getColumnNames . ", " }}
		`)
	if err != nil {
		return err
	}

	cursor, err := stmt.QueryContext(ctx, append([]interface{}{ {{ GetPrimaryValues . "m"}} }, conf.Values()...)... )
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
{{ end }}
