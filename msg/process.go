package msg

import (
	"encoding/binary"
	"errors"
	"io"
)

var MaxMsgLength int64 = 1024

func readMsg(r io.Reader) (t byte, b []byte, err error) {
	var (
		n  int64
		b1 [1]byte
	)

	_, err = r.Read(b1[:])
	if err != nil {
		return
	}

	if t = b1[0]; int(t) >= len(typeMsgs) {
		err = errors.New("message type error")
		return
	}

	err = binary.Read(r, binary.BigEndian, &n)
	if err != nil {
		return
	}

	if n > MaxMsgLength {
		err = errors.New("message length exceed the limit")
		return
	}

	b = make([]byte, n)
	rn, err := io.ReadFull(r, b)
	if err != nil {
		return
	}

	if int64(rn) != n {
		err = errors.New("message format error")
	}
	return
}

func Read(r io.Reader) (msg Message, err error) {
	t, b, err := readMsg(r)
	if err != nil {
		return
	}

	msg, err = UnPack(t, b)
	return
}

func ReadInto(r io.Reader, msg Message) (err error) {
	_, b, err := readMsg(r)
	if err != nil {
		return
	}

	err = UnPackInto(b, msg)
	return
}

func Write(w io.Writer, msg interface{}) (err error) {
	b, err := Pack(msg)
	if err != nil {
		return
	}

	_, err = w.Write(b)
	return
}
