package sqlgen

func GetRemoteFieldname(remote *Table, data *Table) string {
	for _, f := range data.Cols {
		if remote.Name == f.dbfkTable {
			for _, rf := range remote.Cols {
				if rf.ColName == f.dbfkField {
					return rf.desc.GetName()
				}
			}
		}
	}
	return ""
}
func GetRemoteListName(remote *Table, data *Table) string {
	for _, f := range data.Cols {
		if remote.Name == f.dbfkTable {
			for _, rf := range remote.Cols {
				if rf.ColName == f.dbfkField {
					return rf.desc.GetName()
				}
			}
		}
	}
	return ""
}
