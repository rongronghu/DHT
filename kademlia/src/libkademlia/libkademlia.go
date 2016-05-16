package libkademlia

// Contains the core kademlia type. In addition to core state, this type serves
// as a receiver for the RPC methods, which is required by that package.

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"container/list"
	"sort"
	//"time"
)

const (
	alpha = 3
	b     = 8 * IDBytes
	k     = 20
)

// Kademlia type. You can put whatever state you need in this.
type ContactsTable [b]*list.List

type ValueReq struct {
	Key      ID
	//Value    []byte
	Result   chan []byte

}

type  NodeReq struct {
	Key     ID
	Result 	chan []Contact
}

type PingReq struct {
	contact *Contact
	ReturnACK chan int
}

type IterContact struct {
	contact Contact
	Dist    int
}

type ByDist []IterContact

func (d ByDist) Len() int {
	 return len(d) 
}
func (d ByDist) Swap(i, j int) {
	 d[i], d[j] = d[j], d[i] 
}
func (d ByDist) Less(i, j int) bool { 
	return d[i].Dist > d[j].Dist 
}

type Response struct {
	contacts []Contact
	status 	 bool
	id 		 ID
}

type Kademlia struct {
	NodeID      ID
	SelfContact Contact
	//---4.22-----------------------------------------//
	//RoutingTable [b]*list.List
	Contacts        ContactsTable
	HashMap         map[ID][]byte
	contactChan     chan PingReq
	// conRespChan			chan *PingRes
	HashMapChan     chan *StoreRequest
	SearchValueChan	chan *ValueReq
	ReturnValueChan chan []byte
	SearchNodeChan  chan *NodeReq
	ReturnNodeChan 	chan []Contact
	//ReturnACK		chan int
	//For Project 2

}



func InitialTable() (contacts ContactsTable){
	for i := 0; i < b; i++{
		contacts[i] = list.New()
	}
	return
}

func NewKademliaWithId(laddr string, nodeID ID) *Kademlia {
	k := new(Kademlia)
	k.NodeID = nodeID

	// TODO: Initialize other state here as you add functionality.

	//---4.22-----------------------------------------//
	//k.RoutingTable = make([][]Contact, b)
	k.Contacts = InitialTable()
	k.HashMap = make(map[ID][]byte)
	k.contactChan = make(chan PingReq)
	// k.conRespChan = make(chan *PingRes)
	k.HashMapChan = make(chan *StoreRequest)
	k.SearchValueChan = make(chan *ValueReq)
	k.ReturnValueChan = make(chan []byte)
	k.SearchNodeChan = make(chan *NodeReq)
	k.ReturnNodeChan = make(chan []Contact)

	//k.ReturnACK = make(chan int)
	//------------------------------------------------//


	// Set up RPC server
	// NOTE: KademliaRPC is just a wrapper around Kademlia. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&KademliaRPC{k})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}
	s.HandleHTTP(rpc.DefaultRPCPath+port,
		rpc.DefaultDebugPath+port)
	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Run RPC server forever.
	go http.Serve(l, nil)

	// Add self contact
	hostname, port, _ = net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	k.SelfContact = Contact{k.NodeID, host, uint16(port_int)}

	go func (k *Kademlia) {
		for {
			select {
				case newcon1 := <-k.contactChan:
				//case newcon1 := <- k.contactChan:
					fmt.Println("trying to update Contact")
					k.update(newcon1.contact)
					newcon1.ReturnACK <- 1

					// pingres := <- k.conRespChan
					// pingres.Result <- 1
				case newcon2 := <- k.HashMapChan:
					fmt.Println("trying to update Hashmap")
					k.HashMap[newcon2.Key] = newcon2.Value
				case valuereq := <- k.SearchValueChan:
					fmt.Println("trying to search value")
					valuereq.Result <- k.HashMap[valuereq.Key]
				case nodereq := <- k.SearchNodeChan:
					fmt.Println("trying to search nodes")
					nodereq.Result <- k.findCloseNodes(nodereq.Key, 20)

			}
		}

	}(k)
	return k
}

func NewKademlia(laddr string) *Kademlia {
	return NewKademliaWithId(laddr, NewRandomID())
}

