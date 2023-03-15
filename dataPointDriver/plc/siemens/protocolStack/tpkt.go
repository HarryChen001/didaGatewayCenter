package protocolStack

type Tpkt struct {
	version  byte
	reserved byte
	length1  byte
	length2  byte
}

func (t *Tpkt) Byte() []byte {
	return []byte{
		t.version,
		t.reserved,
		t.length1,
		t.length2,
	}
}
func (t *Tpkt) SetVersion(version byte) *Tpkt {
	t.version = version
	return t
}
func (t *Tpkt) SetLength(length int) *Tpkt {
	t.length1 = byte(length / 256)
	t.length2 = byte(length % 256)
	return t
}
func NewTPKT() *Tpkt {
	t := Tpkt{}
	return &t
}
