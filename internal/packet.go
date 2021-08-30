package internal

type (
	Packet struct {
		ContentLength uint8
		Content []byte
	}
)
