// Code generated by "enumer -type=PirType"; DO NOT EDIT.

//
package boosted

import (
	"fmt"
)

const _PirTypeName = "NoneMatrixPuncPermDPFNonPrivate"

var _PirTypeIndex = [...]uint8{0, 4, 10, 14, 18, 21, 31}

func (i PirType) String() string {
	if i < 0 || i >= PirType(len(_PirTypeIndex)-1) {
		return fmt.Sprintf("PirType(%d)", i)
	}
	return _PirTypeName[_PirTypeIndex[i]:_PirTypeIndex[i+1]]
}

var _PirTypeValues = []PirType{0, 1, 2, 3, 4, 5}

var _PirTypeNameToValueMap = map[string]PirType{
	_PirTypeName[0:4]:   0,
	_PirTypeName[4:10]:  1,
	_PirTypeName[10:14]: 2,
	_PirTypeName[14:18]: 3,
	_PirTypeName[18:21]: 4,
	_PirTypeName[21:31]: 5,
}

// PirTypeString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func PirTypeString(s string) (PirType, error) {
	if val, ok := _PirTypeNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to PirType values", s)
}

// PirTypeValues returns all values of the enum
func PirTypeValues() []PirType {
	return _PirTypeValues
}

// IsAPirType returns "true" if the value is listed in the enum definition. "false" otherwise
func (i PirType) IsAPirType() bool {
	for _, v := range _PirTypeValues {
		if i == v {
			return true
		}
	}
	return false
}
