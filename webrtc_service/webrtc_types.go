package main

import (
	"context"
	"sync"

	"github.com/pion/webrtc/v4"
)

type Room struct {
	id      string
	peers   map[string]*Peer
	peersMU sync.RWMutex

	trackForwarders  map[string]*TrackForwarder
	trackForwarderMU sync.RWMutex
}

type TrackForwarder struct {
	source        *webrtc.TrackRemote
	sourcePeerID  string
	localTracks   map[string]*webrtc.TrackLocalStaticRTP
	localTracksMU sync.RWMutex
	cancel        context.CancelFunc
}

type RTCMessage struct {
	Type      string                     `json:"type"`
	Offer     *webrtc.SessionDescription `json:"offer"`
	Answer    *webrtc.SessionDescription `json:"answer"`
	Candidate *webrtc.ICECandidateInit   `json:"candidate"`
	PeerID    string                     `json:"peer_id,omitempty"`
}