type ContactNotFoundError struct {
	id  ID
	msg string
}

func (e *ContactNotFoundError) Error() string {
	return fmt.Sprintf("%x %s", e.id, e.msg)
}


func (k *Kademlia) FindContact(nodeId ID) (*Contact, error) {
	// TODO: Search through contacts, find specified ID
	// Find contact with provided ID
	if nodeId == k.SelfContact.NodeID {
		return &k.SelfContact, nil
	}

	//----------------------4.22---------------------//
	prefixlen := k.NodeID.Xor(nodeId).PrefixLen()
	contacts := k.Contacts[prefixlen]
	// for id, con := range contacts {
	// 	if id == nodeId {
	// 		return &contact, nil
	// 	}
	// }
	for con := contacts.Front(); con != nil ;con = con.Next(){
		// if con.Value.(Contact).NodeID.Equals(nodeId) {

		// 	return con.Value.(Contact), nil
		// }
		temp := con.Value.(*Contact)
		if temp.NodeID.Equals(nodeId){
			return temp, nil
		}
	}
	//-----------------------------------------------//
	return nil, &ContactNotFoundError{nodeId, "Not found"}
}

type CommandFailed struct {
	msg string
}

func (e *CommandFailed) Error() string {
	return fmt.Sprintf("%s", e.msg)
}

func (k *Kademlia) DoPing(host net.IP, port uint16) (*Contact, error) {
	// TODO: Implement

	portstr := strconv.Itoa(int(port))
	client, err := rpc.DialHTTPPath("tcp", host.String() + ":" + portstr, rpc.DefaultRPCPath + portstr)
	if err != nil {
		//fmt.Println("test")
		log.Fatal("dialing:", err)
	}

	if client != nil {
		fmt.Println("Client exists")
	} else {
		fmt.Println("Client does not exist")

	}

	defer  client.Close()

	ping := PingMessage{k.SelfContact, NewRandomID()}
	var pong PongMessage

	err = client.Call("KademliaRPC.Ping", ping, &pong)
	if err != nil {
        log.Fatal("Call: ", err)
        //fmt.Println("call failed")
    }

 	// fmt.Printf("ping msgID: %s\n", ping.MsgID.AsString())
	// fmt.Printf("pong msgID: %s\n\n", pong.MsgID.AsString())

	if pong.MsgID.Equals(ping.MsgID){
		//
		fmt.Println("pong success")
		newone := PingReq{&pong.Sender, make(chan int)}
		k.contactChan <- newone
		ACK := <-newone.ReturnACK
		//k.contactChan <- &pong.Sender
		// ACK := <- k.ReturnACK
		if ACK != 1 {
			fmt.Println("ACK false")
		}
	}


	return &k.SelfContact, err


	// return nil, &CommandFailed{
	// 	"Unable to ping " + fmt.Sprintf("%s:%v", host.String(), port)}
}

func (k *Kademlia) LocalPing(contact *Contact) (int) {
	newACK := PingReq{contact, make(chan int)}
	k.contactChan <- newACK
	ACK := <- newACK.ReturnACK
	return ACK
}

func (k *Kademlia) DoStore(contact *Contact, key ID, value []byte) error {
	// TODO: Implement

	portstr := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String() + ":" + portstr, rpc.DefaultRPCPath + portstr)
	if err != nil {
		//fmt.Println("test")
		log.Fatal("dialing:", err)
	}

	if client != nil {
		fmt.Println("Client exists")
	} else {
		fmt.Println("Client does not exist")
	}
	defer client.Close()
	//sender contact in Request?
	storeRequest := StoreRequest{k.SelfContact, NewRandomID(), key, value}
	var storeResult StoreResult
	err = client.Call("KademliaRPC.Store", storeRequest, &storeResult)
	if err != nil {
		log.Fatal("Call: ", err)
	}

    fmt.Printf("storeRequest msgID: %s\n", storeRequest.MsgID.AsString())
	fmt.Printf("storeResult: %s\n\n", storeResult.MsgID.AsString())

	if storeRequest.MsgID.Equals(storeResult.MsgID){
		//
		return nil
	}

	return &CommandFailed{"Fatal errot"}
}

