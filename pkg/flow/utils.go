package flow

import (
	"github.com/onflow/cadence"
	flowGo "github.com/onflow/flow-go-sdk"
)

type FreeflowDeposit flowGo.Event

func (evt FreeflowDeposit) ID() uint64 {
	return evt.Value.Fields[0].(cadence.UInt64).ToGoValue().(uint64)
}

func (evt FreeflowDeposit) Address() []byte {
	return evt.Value.Fields[1].(cadence.Optional).Value.(cadence.Address).Bytes()
}

// func (evt FreeflowDeposit) EventIndex() uint64 {
// 	return evt.Value.Fields[2].(cadence.UFix64).ToGoValue().(uint64)
// }
