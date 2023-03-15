package domain

type pduType byte
type parameterCodeType byte

const (
	connectRequestCR               pduType = 0x0e0
	connectConfirmCC               pduType = 0x0d0
	disconnectRequestDR            pduType = 0x080
	disconnectConfirmDC            pduType = 0x0c0
	dataDT                         pduType = 0x0f0
	expeditedDataED                pduType = 0x010
	dataAcknowledgementAK          pduType = 0x060
	expeditedDataAcknowledgementEA pduType = 0x020
	rejectRJ                       pduType = 0x050
	tPDUErrorER                    pduType = 0x070
)
const (
	// SrcTSAP Source Transport Service Access Point
	SrcTSAP  parameterCodeType = 0xc1
	DstTSAP  parameterCodeType = 0xc2
	TPDUSize parameterCodeType = 0xc0
)

type cotpParameter struct {
	paramCode   parameterCodeType // parameter code
	paramLength byte              // parameter length
	data        []byte            // parameter data
}

// CoTP Connection-oriented Transport Protocol
type CoTP struct {
	length     byte
	pduType    pduType // protocol data unit
	destRef1   byte    // destination reference
	destRef2   byte
	sourceRef1 byte // source reference
	sourceRef2 byte
	dataUnit   byte // data unit
	param      []cotpParameter
}

type ICoptUsecase interface {
}