func (k *Kademlia) DoFindNode(contact *Contact, searchKey ID) ([]Contact, error) {
	// TODO: Implement
	portstr := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String() + ":" + portstr, rpc.DefaultRPCPath + portstr)
	if err != nil {
		//fmt.Println("test")
		log.Fatal("dialing:", err)
	}

	if client != nil {
		fmt.Println("Client exists")
	} else {
		fmt.Println("Client does not exist")
	}

	defer client.Close()

	findNodeRequest := FindNodeRequest{k.SelfContact, NewRandomID(), searchKey}
	var findNodeResult FindNodeResult
	err = client.Call("KademliaRPC.FindNode", findNodeRequest, &findNodeResult)
	if err != nil {
		log.Fatal("Call: ", err)
		return nil, &CommandFailed{"Err to find nodes"}
	} else {
		fmt.Println("suceess find node")
	}
	// if len(findNodeResult.Nodes) != 0 {
	// 	fmt.Println("Nodes found")
	// 	fmt.Println(len(findNodeResult.Nodes))
	// 	// fmt.Println(findNodeResult.MsgID.AsString())
	// 	for _, con := range findNodeResult.Nodes {
	// 		fmt.Println(con.NodeID.AsString())
	// 	}
	// }
	return findNodeResult.Nodes, nil
}

func (k *Kademlia) DoFindValue(contact *Contact,
	searchKey ID) (value []byte, contacts []Contact, err error) {
	// TODO: Implement
	portstr := strconv.Itoa(int(contact.Port))
	client, err := rpc.DialHTTPPath("tcp", contact.Host.String() + ":" + portstr, rpc.DefaultRPCPath + portstr)
	if err != nil {
		//fmt.Println("test")
		log.Fatal("dialing:", err)
	}

	if client != nil {
		fmt.Println("Client exists")
	} else {
		fmt.Println("Client does not exist")
	}

	defer client.Close()

	findValueRequest := FindValueRequest{k.SelfContact, NewRandomID(), searchKey}
	var findValueResult FindValueResult
	err = client.Call("KademliaRPC.FindValue", findValueRequest, &findValueResult)
	if err != nil {
		return nil, nil, &CommandFailed{"Err to find value"}
	}
	return findValueResult.Value, findValueResult.Nodes, nil
}

func (k *Kademlia) LocalFindValue(searchKey ID) (value []byte, err error) {
	// TODO: Implement
		newreq := ValueReq{searchKey, make(chan []byte)}
		k.SearchValueChan <- &newreq
		value = <- newreq.Result
		return value, nil

}


//-------------------------4.23--------------------------//
func (k *Kademlia) update(contact *Contact) {
	prefixlen := k.NodeID.Xor(contact.NodeID).PrefixLen()

	if prefixlen == 160 {
		return
	}

	contacts := k.Contacts[prefixlen]
	var exist bool
	var newone *list.Element
	for con := contacts.Front(); con != nil; con = con.Next() {
		//fmt.Printf("ping msgID: %s\n", contact.Value.(Contact).NodeID.AsString())
		if con.Value.(*Contact).NodeID.Equals(contact.NodeID) {
			//fmt.Printf("aaaaaaaaaa")
			exist = true
			newone = con
			break
		}
	}

	if exist {
		contacts.MoveToFront(newone)
		return
	} else {

		if contacts.Len() < 20 {
			contacts.PushFront(contact)
		} else {
			//ping  the last one to decide

			backitem := contacts.Back()

			port := backitem.Value.(*Contact).Port
			host := backitem.Value.(*Contact).Host

			portstr := strconv.Itoa(int(port))
			client, err := rpc.DialHTTPPath("tcp", host.String() + ":" + portstr, rpc.DefaultRPCPath + portstr)
			if err != nil {
			log.Fatal("dialing:", err)
			}

			if client != nil {
				fmt.Println("Client exists")
				contacts.MoveToFront(backitem)
			} else {
				fmt.Println("Client does not exist")
				contacts.Remove(backitem)
				contacts.PushFront(contact)
			}
		}
	}
	return

}

