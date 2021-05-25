module checklist

go 1.16

replace github.com/dkales/dpf-go => ./modules/dpf-go

require (
	github.com/dkales/dpf-go v0.0.0-20210304170054-6eae87348848
	github.com/elliotchance/orderedmap v1.4.0
	github.com/golang/protobuf v1.4.3
	github.com/lukechampine/fastxor v0.0.0-20210322201628-b664bed5a5cc
	github.com/paulbellamy/ratecounter v0.2.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rocketlaunchr/https-go v0.0.0-20200218083740-ba6c48f29f4d
	github.com/ugorji/go/codec v1.2.4
	github.com/zserge/metric v0.1.0
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf // indirect
	golang.org/x/mobile v0.0.0-20210220033013-bdb1ca9a1e08 // indirect
	golang.org/x/sys v0.0.0-20210503173754-0981d6026fa6 // indirect
	gotest.tools v2.2.0+incompatible
)
