package nex

import "fmt"


type Error struct {
	ResultCode uint32
	Message    string
	Packet     PacketInterface
}


func (e Error) Error() string {
	resultCode := e.ResultCode

	if int(resultCode)&errorMask != 0 {

		resultCode = resultCode & ^uint32(errorMask)
	}

	return fmt.Sprintf("[%s] %s", ResultCodeToName(resultCode), e.Message)
}


func NewError(resultCode uint32, message string) *Error {
	if int(resultCode)&errorMask == 0 {

		resultCode = uint32(int(resultCode) | errorMask)
	}

	return &Error{
		ResultCode: resultCode,
		Message:    message,
	}
}
