// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

/*
Package messages provides encoding/decoding feature of GTPv1 protocol.
*/
package messages

/**
Refer by TS 29.281
		Bits
Octets		8	7	6	5	4	3	2	1
1		Version	PT	(*)	E	S	PN
2		Message Type
3		Length (1st Octet)
4		Length (2nd Octet)
5		Tunnel Endpoint Identifier (1st Octet)
6		Tunnel Endpoint Identifier (2nd Octet)
7		Tunnel Endpoint Identifier (3rd Octet)
8		Tunnel Endpoint Identifier (4th Octet)
9		Sequence Number (1st Octet)1) 4)
10		Sequence Number (2nd Octet)1) 4)
11		N-PDU Number2) 4)
12		Next Extension Header Type3) 4)

Message Type value (Decimal)	Message	Reference	GTP-C	GTP-U	GTP'
1	    Echo Request		X	X	x
2	    Echo Response		X	X	x
3-25	Reserved in 3GPP TS 32.295 [8] and 3GPP TS 29.060 [6]
26	    Error Indication			X
27-30	Reserved in 3GPP TS 29.060 [6]
31	    Supported Extension Headers Notification		X	X
32-253	Reserved in 3GPP TS 29.060 [6]
254	    End Marker			X
255	    G-PDU			X
*/
// Message Type definitions.
const (
	_ uint8 = iota
	MsgTypeEchoRequest
	MsgTypeEchoResponse
	MsgTypeVersionNotSupported
	MsgTypeNodeAliveRequest
	MsgTypeNodeAliveResponse
	MsgTypeRedirectionRequest
	MsgTypeRedirectionResponse
	_
	_
	_
	_
	_
	_
	_
	_
	MsgTypeCreatePDPContextRequest // 16
	MsgTypeCreatePDPContextResponse
	MsgTypeUpdatePDPContextRequest
	MsgTypeUpdatePDPContextResponse
	MsgTypeDeletePDPContextRequest
	MsgTypeDeletePDPContextResponse
	MsgTypeCreateAAPDPContextRequest
	MsgTypeCreateAAPDPContextResponse
	MsgTypeDeleteAAPDPContextRequest
	MsgTypeDeleteAAPDPContextResponse
	MsgTypeErrorIndication
	MsgTypePDUNotificationRequest
	MsgTypePDUNotificationResponse
	MsgTypePDUNotificationRejectRequest
	MsgTypePDUNotificationRejectResponse
	_
	MsgTypeSendRoutingInfoRequest
	MsgTypeSendRoutingInfoResponse
	MsgTypeFailureReportRequest
	MsgTypeFailureReportResponse
	MsgTypeNoteMSPresentRequest
	MsgTypeNoteMSPresentResponse
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	_
	MsgTypeIdentificationRequest // 48
	MsgTypeIdentificationResponse
	MsgTypeSGSNContextRequest
	MsgTypeSGSNContextResponse
	MsgTypeSGSNContextAcknowledge
	MsgTypeDataRecordTransferRequest  uint8 = 240
	MsgTypeDataRecordTransferResponse uint8 = 241
	MsgTypeTPDU                       uint8 = 255
)

// Message is an interface that defines Message messages.
type Message interface {
	MarshalTo([]byte) error
	UnmarshalBinary(b []byte) error
	MarshalLen() int
	Version() int
	MessageType() uint8
	MessageTypeName() string
	TEID() uint32
	SetTEID(uint32)
	Sequence() uint16
	SetSequenceNumber(uint16)

	// deprecated
	SerializeTo([]byte) error
	DecodeFromBytes(b []byte) error
}

// Marshal returns the byte sequence generated from a Message instance.
// Better to use MarshalXxx instead if you know the name of message to be serialized.
func Marshal(g Message) ([]byte, error) {
	b := make([]byte, g.MarshalLen())
	if err := g.MarshalTo(b); err != nil {
		return nil, err
	}

	return b, nil
}

// Parse decodes the given bytes as Message.
func Parse(b []byte) (Message, error) {
	var m Message

	switch b[1] {
	case MsgTypeEchoRequest:
		m = &EchoRequest{}
	case MsgTypeEchoResponse:
		m = &EchoResponse{}
	case MsgTypeCreatePDPContextRequest:
		m = &CreatePDPContextRequest{}
	case MsgTypeCreatePDPContextResponse:
		m = &CreatePDPContextResponse{}
	case MsgTypeUpdatePDPContextRequest:
		m = &UpdatePDPContextRequest{}
	case MsgTypeUpdatePDPContextResponse:
		m = &UpdatePDPContextResponse{}
	case MsgTypeDeletePDPContextRequest:
		m = &DeletePDPContextRequest{}
	case MsgTypeVersionNotSupported:
		m = &VersionNotSupported{}
	case MsgTypeDeletePDPContextResponse:
		m = &DeletePDPContextResponse{}
	/* XXX - Implement!
	case MsgTypeNodeAliveRequest:
		m = &NodeAliveReq{}
	case MsgTypeNodeAliveResponse:
		m = &NodeAliveRes{}
	case MsgTypeRedirectionRequest:
		m = &RedirectionReq{}
	case MsgTypeRedirectionResponse:
		m = &RedirectionRes{}
	case MsgTypeCreateAaPDPContextRequest:
		m = &CreateAaPDPContextReq{}
	case MsgTypeCreateAaPDPContextResponse:
		m = &CreateAaPDPContextRes{}
	case MsgTypeDeleteAaPDPContextRequest:
		m = &DeleteAaPDPContextReq{}
	case MsgTypeDeleteAaPDPContextResponse:
		m = &DeleteAaPDPContextRes{}
	*/
	case MsgTypeErrorIndication:
		m = &ErrorIndication{}
	/* XXX - Implement!
	case MsgTypePduNotificationRequest:
		m = &PduNotificationReq{}
	case MsgTypePduNotificationResponse:
		m = &PduNotificationRes{}
	case MsgTypePduNotificationRejectRequest:
		m = &PduNotificationRejectReq{}
	case MsgTypePduNotificationRejectResponse:
		m = &PduNotificationRejectRes{}
	case MsgTypeSendRoutingInfoRequest:
		m = &SendRoutingInfoReq{}
	case MsgTypeSendRoutingInfoResponse:
		m = &SendRoutingInfoRes{}
	case MsgTypeFailureReportRequest:
		m = &FailureReportReq{}
	case MsgTypeFailureReportResponse:
		m = &FailureReportRes{}
	case MsgTypeNoteMsPresentRequest:
		m = &NoteMsPresentReq{}
	case MsgTypeNoteMsPresentResponse:
		m = &NoteMsPresentRes{}
	case MsgTypeIdentificationRequest:
		m = &IdentificationReq{}
	case MsgTypeIdentificationResponse:
		m = &IdentificationRes{}
	case MsgTypeSgsnContextRequest:
		m = &SgsnContextReq{}
	case MsgTypeSgsnContextResponse:
		m = &SgsnContextRes{}
	case MsgTypeSgsnContextAcknowledge:
		m = &SgsnContextAck{}
	case MsgTypeDataRecordTransferRequest:
		m = &DataRecordTransferReq{}
	case MsgTypeDataRecordTransferResponse:
		m = &DataRecordTransferRes{}
	*/
	case MsgTypeTPDU:
		m = &TPDU{}
	default:
		m = &Generic{}
	}

	if err := m.UnmarshalBinary(b); err != nil {
		return nil, err
	}
	return m, nil
}
