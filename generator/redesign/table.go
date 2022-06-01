package redesign

type Table struct {
	Name  string
	Table string

	GoPackageName   string
	GoPackageImport string
	Engine          string

	Columns []*Field
	Configs Configs
}

type Configs struct {
	JSONB  bool
	Create bool
	Read   bool
	Update bool
	Delete bool
}
