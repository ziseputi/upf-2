package service

import (
	"log"
	"net"
	"upf/gtp/v2"
	"upf/gtp/v2/ies"
	"upf/gtp/v2/messages"
)

type SessionMessage struct {
	MessageType uint8      //
	Teid        uint32     //
	Sequence    uint32     //
	Version     int        //
	PeerIp      string     //
	UeIp        string     //
	Imsi        string     //
	ErrCh       chan error //
}

func (node *Node) CreateSessionRequest(message SessionMessage) error {
	log.Printf("Received %s from %s", message.MessageType, message.PeerIp)

	//TODO 1.parse args

	//TODO 2.check IMSI and remove session

	//TODO 3.check MSISDN

	//TODO 4.check MEI

	//TODO 5.check APN

	//TODO 6.check MNC MCC

	//TODO 7.check RATType

	//TODO 8.check TEID

	//TODO 9.register Session

	//start u
	if err := node.setupUPlane(net.ParseIP(message.PeerIp), net.ParseIP(message.UeIp), message.Teid, message.Teid); err != nil {
		return err
	}

	log.Printf("Session created with UPF for subscriber: %s;\n\tS5C ran: %s, TEID->: %#x, TEID<-: %#x",
		message.Imsi, message.PeerIp, message.Teid, message.Teid,
	)
	return nil
}

func (u *Node) DeleteSessionRequest(c *v2.Conn, sourceAddr net.Addr, msg messages.Message) error {
	log.Printf("Received %s from %s", msg.MessageTypeName(), sourceAddr)

	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	session, err := c.GetSessionByTEID(msg.TEID(), sourceAddr)
	if err != nil {
		dsr := messages.NewDeleteSessionResponse(
			0, 0,
			ies.NewCause(v2.CauseIMSIIMEINotKnown, 0, 0, 0, nil),
		)
		if err := c.RespondTo(sourceAddr, msg, dsr); err != nil {
			return err
		}

		return err
	}

	// respond to S-GW with DeleteSessionResponse.
	teid, err := session.GetTEID(v2.IFTypeS5S8SGWGTPC)
	if err != nil {
		log.Println(err)
		return nil
	}
	dsr := messages.NewDeleteSessionResponse(
		teid, 0,
		ies.NewCause(v2.CauseRequestAccepted, 0, 0, 0, nil),
	)
	if err := c.RespondTo(sourceAddr, msg, dsr); err != nil {
		return err
	}

	log.Printf("Session deleted for Subscriber: %s", session.IMSI)
	c.RemoveSession(session)
	return nil
}
