package redesign

type Field struct {
	Name   string
	Column string
	Table  *Table
}

type MessageField struct {
	Field
	IsRepeated bool
	Target     Field
}
