package hpack

type dynamicTable struct {
	maxSize      uint32
	HeaderFields []*HeaderField
}

func NewDynamicTable(size uint32) *dynamicTable {
	return &dynamicTable{
		maxSize:      size,
		HeaderFields: make([]*HeaderField, 0),
	}
}

func (dynT *dynamicTable) add(hf *HeaderField) {
	dynT.HeaderFields = append(dynT.HeaderFields, []*HeaderField{hf}...)
}

func (dynT *dynamicTable) length() uint32 {
	return uint32(len(dynT.HeaderFields))
}

func (dynT *dynamicTable) get(index uint32) *HeaderField {
	if index < 0 || index >= uint32(len(dynT.HeaderFields)) {
		return nil
	}
	return dynT.HeaderFields[index]
}

func (dynT *dynamicTable) setMaxSize(size uint32) {
	dynT.maxSize = size
}

func (dynT *dynamicTable) remove(index uint32) *HeaderField {
	removed := dynT.HeaderFields[index]

	if index < dynT.length()-1 {
		dynT.HeaderFields = append(dynT.HeaderFields[:index], dynT.HeaderFields[(index+1):]...)
	} else {
		dynT.HeaderFields = dynT.HeaderFields[:index]
	}
	return removed
}
