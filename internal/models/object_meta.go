package models

type ObjectMeta struct {
	Name      string
	TotalSize uint64
	Nodes     []int //nodes on which chunks would be stored
}
