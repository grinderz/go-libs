// Code generated by "stringer -type=EncodingEnum -linecomment -output encoding_enum_string.go"; DO NOT EDIT.

package libzap

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EncodingUnknown-0]
	_ = x[EncodingConsole-1]
	_ = x[EncodingJSON-2]
}

const _EncodingEnum_name = "unknownconsolejson"

var _EncodingEnum_index = [...]uint8{0, 7, 14, 18}

func (i EncodingEnum) String() string {
	if i < 0 || i >= EncodingEnum(len(_EncodingEnum_index)-1) {
		return "EncodingEnum(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EncodingEnum_name[_EncodingEnum_index[i]:_EncodingEnum_index[i+1]]
}