func (k *Kademlia) findCloseNodes(searchKey ID, count int)(res []Contact){
	prefixLen := k.NodeID.Xor(searchKey).PrefixLen()
	fmt.Println("In findCloseNodes...........")
	fmt.Println(prefixLen)
	contacts := k.Contacts[prefixLen]
	var lenOfRes int
	lenOfRes = 0
	for con := contacts.Front(); con != nil; con = con.Next() {
		if lenOfRes < count{
			fmt.Println("In the for loop")
			tmpNodeID := con.Value.(*Contact).NodeID
			tmpHost := con.Value.(*Contact).Host
			tmpPort := con.Value.(*Contact).Port
			var new_contact Contact
			new_contact.NodeID = tmpNodeID
			new_contact.Host = tmpHost
			new_contact.Port = tmpPort
			res = append(res, new_contact)
			lenOfRes++
		}
	}

	index := prefixLen
	if len(res) < count{
		if index == b{
			index -= 1
			for ; index >= 0 ; {
				contacts := k.Contacts[index]
				for con:= contacts.Front(); con != nil; con = con.Next() {
					tmpNodeID := con.Value.(*Contact).NodeID
					tmpHost := con.Value.(*Contact).Host
					tmpPort := con.Value.(*Contact).Port
					var new_contact Contact
					new_contact.NodeID = tmpNodeID
					new_contact.Host = tmpHost
					new_contact.Port = tmpPort
					res = append(res, new_contact)
					if len(res) == b{
						return res
					}
				}
				index -= 1
			}
			return res
		} else {
			index += 1
			for ; index < b; {
				contacts := k.Contacts[index]
				for con := contacts.Front(); con != nil; con = con.Next(){
					tmpNodeID := con.Value.(*Contact).NodeID
					tmpHost := con.Value.(*Contact).Host
					tmpPort := con.Value.(*Contact).Port
					var new_contact Contact
					new_contact.NodeID = tmpNodeID
					new_contact.Host = tmpHost
					new_contact.Port = tmpPort
					res = append(res, new_contact)
					if len(res) == b{
						return res
					}
				}
				index += 1
			}
			index = prefixLen-1
			for ; index >= 0; {
				contacts := k.Contacts[index]
				for con := contacts.Front(); con != nil; con = con.Next(){
					tmpNodeID := con.Value.(*Contact).NodeID
					tmpHost := con.Value.(*Contact).Host
					tmpPort := con.Value.(*Contact).Port
					var new_contact Contact
					new_contact.NodeID = tmpNodeID
					new_contact.Host = tmpHost
					new_contact.Port = tmpPort
					res = append(res, new_contact)
					if len(res) == b{
						return res
					}
				}
				index -= 1
			}
			return res
		}
	}
	return res
}

// func (k *Kademlia) findCloseNodes(searchKey ID, count int)(res []Contact){
// 	//res = make([20]Contact)
// 	prefixLen := k.NodeID.Xor(searchKey).PrefixLen()
// 	fmt.Println("In findCloseNodes...........")
// 	fmt.Println(prefixLen)
// 	contacts := k.Contacts[prefixLen]
// 	for con := contacts.Front(); con != nil; con = con.Next() {
// 		fmt.Println("In the for loop")
// 		tmpNodeID := con.Value.(*Contact).NodeID
// 		tmpHost := con.Value.(*Contact).Host
// 		tmpPort := con.Value.(*Contact).Port
// 		var new_contact Contact
// 		new_contact.NodeID = tmpNodeID
// 		new_contact.Host = tmpHost
// 		new_contact.Port = tmpPort
// 		res = append(res, new_contact)
// 	}

