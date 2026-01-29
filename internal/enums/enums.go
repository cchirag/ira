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
