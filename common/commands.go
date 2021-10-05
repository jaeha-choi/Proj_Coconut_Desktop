package common

type Command string

// Commands and responses. Perhaps using string isn't necessary.
const (
	// Commands

	Quit          Command = "QUIT"
	RequestRelay  Command = "RELY"
	EndRelay      Command = "ERLY"
	GetAddCode    Command = "GADC"
	RemoveAddCode Command = "RADC"
	GetPubKey     Command = "GPUB"
	RequestPubKey Command = "RPUB"
	GetPTPip      Command = "GPTP"
	RequestPTPip  Command = "RPTP"
	GetPTPInit    Command = "PTPI"
)

func (c Command) String() string {
	return string(c)
}
