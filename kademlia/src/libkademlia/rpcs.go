package libkademlia

// Contains definitions mirroring the Kademlia spec. You will need to stick
// strictly to these to be compatible with the reference implementation and
// other groups' code.

import (
	"net"
	"fmt"
)

type KademliaRPC struct {
	kademlia *Kademlia
}

// Host identification.
type Contact struct {
	NodeID ID
	Host   net.IP
	Port   uint16
}


///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
	MsgID  ID
}

type PongMessage struct {
	MsgID  ID
	Sender Contact
}


func (k *KademliaRPC) Ping(ping PingMessage, pong *PongMessage) error {
	// TODO: Finish implementation
	pong.MsgID = CopyID(ping.MsgID)
	//fmt.Println("copy success")
	//var packages Package

	pong.Sender = k.kademlia.SelfContact

	k.kademlia.LocalPing(&ping.Sender)

	//k.kademlia.contactChan <- &ping.Sender

	//k.kademlia.update(&ping.Sender)
	// Specify the sender
	// Update contact, etc
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// STORE
///////////////////////////////////////////////////////////////////////////////
type StoreRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
	Value  []byte
}

type StoreResult struct {
	MsgID ID
	Err   error
}

func (k *KademliaRPC) Store(req StoreRequest, res *StoreResult) error {
	res.MsgID = CopyID(req.MsgID)
	// TODO: Implement.
	//k.kademlia.HashMap[req.Key] = req.Value
	//Or use channal to pass contact

	//k.kademlia.contactChan <- &req.Sender
	if !k.kademlia.NodeID.Equals(req.Sender.NodeID){
			k.kademlia.LocalPing(&req.Sender)
	}

	fmt.Println("in the store function")
	k.kademlia.HashMapChan <- &req
	//k.kademlia.update(&req.Sender)

	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_NODE
///////////////////////////////////////////////////////////////////////////////
type FindNodeRequest struct {
	Sender Contact
	MsgID  ID
	NodeID ID
}

type FindNodeResult struct {
	MsgID ID
	Nodes []Contact
	Err   error
}

func (k *KademliaRPC) FindNode(req FindNodeRequest, res *FindNodeResult) error {
	// TODO: Implement.
	fmt.Println("In RPC findNode")
	contacts := k.kademlia.findCloseNodes(req.NodeID, 20)
	res.MsgID = CopyID(req.MsgID)
	// res.Nodes = make([]Contact, len(contacts))

	res.Nodes = contacts
	if (len(contacts) == 0 || contacts == nil) {
		fmt.Println("No contacts found")
	}
	res.Err = nil

	// res{req.MsgID, contacts, nil}
	return nil
}

///////////////////////////////////////////////////////////////////////////////
// FIND_VALUE
///////////////////////////////////////////////////////////////////////////////
type FindValueRequest struct {
	Sender Contact
	MsgID  ID
	Key    ID
}

// If Value is nil, it should be ignored, and Nodes means the same as in a
// FindNodeResult.
type FindValueResult struct {
	MsgID ID
	Value []byte
	Nodes []Contact
	Err   error
}

func (k *KademliaRPC) FindValue(req FindValueRequest, res *FindValueResult) error {
	// TODO: Implement.
	res.MsgID = CopyID(req.MsgID)
	res.Err = nil
	values, _ := k.kademlia.LocalFindValue(req.Key)
	//values := k.kademlia.HashMap[req.Key]
	if (values == nil || len(values) == 0 ){
		contacts := k.kademlia.findCloseNodes(req.Key, 20)
		res.Value = nil
		res.Nodes = contacts
		return nil
	} else {
	// values, contacts := k.kademlia.findCloseValues(req.key, 20)
		res.Value = values
		res.Nodes  = nil
		return nil
	}
}

// For Project 3

type GetVDORequest struct {
	Sender Contact
	VdoID  ID
	MsgID  ID
}

type GetVDOResult struct {
	MsgID ID
	VDO   VanashingDataObject
}

func (k *KademliaRPC) GetVDO(req GetVDORequest, res *GetVDOResult) error {
	// TODO: Implement.
	return nil
}