// 	index := prefixLen
// 	if len(res) < count{
// 		if index == b{
// 			index -= 1
// 			for ; index >= 0 ; {
// 				contacts := k.Contacts[index]
// 				for con:= contacts.Front(); con != nil; con = con.Next() {
// 					tmpNodeID := con.Value.(*Contact).NodeID
// 					tmpHost := con.Value.(*Contact).Host
// 					tmpPort := con.Value.(*Contact).Port
// 					var new_contact Contact
// 					new_contact.NodeID = tmpNodeID
// 					new_contact.Host = tmpHost
// 					new_contact.Port = tmpPort
// 					res = append(res, new_contact)
// 					if len(res) == b{
// 						return res
// 					}
// 				}
// 				index -= 1
// 			}
// 			return res
// 		} else {
// 			index += 1
// 			for ; index < b; {
// 				contacts := k.Contacts[index]
// 				for con := contacts.Front(); con != nil; con = con.Next(){
// 					tmpNodeID := con.Value.(*Contact).NodeID
// 					tmpHost := con.Value.(*Contact).Host
// 					tmpPort := con.Value.(*Contact).Port
// 					var new_contact Contact
// 					new_contact.NodeID = tmpNodeID
// 					new_contact.Host = tmpHost
// 					new_contact.Port = tmpPort
// 					res = append(res, new_contact)
// 					if len(res) == b{
// 						return res
// 					}
// 				}
// 				index += 1
// 			}
// 			index = prefixLen-1
// 			for ; index >= 0; {
// 				contacts := k.Contacts[index]
// 				for con := contacts.Front(); con != nil; con = con.Next(){
// 					tmpNodeID := con.Value.(*Contact).NodeID
// 					tmpHost := con.Value.(*Contact).Host
// 					tmpPort := con.Value.(*Contact).Port
// 					var new_contact Contact
// 					new_contact.NodeID = tmpNodeID
// 					new_contact.Host = tmpHost
// 					new_contact.Port = tmpPort
// 					res = append(res, new_contact)
// 					if len(res) == b{
// 						return res
// 					}
// 				}
// 				index -= 1
// 			}
// 			return res
// 		}
// 	}
// 	return res
// }

// func (k *Kademlia) findCloseValues(searchKey ID, count int)(value []byte, contacts []Contact){
// 	contacts = make([]Contact, 20)
// 	value = make([]byte)
// 	prefixLen := k.NodeID.Xor(searchKey).PrefixLen()
//
// 	contacts := k.Contacts[prefixLen]
// 	for con := contacts.Front(); con != nil; con = con.Next() {
// 		res = append(res, con.Value.(Contact))
// 	}
//
// 	index := prefixLen
// 	if len(res) < count{
// 		if index == b{
// 			index -= 1
// 			for ; index >= 0 ; {
// 				contacts := k.Contacts[index]
// 				for con:= contacts.Front(); con != nil; con = con.Next() {
// 					res = append(res, con.Value.(Contact))
// 					if len(res) == b{
// 						return res
// 					}
// 				}
// 				index -= 1
// 			}
// 			return res
// 		} else {
// 			index += 1
// 			for ; index <= b; {
// 				contacts := k.Contacts[index]
// 				for con := contacts.Front(); con != nil; con = con.Next(){
// 					res = append(res, con.Value.(Contact))
// 					if len(res) == b{
// 						return res
// 					}
// 				}
// 				index += 1
// 			}
// 			index = prefixLen-1
// 			for ; index >= 0; {
// 				contacts := k.Contacts[index]
// 				for con := contacts.Front(); con != nil; con = con.Next(){
// 					res = append(res, con.Value.(Contact))
// 					if len(res) == b{
// 						return res
// 					}
// 				}
// 				index -= 1
// 			}
// 			return res
// 		}
// 	}
// 	return res
// }
//-------------------------------------------------------//
// func (k *Kademlia) FindLocalValue(Key ID) (value []byte, err error) {
// 	newreq := ValueReq{Key, make(chan []byte)}
// 	k.SearchValueChan <- &newreq
// 	value = <- newreq.Result
// 	return value, nil
//
// }

func (k *Kademlia) FindLocalNodes(key ID) (contacts []Contact, err error) {
	newreq := NodeReq{key, make(chan []Contact)}
	k.SearchNodeChan <- &newreq
	contacts = <- newreq.Result
	return contacts, nil
}

