package enums

import "fmt"

type SessionStatus int

const (
	Active SessionStatus = iota
	Inactive
	Terminated
)

var SessionStatusName = map[SessionStatus]string{
	Active:     "ACTIVE",
	Inactive:   "INACTIVE",
	Terminated: "TERMINATED",
}

var SessionStatusValue = map[string]SessionStatus{
	"ACTIVE":     Active,
	"INACTIVE":   Inactive,
	"TERMINATED": Terminated,
}

func (s SessionStatus) String() string {
	return SessionStatusName[s]
}

func ToSessionStatus(s string) (SessionStatus, error) {
	if status, ok := SessionStatusValue[s]; !ok {
		return Active, fmt.Errorf("unknown value %s", s)
	} else {
		return status, nil
	}
}

type PaneType int

const (
	HSplit PaneType = iota
	VSplit
	Leaf
)

var PaneTypeName = map[PaneType]string{
	HSplit: "HSPLIT",
	VSplit: "VSPLIT",
	Leaf:   "LEAF",
}

var PaneTypeValue = map[string]PaneType{
	"HSPLIT": HSplit,
	"VSPLIT": VSplit,
	"LEAF":   Leaf,
}

func (p PaneType) String() string {
	return PaneTypeName[p]
}

func ToPaneType(s string) (PaneType, error) {
	if paneType, ok := PaneTypeValue[s]; !ok {
		return Leaf, fmt.Errorf("unknown value %s", s)
	} else {
		return paneType, nil
	}
}
