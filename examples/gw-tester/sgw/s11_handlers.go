// Copyright 2019-2020 upf authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

package main

import (
	"log"
	"net"
	"time"

	"github.com/pkg/errors"

	v2 "upf/gtp/v2"
	"upf/gtp/v2/ies"
	"upf/gtp/v2/messages"
)

func (s *sgw) handleCreateSessionRequest(s11Conn *v2.Conn, mmeAddr net.Addr, msg messages.Message) error {
	log.Printf("Received %s from %s", msg.MessageTypeName(), mmeAddr)
	if s.mc != nil {
		s.mc.messagesReceived.WithLabelValues(mmeAddr.String(), msg.MessageTypeName()).Inc()
	}

	s11Session := v2.NewSession(mmeAddr, &v2.Subscriber{Location: &v2.Location{}})
	s11Bearer := s11Session.GetDefaultBearer()

	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	csReqFromMME := msg.(*messages.CreateSessionRequest)

	var pgwAddrString string
	if ie := csReqFromMME.PGWS5S8FTEIDC; ie != nil {
		ip, err := ie.IPAddress()
		if err != nil {
			return err
		}
		pgwAddrString = ip + v2.GTPCPort

		teid, err := ie.TEID()
		if err != nil {
			return err
		}
		s11Session.AddTEID(v2.IFTypeS5S8PGWGTPC, teid)
	} else {
		return &v2.RequiredIEMissingError{Type: ies.FullyQualifiedTEID}
	}
	if ie := csReqFromMME.SenderFTEIDC; ie != nil {
		teid, err := ie.TEID()
		if err != nil {
			return err
		}
		s11Session.AddTEID(v2.IFTypeS11MMEGTPC, teid)
	} else {
		return &v2.RequiredIEMissingError{Type: ies.FullyQualifiedTEID}
	}

	raddr, err := net.ResolveUDPAddr("udp", pgwAddrString)
	if err != nil {
		return err
	}

	// keep session information retrieved from the message.
	// XXX - should return error if required IE is missing.
	if ie := csReqFromMME.IMSI; ie != nil {
		imsi, err := ie.IMSI()
		if err != nil {
			return err
		}

		// remove previous session for the same subscriber if exists.
		sess, err := s11Conn.GetSessionByIMSI(imsi)
		if err != nil {
			switch err.(type) {
			case *v2.UnknownIMSIError:
				// whole new session. just ignore.
			default:
				return errors.Wrap(err, "got something unexpected")
			}
		} else {
			s11Conn.RemoveSession(sess)
		}

		s11Session.IMSI = imsi
	} else {
		return &v2.RequiredIEMissingError{Type: ies.IMSI}
	}
	if ie := csReqFromMME.MSISDN; ie != nil {
		s11Session.MSISDN, err = ie.MSISDN()
		if err != nil {
			return err
		}
	} else {
		return &v2.RequiredIEMissingError{Type: ies.MSISDN}
	}
	if ie := csReqFromMME.MEI; ie != nil {
		s11Session.IMEI, err = ie.MobileEquipmentIdentity()
		if err != nil {
			return err
		}
	} else {
		return &v2.RequiredIEMissingError{Type: ies.MobileEquipmentIdentity}
	}
	if ie := csReqFromMME.APN; ie != nil {
		s11Bearer.APN, err = ie.AccessPointName()
		if err != nil {
			return err
		}
	} else {
		return &v2.RequiredIEMissingError{Type: ies.AccessPointName}
	}
	if ie := csReqFromMME.ServingNetwork; ie != nil {
		s11Session.MCC, err = ie.MCC()
		if err != nil {
			return err
		}
		s11Session.MNC, err = ie.MNC()
		if err != nil {
			return err
		}
	} else {
		return &v2.RequiredIEMissingError{Type: ies.ServingNetwork}
	}
	if ie := csReqFromMME.RATType; ie != nil {
		s11Session.RATType, err = ie.RATType()
		if err != nil {
			return err
		}
	} else {
		return &v2.RequiredIEMissingError{Type: ies.RATType}
	}
	s11sgwFTEID := s11Conn.NewSenderFTEID(s.s11IP, "")
	s11sgwTEID := s11sgwFTEID.MustTEID()
	s11Conn.RegisterSession(s11sgwTEID, s11Session)

	s5cFTEID := s.s5cConn.NewSenderFTEID(s.s5cIP, "")
	s5uFTEID := s.s5uConn.NewFTEID(v2.IFTypeS5S8SGWGTPU, s.s5uIP, "").WithInstance(2)

	s5Session, seq, err := s.s5cConn.CreateSession(
		raddr,
		csReqFromMME.IMSI, csReqFromMME.MSISDN, csReqFromMME.MEI, csReqFromMME.ServingNetwork,
		csReqFromMME.RATType, csReqFromMME.IndicationFlags, s5cFTEID, csReqFromMME.PGWS5S8FTEIDC,
		csReqFromMME.APN, csReqFromMME.SelectionMode, csReqFromMME.PDNType, csReqFromMME.PAA,
		csReqFromMME.APNRestriction, csReqFromMME.AMBR, csReqFromMME.ULI,
		ies.NewBearerContext(
			ies.NewEPSBearerID(5),
			s5uFTEID,
			ies.NewBearerQoS(1, 2, 1, 0xff, 0, 0, 0, 0),
		),
		csReqFromMME.MMEFQCSID,
		ies.NewFullyQualifiedCSID(s.s5uIP, 1).WithInstance(1),
	)
	if err != nil {
		return err
	}
	s5Session.AddTEID(s5uFTEID.MustInterfaceType(), s5uFTEID.MustTEID())

	log.Printf("Sent Create Session Request to %s for %s", pgwAddrString, s5Session.IMSI)
	if s.mc != nil {
		s.mc.messagesSent.WithLabelValues(mmeAddr.String(), "Create Session Request").Inc()
	}

	var csRspFromSGW *messages.CreateSessionResponse
	s11mmeTEID, err := s11Session.GetTEID(v2.IFTypeS11MMEGTPC)
	if err != nil {
		s11Conn.RemoveSession(s11Session)
		return err
	}
	message, err := s11Session.WaitMessage(seq, 5*time.Second)
	if err != nil {
		csRspFromSGW = messages.NewCreateSessionResponse(
			s11mmeTEID, 0,
			ies.NewCause(v2.CauseNoResourcesAvailable, 0, 0, 0, nil),
		)

		if err := s11Conn.RespondTo(mmeAddr, csReqFromMME, csRspFromSGW); err != nil {
			s11Conn.RemoveSession(s11Session)
			return err
		}
		log.Printf(
			"Sent %s with failure code: %d, target subscriber: %s",
			csRspFromSGW.MessageTypeName(), v2.CausePGWNotResponding, s11Session.IMSI,
		)
		s11Conn.RemoveSession(s11Session)
		return err
	}

	var csRspFromPGW *messages.CreateSessionResponse
	switch m := message.(type) {
	case *messages.CreateSessionResponse:
		// move forward
		csRspFromPGW = m

		bearer := s11Session.GetDefaultBearer()
		if ie := csRspFromPGW.PAA; ie != nil {
			bearer.SubscriberIP, err = ie.IPAddress()
			if err != nil {
				return err
			}
		}
	default:
		s11Conn.RemoveSession(s11Session)
		return &v2.UnexpectedTypeError{Msg: message}
	}

	// if everything in CreateSessionResponse seems OK, relay it to MME.
	s1usgwFTEID := s.s1uConn.NewFTEID(v2.IFTypeS1USGWGTPU, s.s1uIP, "")
	csRspFromSGW = csRspFromPGW
	csRspFromSGW.SenderFTEIDC = s11sgwFTEID
	csRspFromSGW.SGWFQCSID = ies.NewFullyQualifiedCSID(s.s1uIP, 1).WithInstance(1)
	csRspFromSGW.BearerContextsCreated.Add(s1usgwFTEID)
	csRspFromSGW.BearerContextsCreated.Remove(ies.ChargingID, 0)
	csRspFromSGW.SetTEID(s11mmeTEID)
	csRspFromSGW.SetLength()

	s11Session.AddTEID(s11sgwFTEID.MustInterfaceType(), s11sgwTEID)
	s11Session.AddTEID(s1usgwFTEID.MustInterfaceType(), s1usgwFTEID.MustTEID())

	if err := s11Conn.RespondTo(mmeAddr, csReqFromMME, csRspFromSGW); err != nil {
		s11Conn.RemoveSession(s11Session)
		return err
	}
	if s.mc != nil {
		s.mc.messagesSent.WithLabelValues(mmeAddr.String(), csRspFromSGW.MessageTypeName()).Inc()
	}

	s5cpgwTEID, err := s5Session.GetTEID(v2.IFTypeS5S8PGWGTPC)
	if err != nil {
		s11Conn.RemoveSession(s11Session)
		return err
	}
	s5csgwTEID, err := s5Session.GetTEID(v2.IFTypeS5S8SGWGTPC)
	if err != nil {
		s11Conn.RemoveSession(s11Session)
		return err
	}

	if err := s11Session.Activate(); err != nil {
		s11Conn.RemoveSession(s11Session)
		return err
	}

	log.Printf(
		"Session created with MME and P-GW for Subscriber: %s;\n\tS11 MME:  %s, TEID->: %#x, TEID<-: %#x\n\tS5C P-GW: %s, TEID->: %#x, TEID<-: %#x",
		s5Session.Subscriber.IMSI, mmeAddr, s11mmeTEID, s11sgwTEID, pgwAddrString, s5cpgwTEID, s5csgwTEID,
	)
	return nil
}

