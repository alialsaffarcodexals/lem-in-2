package utils

const (
	MaxPaths = 100
	MaxAnts  = 100000
)

type Room struct {
	Name  string
	X, Y  int
	Links []*Room
}

type Graph struct {
	Ants  int
	Rooms map[string]*Room
	Start *Room
	End   *Room
}

type LemError struct {
	Msg    string
	Reason string
}

func (e LemError) Error() string {
	return e.Msg + "\nReason: " + e.Reason
}
