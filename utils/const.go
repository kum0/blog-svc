package utils

type (
	Status  = int32
	Event   = string
	Literal = string
)

const (
	OK       Status = 0
	Error    Status = 1
	SerError Status = 2

	Message          Event = "Message"
	SubscribeMessage Event = "SubscribeMessage"
	AlreadyMessage   Event = "AlreadyMessage"

	// CodeExpiration = 10 * 60

	IdKey Literal = "userId"
)
