// Code generated by "stringer -type=PresetEnum -linecomment -output preset_enum_string.go"; DO NOT EDIT.

package libzap

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[PresetUnknown-0]
	_ = x[PresetDevelopment-1]
	_ = x[PresetProduction-2]
}

const _PresetEnum_name = "unknowndevelopmentproduction"

var _PresetEnum_index = [...]uint8{0, 7, 18, 28}

func (i PresetEnum) String() string {
	if i < 0 || i >= PresetEnum(len(_PresetEnum_index)-1) {
		return "PresetEnum(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _PresetEnum_name[_PresetEnum_index[i]:_PresetEnum_index[i+1]]
}
