package protocolStack

type pduType byte
type parameterCodeType byte

const (
	ConnectRequestCR               pduType = 0x0e0
	connectConfirmCC               pduType = 0x0d0
	disconnectRequestDR            pduType = 0x080
	disconnectConfirmDC            pduType = 0x0c0
	DataDT                         pduType = 0x0f0
	expeditedDataED                pduType = 0x010
	dataAcknowledgementAK          pduType = 0x060
	expeditedDataAcknowledgementEA pduType = 0x020
	rejectRJ                       pduType = 0x050
	tPDUErrorER                    pduType = 0x070
)
const (
	// Source Transport Service Access Point
	SrcTSAP  parameterCodeType = 0xc1
	DstTSAP  parameterCodeType = 0xc2
	TPDUSize parameterCodeType = 0xc0
)

type CotpParameter struct {
	ParamCode   parameterCodeType // parameter code
	ParamLength byte              // parameter length
	Data        []byte            // parameter data
}

// CoTP Connection-oriented Transport Protocol
type CoTP struct {
	length     byte
	pduType    pduType // protocol Data unit
	destRef1   byte    // destination reference
	destRef2   byte
	sourceRef1 byte // source reference
	sourceRef2 byte
	dataUnit   byte // Data unit
	param      []CotpParameter
}

func (c *CoTP) SetPduType(pduType2 pduType) *CoTP {
	c.pduType = pduType2
	switch pduType2 {
	case ConnectRequestCR:
		c.length = 17
		c.dataUnit = 0
		c.destRef1 = 0x00
		c.destRef2 = 0x00
		c.sourceRef1 = 0x00
		c.sourceRef2 = 0x01
	case DataDT:
		c.length = 2
		c.dataUnit = 0x80
	}
	return c
}
func (c *CoTP) SetParameter(m []CotpParameter) *CoTP {
	c.param = m
	return c
}
func (c *CoTP) Length() int {
	switch c.pduType {
	case ConnectRequestCR:
		return 17
	case DataDT:
		return 2
	}
	return 0
}
func (c *CoTP) Byte() (r []byte) {
	switch c.pduType {
	case ConnectRequestCR:
		r = append(r, c.length, byte(c.pduType), c.destRef1, c.destRef2, c.sourceRef1, c.sourceRef2, c.dataUnit)
		for _, singlePM := range c.param {
			r = append(r, singlePM.Byte()...)
		}
	case DataDT:
		r = append(r, c.length, byte(c.pduType), c.dataUnit)
	}
	return
}
func (c *CotpParameter) Byte() []byte {
	var result []byte
	result = append(result, byte(c.ParamCode), c.ParamLength)
	result = append(result, c.Data...)
	return result
}
func NewCoTPUsecase() *CoTP {
	c := CoTP{}
	return &c
}