// For project 2!
func (k *Kademlia) DoIterativeFindNode(id ID) ([]Contact, error) {
	shortlist   := make([]IterContact, 0)
	visited     := make(map[ID] int)
	active      := make(map[ID] int)
	outNodeChan := make(chan Response) 
	count 		:= 0
	contacts 	:= k.findCloseNodes(id, 3)
	//len not equal to 3
	lim := len(contacts)
	if lim > 3 {
		lim = alpha
	}
	for i := 0; i < lim; i++ {
		shortlist = append(shortlist, IterContact{contacts[i], contacts[i].NodeID.Xor(id).PrefixLen()})
	}

	if len(shortlist) == 0 {
		return nil, &CommandFailed{"No alpha nodes found. Cannot initiate iterative find"}
	}

	alphaNodes := make([]Contact, 0)
	alpha_count := 0

	for count < 20 {
		sort.Sort(ByDist(shortlist))
		alphaNodes = make([]Contact, 0)
		alpha_count = 0

		for i := 0; i < len(shortlist); i++ {
			if (visited[shortlist[i].contact.NodeID] == 0) {
				visited[shortlist[i].contact.NodeID] = 1
				alphaNodes = append(alphaNodes, shortlist[i].contact)
				alpha_count++;
				if alpha_count == 20 - alpha || alpha_count == alpha {
					break
				}
			}
		}

		if alpha_count == 0 {
			break
		}



		for i := 0; i < alpha_count; i++ {
			go k.SubIterFind(alphaNodes[i], id, outNodeChan)
		}

		var backNodes Response

		for i := 0; i < alpha_count; i++ {
			backNodes = <- outNodeChan
			if backNodes.status {
				count++
				active[backNodes.id] = 1
				for _, bnode := range backNodes.contacts {

					if visited[bnode.NodeID] == 0 {
						// fmt.Println("Comapring Dist....................")
						// fmt.Println(shortlist[len(shortlist) - 1].Dist)
						// fmt.Println(bnode.NodeID.AsString())
						// fmt.Println(id.AsString())
						// fmt.Println(bnode.NodeID.Xor(id).PrefixLen())
						if shortlist[len(shortlist) - 1].Dist < bnode.NodeID.Xor(id).PrefixLen() {  //filter for less distance
							shortlist = append(shortlist, IterContact{bnode, bnode.NodeID.Xor(id).PrefixLen()})
							fmt.Println("added")
						}
						sort.Sort(ByDist(shortlist))
					}
				}
			}
		}



	}

 	// return nil, &CommandFailed{"Not implemented"}

 	ret := make([]Contact, 0)
	nres := 0
	for i := 0; i < len(shortlist) && nres < 20; i++ {
		if active[shortlist[i].contact.NodeID] == 1{
			ret = append(ret, shortlist[i].contact)
			nres++
			fmt.Println(shortlist[i].contact.Port)
		}
	}
	return ret, nil
}

func (k *Kademlia) SubIterFind(con Contact, id ID, outNodeChan chan Response) {
	contacts, err :=  k.DoFindNode(&con, id)
	if err == nil {
		outNodeChan <- Response{contacts, true, con.NodeID}
	} else {
		outNodeChan <- Response{contacts, false, con.NodeID}
	}
}

func (k *Kademlia) DoIterativeStore(key ID, value []byte) ([]Contact, error) {
	res, err := k.DoIterativeFindNode(key)
	if err != nil {
		return nil, &CommandFailed{"IterativeStore is failed"}
	} else {
		for _, con := range res {
			k.DoStore(&con, key, value)
		}
		return res, nil
	}
}
func (k *Kademlia) DoIterativeFindValue(key ID) (value []byte, err error) {
	// return nil, &CommandFailed{"Not implemented"}
	var rvalue []byte
	res, err := k.DoIterativeFindNode(key)
	if err != nil {
		return nil, &CommandFailed{"IterativeFindValue is failed"}
	} else {
		for _, con := range res {
			if con.NodeID.Equals(key) {
				rvalue, _, err = k.DoFindValue(&con, key)
				return rvalue, err
			}
		}
	}

	return nil, nil
}

// For project 3!
func (k *Kademlia) Vanish(data []byte, numberKeys byte,
	threshold byte, timeoutSeconds int) (vdo VanashingDataObject) {
	return
}

func (k *Kademlia) Unvanish(searchKey ID) (data []byte) {
	return nil
}

func Dest(host net.IP, port uint16) string {
	return host.String() + ":" + strconv.FormatInt(int64(port), 10)
}
