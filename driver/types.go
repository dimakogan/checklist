package driver

import (
	"reflect"

	"checklist/pir"
	"checklist/updatable"
)

var registeredObjs = []interface{}{
	pir.PuncHintReq{},
	pir.PuncHintResp{},
	pir.PuncQueryReq{},
	pir.PuncQueryResp{},
	pir.DPFHintReq{},
	pir.DPFHintResp{},
	pir.DPFQueryReq{},
	pir.DPFQueryResp{},
	pir.MatrixHintReq{},
	pir.MatrixHintResp{},
	pir.MatrixQueryReq{},
	pir.MatrixQueryResp{},
	pir.NonPrivateHintReq{},
	pir.NonPrivateHintResp{},
	pir.NonPrivateQueryReq{},
	pir.NonPrivateQueryResp{},
	updatable.UpdatableHintReq{},
	updatable.UpdatableQueryReq{},
	updatable.UpdatableQueryResp{},
}

func RegisteredTypes() []reflect.Type {
	types := make([]reflect.Type, 0, len(registeredObjs))
	for _, obj := range registeredObjs {
		types = append(types, reflect.TypeOf(obj))
	}
	return types
}