func (s *sgw) handleModifyBearerRequest(s11Conn *v2.Conn, mmeAddr net.Addr, msg messages.Message) error {
	log.Printf("Received %s from %s", msg.MessageTypeName(), mmeAddr)
	if s.mc != nil {
		s.mc.messagesReceived.WithLabelValues(mmeAddr.String(), msg.MessageTypeName()).Inc()
	}

	s11Session, err := s11Conn.GetSessionByTEID(msg.TEID(), mmeAddr)
	if err != nil {
		return err
	}
	s5cSession, err := s.s5cConn.GetSessionByIMSI(s11Session.IMSI)
	if err != nil {
		return err
	}
	s1uBearer := s11Session.GetDefaultBearer()
	s5uBearer := s5cSession.GetDefaultBearer()

	var enbIP string
	mbReqFromMME := msg.(*messages.ModifyBearerRequest)
	if brCtxIE := mbReqFromMME.BearerContextsToBeModified; brCtxIE != nil {
		for _, ie := range brCtxIE.ChildIEs {
			switch ie.Type {
			case ies.Indication:
				// do nothing in this implementation.
			case ies.FullyQualifiedTEID:
				if err := s.handleFTEIDU(ie, s11Session, s1uBearer); err != nil {
					return err
				}
				enbIP, err = ie.IPAddress()
				if err != nil {
					return err
				}
			}
		}
	} else {
		return &v2.RequiredIEMissingError{Type: ies.BearerContext}
	}

	s11mmeTEID, err := s11Session.GetTEID(v2.IFTypeS11MMEGTPC)
	if err != nil {
		return err
	}
	s1usgwTEID, err := s11Session.GetTEID(v2.IFTypeS1USGWGTPU)
	if err != nil {
		return err
	}
	s5usgwTEID, err := s5cSession.GetTEID(v2.IFTypeS5S8SGWGTPU)
	if err != nil {
		return err
	}
	pgwIP, _, err := net.SplitHostPort(s5uBearer.RemoteAddress().String())
	if err != nil {
		return err
	}

	if err := s.s1uConn.AddTunnelOverride(
		net.ParseIP(enbIP), net.ParseIP(s1uBearer.SubscriberIP), s1uBearer.OutgoingTEID(), s1usgwTEID,
	); err != nil {
		return err
	}
	if err := s.s5uConn.AddTunnelOverride(
		net.ParseIP(pgwIP), net.ParseIP(s5uBearer.SubscriberIP), s5uBearer.OutgoingTEID(), s5usgwTEID,
	); err != nil {
		return err
	}

	mbRspFromSGW := messages.NewModifyBearerResponse(
		s11mmeTEID, 0,
		ies.NewCause(v2.CauseRequestAccepted, 0, 0, 0, nil),
		ies.NewBearerContext(
			ies.NewCause(v2.CauseRequestAccepted, 0, 0, 0, nil),
			ies.NewEPSBearerID(s1uBearer.EBI),
			ies.NewFullyQualifiedTEID(v2.IFTypeS1USGWGTPU, s1usgwTEID, s.s1uIP, ""),
		),
	)

	if err := s11Conn.RespondTo(mmeAddr, msg, mbRspFromSGW); err != nil {
		return err
	}
	if s.mc != nil {
		s.mc.messagesSent.WithLabelValues(mmeAddr.String(), mbRspFromSGW.MessageTypeName()).Inc()
	}

	log.Printf(
		"Started listening on U-Plane for Subscriber: %s;\n\tS1-U: %s\n\tS5-U: %s",
		s11Session.IMSI, s.s1uConn.LocalAddr(), s.s5uConn.LocalAddr(),
	)
	return nil
}

