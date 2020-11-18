package sqlgen

import (
	"text/template"
)

var selectTpl = `
func (s *{{ Store }}) {{ MessageName .  }}(ctx context.Context, opts ...{{ MessageName .  }}Option) (map[string]*{{ MessageName .  }}, error) {
	config := &query{{ MessageName .  }}Config{
		Store: s,
		filter: pg.NONE(),
		rows: make(map[string]*{{ MessageName .  }}),
	}
	for _, o := range opts {
		o(config)
	}

	err := s.select{{ MessageName .  }}(ctx, config.filter, func(row *{{ MessageName .  }}) {
		config.rows[row.Id] = row
		for _, cb := range config.cb {
			cb(row)
		}
	})
	if err != nil {
		return config.rows, err
	}
	{{ range $i, $f := SubQueries . }}
	if config.load{{ getFieldName $f  }} {
	 	_, err = s.{{ MessageName $f.FK.Remote.Table  }}(ctx, config.opts{{ getFieldName $f }}...)
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
func (s *{{ Store }}) select{{ MessageName . }}(ctx context.Context, filter pg.Where, withRow func(*{{ MessageName .  }})) error {
	where, vals := pg.GetWhereClause(filter)
	stmt, err := s.conn.PrepareContext(ctx, ` + "`" + `
	SELECT {{ getColumnNames .  ", " }} 
	FROM {{ TableName . }}
	` + "`" + `+where)
	if err != nil {
		return err
	}
	cursor, err := stmt.QueryContext(ctx, vals...)
	if err != nil {
		return err
	}
	defer cursor.Close()
	for cursor.Next() {
		row := new({{ MessageName .  }})
		err := cursor.Scan( &row.{{ getFieldNames . ", &row." }} )
		if err != nil {
			return err
		}
		withRow(row)
	}
	return nil
}
`

func LoadSelectTemplate(p Printer) *template.Template {
	tpl, err := template.New("Selects").Funcs(GetTemplateFuns(p)).Parse(selectTpl)
	if err != nil {
		panic(err)
	}
	return tpl
}
func (m *Table) Querier(g Printer) {
	err := LoadSelectTemplate(g).Execute(g, m)
	if err != nil {
		panic(err)
	}
}
