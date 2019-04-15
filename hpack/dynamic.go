package hpack

type DynamicTable struct {
	maxSize      int
	HeaderFields []*HeaderField
}

func NewDynamicTable(size int) *DynamicTable {
	return &DynamicTable{
		maxSize:      size,
		HeaderFields: make([]*HeaderField, 0),
	}
}

func (dynT *DynamicTable) Add(hf *HeaderField) {
	dynT.HeaderFields = append(dynT.HeaderFields, []*HeaderField{hf}...)
}

func (dynT *DynamicTable) Length() int {
	return len(dynT.HeaderFields)
}

func (dynT *DynamicTable) Remove(index int) *HeaderField {
	removed := dynT.HeaderFields[index]

	if index < dynT.Length()-1 {
		dynT.HeaderFields = append(dynT.HeaderFields[:index], dynT.HeaderFields[(index+1):]...)
	} else {
		dynT.HeaderFields = dynT.HeaderFields[:index]
	}
	return removed
}