func (s *sgw) handleDeleteSessionRequest(s11Conn *v2.Conn, mmeAddr net.Addr, msg messages.Message) error {
	log.Printf("Received %s from %s", msg.MessageTypeName(), mmeAddr)
	if s.mc != nil {
		s.mc.messagesReceived.WithLabelValues(mmeAddr.String(), msg.MessageTypeName()).Inc()
	}

	// assert type to refer to the struct field specific to the message.
	// in general, no need to check if it can be type-asserted, as long as the MessageType is
	// specified correctly in AddHandler().
	dsReqFromMME := msg.(*messages.DeleteSessionRequest)

	s11Session, err := s11Conn.GetSessionByTEID(msg.TEID(), mmeAddr)
	if err != nil {
		return err
	}

	s5Session, err := s.s5cConn.GetSessionByIMSI(s11Session.IMSI)
	if err != nil {
		return err
	}

	s5cpgwTEID, err := s5Session.GetTEID(v2.IFTypeS5S8PGWGTPC)
	if err != nil {
		return err
	}

	seq, err := s.s5cConn.DeleteSession(
		s5cpgwTEID, s5Session,
		ies.NewEPSBearerID(s5Session.GetDefaultBearer().EBI),
	)
	if err != nil {
		return err
	}

	var dsRspFromSGW *messages.DeleteSessionResponse
	s11mmeTEID, err := s11Session.GetTEID(v2.IFTypeS11MMEGTPC)
	if err != nil {
		return err
	}

	message, err := s11Session.WaitMessage(seq, 5*time.Second)
	if err != nil {
		dsRspFromSGW = messages.NewDeleteSessionResponse(
			s11mmeTEID, 0,
			ies.NewCause(v2.CausePGWNotResponding, 0, 0, 0, nil),
		)

		if err := s11Conn.RespondTo(mmeAddr, dsReqFromMME, dsRspFromSGW); err != nil {
			return err
		}
		log.Printf(
			"Sent %s with failure code: %d, target subscriber: %s",
			dsRspFromSGW.MessageTypeName(), v2.CausePGWNotResponding, s11Session.IMSI,
		)
		if s.mc != nil {
			s.mc.messagesSent.WithLabelValues(mmeAddr.String(), dsRspFromSGW.MessageTypeName()).Inc()
		}
		return err
	}

	// use the cause as it is.
	switch m := message.(type) {
	case *messages.DeleteSessionResponse:
		// move forward
		dsRspFromSGW = m
	default:
		return &v2.UnexpectedTypeError{Msg: message}
	}

	dsRspFromSGW.SetTEID(s11mmeTEID)
	if err := s11Conn.RespondTo(mmeAddr, msg, dsRspFromSGW); err != nil {
		return err
	}

	log.Printf("Session deleted for Subscriber: %s", s11Session.IMSI)
	if s.mc != nil {
		s.mc.messagesSent.WithLabelValues(mmeAddr.String(), dsRspFromSGW.MessageTypeName()).Inc()
	}

	s11Conn.RemoveSession(s11Session)
	return nil
}

