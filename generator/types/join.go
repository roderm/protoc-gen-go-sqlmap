package types

type Join struct {
	IsRepeated bool
	Target     *Field
	Source     *Field

	TargetMessageName string
	TargetFieldName   string

	TargetSourceKeyField string

	SourceMessageName string
	SourceFieldName   string

	SourceColumnName     string
	SourceTargetKeyField string
	SourcePackagePrefix  string

	TargetIsOneOf    bool
	TargetOneOfField string
	TargetOneOfType  string
}
