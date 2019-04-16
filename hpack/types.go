package hpack

type headerFieldRepr byte

var indexed headerFieldRepr = headerFieldRepr(0)
var litrWithIndex headerFieldRepr = headerFieldRepr(1)
var litrWithoutIndex headerFieldRepr = headerFieldRepr(2)
var litrNeverIndexed headerFieldRepr = headerFieldRepr(3)
var dynTabSizeUpdate headerFieldRepr = headerFieldRepr(4)
var invalidHFRepr headerFieldRepr = headerFieldRepr(5)
