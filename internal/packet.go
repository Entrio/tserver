package internal

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type (
	Packet struct {
		buffer      []byte
		cursor      uint16
		client      *Client
		fragmented  bool
		payloadSize uint16
	}
)

func NewPacket(data []byte) *Packet {
	return &Packet{
		buffer: data,
		cursor: 0,
	}
}

// GetBytes gets the packet as a byte array, ready for sending
func (packet *Packet) GetBytes() []byte {
	pSize := make([]byte, 2)
	binary.LittleEndian.PutUint16(pSize, uint16(len(packet.buffer))) // +2 for packet type

	return append(pSize, packet.buffer...)
}

// GetRemainderBytes retrieves the remainder of unread buffer
func (packet *Packet) GetRemainderBytes() []byte {
	return packet.buffer[packet.cursor:]
}

func (packet *Packet) SetBytes(data []byte) {
	packet.buffer = data
	packet.cursor = 0
}

// Length returns the total packet buffer size
func (packet *Packet) Length() uint16 {
	return uint16(len(packet.buffer))
}

// UnreadLength calculates how many bytes are still unread
func (packet *Packet) UnreadLength() uint16 {
	if packet.cursor >= packet.Length() {
		return 0
	}
	return packet.Length() - packet.cursor
}

// ReadUInt32 reads a 4 byte unsigned integer and move the cursor by 4 positions
func (packet *Packet) ReadUInt32() (val uint32) {
	buf := bytes.NewBuffer(packet.buffer[packet.cursor : packet.cursor+4])
	if err := binary.Read(buf, binary.LittleEndian, &val); err != nil {
		panic(err)
		return
	}

	packet.cursor += 4
	return val
}

// ReadUint16 reads a 2 byte unsigned integer and moving the cursor by 2 positions
func (packet *Packet) ReadUint16() uint16 {
	val := uint16(0)
	val = binary.LittleEndian.Uint16(packet.buffer[packet.cursor : packet.cursor+2])
	packet.cursor += 2
	return val
}

// ReadBytes reads lenToRead bytes from the buffer
func (packet *Packet) ReadBytes(lenToRead uint16) []byte {

	// Check to see if we are not reading past the buffer lenToRead

	if packet.Length() >= (packet.cursor + lenToRead) {
		// we got something
		data := packet.buffer[packet.cursor : packet.cursor+lenToRead]

		// move the cursor
		packet.cursor += lenToRead

		return data
	} else {
		// TODO: Panic or some sort of error
		panic(
			fmt.Errorf(
				"Attempted to read outisde of slice bounds. Packet length: %d, reading %d bytes. Read index from %d - %d\nUnread: %d",
				packet.Length(), lenToRead, packet.cursor, packet.cursor+lenToRead, packet.UnreadLength(),
			),
		)
	}
}

// WriteByte writes a single byte to the buffer
func (packet *Packet) WriteByte(data byte) {
	packet.buffer = append(packet.buffer, data)
}

// WriteUint8 is an alias method for WriteByte
func (packet *Packet) WriteUint8(data uint8) *Packet {
	packet.WriteByte(data)
	return packet
}

// Reset is used when a packet instance need to reused, if it is forced, the cursor is reset and the buffer cleared
func (packet *Packet) Reset(force bool) {
	if force {
		packet.buffer = make([]byte, 0)
		packet.cursor = 0
	} else {
		packet.cursor -= 2
	}
}

func (packet *Packet) Seek(pos uint16) {
	packet.cursor += pos
}

func (packet *Packet) ResetCursor() *Packet {
	packet.cursor = 0
	return packet
}
