// Package StockCommands
// Dafny module StockCommands compiled into Go

package StockCommands

import (
	os "os"

	m_CartCommands "wholesale.local/orbit/generated/CartCommands"
	m_CommandModel "wholesale.local/orbit/generated/CommandModel"
	m_Domain "wholesale.local/orbit/generated/Domain"
	m_OrderCommands "wholesale.local/orbit/generated/OrderCommands"
	m__System "wholesale.local/orbit/generated/System_"
	_dafny "wholesale.local/orbit/generated/dafny"
)

var _ = os.Args
var _ _dafny.Dummy__
var _ m__System.Dummy__
var _ m_Domain.Dummy__
var _ m_CommandModel.Dummy__
var _ m_CartCommands.Dummy__
var _ m_OrderCommands.Dummy__

type Dummy__ struct{}

// Definition of class Default__
type Default__ struct {
	dummy byte
}

func New_Default___() *Default__ {
	_this := Default__{}

	return &_this
}

type CompanionStruct_Default___ struct {
}

var Companion_Default___ = CompanionStruct_Default___{}

func (_this *Default__) Equals(other *Default__) bool {
	return _this == other
}

func (_this *Default__) EqualsGeneric(x interface{}) bool {
	other, ok := x.(*Default__)
	return ok && _this.Equals(other)
}

func (*Default__) String() string {
	return "StockCommands.Default__"
}
func (_this *Default__) ParentTraits_() []*_dafny.TraitID {
	return [](*_dafny.TraitID){}
}

var _ _dafny.TraitOffspring = &Default__{}

func (_static *CompanionStruct_Default___) AdjustStock(input m_CommandModel.CommandInput, product _dafny.Sequence, version _dafny.Int) m_CommandModel.CommandPlan {
	if !(m_CommandModel.Companion_Default___.VersionMatches((input).Dtor_expectedVersion(), version)) {
		return m_CommandModel.Companion_CommandPlan_.Create_Reject_(m_CommandModel.Companion_Default___.StaleVersion())
	} else {
		return m_CommandModel.Companion_CommandPlan_.Create_Adjust_(product, (input).Dtor_stockDelta())
	}
}

// End of class Default__
