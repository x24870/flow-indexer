package flow

import (
	"github.com/onflow/cadence"
	flowGo "github.com/onflow/flow-go-sdk"
)

const (
	FreeflowDepositEventType  = "A.88dd257fcf26d3cc.Inscription.Deposit"
	FreeflowWithdrawEventType = "A.88dd257fcf26d3cc.Inscription.Withdraw"
)

type FreeflowDeposit flowGo.Event

func (evt FreeflowDeposit) ID() uint64 {
	return evt.Value.Fields[0].(cadence.UInt64).ToGoValue().(uint64)
}

func (evt FreeflowDeposit) Address() []byte {
	return evt.Value.Fields[1].(cadence.Optional).Value.(cadence.Address).Bytes()
}

type FreeflowWithdraw flowGo.Event

func (evt FreeflowWithdraw) ID() uint64 {
	return evt.Value.Fields[0].(cadence.UInt64).ToGoValue().(uint64)
}

func (evt FreeflowWithdraw) Address() []byte {
	return evt.Value.Fields[1].(cadence.Optional).Value.(cadence.Address).Bytes()
}
