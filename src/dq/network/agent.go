package network

type Agent interface {
	Run()
	OnClose()
	GetConnectId() int
	GetModeType() string
}