func (s *sgw) handleDeleteBearerResponse(s11Conn *v2.Conn, mmeAddr net.Addr, msg messages.Message) error {
	log.Printf("Received %s from %s", msg.MessageTypeName(), mmeAddr)
	if s.mc != nil {
		s.mc.messagesReceived.WithLabelValues(mmeAddr.String(), msg.MessageTypeName()).Inc()
	}

	s11Session, err := s11Conn.GetSessionByTEID(msg.TEID(), mmeAddr)
	if err != nil {
		return err
	}

	s5Session, err := s.s5cConn.GetSessionByIMSI(s11Session.IMSI)
	if err != nil {
		return err
	}

	if err := v2.PassMessageTo(s5Session, msg, 5*time.Second); err != nil {
		return err
	}

	// remove bearer in handleDeleteBearerRequest instead of doing here,
	// as Delete Bearer Request does not necessarily have EBI.
	return nil
}

func (s *sgw) handleFTEIDU(ie *ies.IE, session *v2.Session, bearer *v2.Bearer) error {
	if ie.Type != ies.FullyQualifiedTEID {
		return &v2.UnexpectedIEError{IEType: ie.Type}
	}

	ip, err := ie.IPAddress()
	if err != nil {
		return err
	}
	addr, err := net.ResolveUDPAddr("udp", ip+v2.GTPUPort)
	if err != nil {
		return err
	}
	bearer.SetRemoteAddress(addr)

	teid, err := ie.TEID()
	if err != nil {
		return err
	}
	bearer.SetOutgoingTEID(teid)

	it, err := ie.InterfaceType()
	if err != nil {
		return err
	}
	session.AddTEID(it, teid)
	return nil
}
