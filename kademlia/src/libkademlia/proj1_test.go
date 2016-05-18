package libkademlia

import (
	"bytes"
//	"container/heap"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
	"fmt"
)

func StringToIpPort(laddr string) (ip net.IP, port uint16, err error) {
	hostString, portString, err := net.SplitHostPort(laddr)
	if err != nil {
		return
	}
	ipStr, err := net.LookupHost(hostString)
	if err != nil {
		return
	}
	for i := 0; i < len(ipStr); i++ {
		ip = net.ParseIP(ipStr[i])
		if ip.To4() != nil {
			break
		}
	}
	portInt, err := strconv.Atoi(portString)
	port = uint16(portInt)
	return
}

func TestPing(t *testing.T) {
	instance1 := NewKademlia("localhost:7890")
	instance2 := NewKademlia("localhost:7891")
	host2, port2, _ := StringToIpPort("localhost:7891")
	contact2, err := instance2.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("A node cannot find itself's contact info")
	}
	contact2, err = instance2.FindContact(instance1.NodeID)
	if err == nil {
		t.Error("Instance 2 should not be able to find instance " +
			"1 in its buckets before ping instance 1")
	}
	instance1.DoPing(host2, port2)
	contact2, err = instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	wrong_ID := NewRandomID()
	_, err = instance2.FindContact(wrong_ID)
	if err == nil {
		t.Error("Instance 2 should not be able to find a node with the wrong ID")
	}

	contact1, err := instance2.FindContact(instance1.NodeID)
	if err != nil {
		t.Error("Instance 1's contact not found in Instance 2's contact list")
		return
	}
	if contact1.NodeID != instance1.NodeID {
		t.Error("Instance 1 ID incorrectly stored in Instance 2's contact list")
	}
	if contact2.NodeID != instance2.NodeID {
		t.Error("Instance 2 ID incorrectly stored in Instance 1's contact list")
	}
	return
}

func TestStore(t *testing.T) {
	// test Dostore() function and LocalFindValue() function
	instance1 := NewKademlia("localhost:7892")
	instance2 := NewKademlia("localhost:7893")
	host2, port2, _ := StringToIpPort("localhost:7893")
	instance1.DoPing(host2, port2)
	contact2, err := instance1.FindContact(instance2.NodeID)
	if err != nil {
		t.Error("Instance 2's contact not found in Instance 1's contact list")
		return
	}
	key := NewRandomID()
	value := []byte("Hello World")
	err = instance1.DoStore(contact2, key, value)
	if err != nil {
		t.Error("Can not store this value")
	}
	storedValue, err := instance2.LocalFindValue(key)
	if err != nil {
		t.Error("Stored value not found!")
	}
	if !bytes.Equal(storedValue, value) {
		t.Error("Stored value did not match found value")
	}
	return
}

func TestFindNode(t *testing.T) {
	// tree structure;
	// A->B->tree
	/*
	         C
	      /
	  A-B -- D
	      \
	         E
	*/
	instance1 := NewKademlia("localhost:7894")
	instance2 := NewKademlia("localhost:7895")
	host2, port2, _ := StringToIpPort("localhost:7895")
	instance1.DoPing(host2, port2)
	//contact2, err := instance1.FindContact(instance2.NodeID)
	// if err != nil {
	// 	t.Error("Instance 2's contact not found in Instance 1's contact list")
	// 	return
	// }
	tree_node := make([]*Kademlia, 3)
	for i := 0; i < 3; i++ {
		address := "localhost:" + strconv.Itoa(7896+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)

		// newone := PingReq{&(tree_node[i].SelfContact), make(chan int)}
		// instance1.contactChan <- &newone
		// ACK := <-newone.ReturnACK
		// if ACK != 1 {
		// 	fmt.Println("failed")
		// }
		//instance1.DoPing(host_number, port_number)
	}
	key := NewRandomID()
	fmt.Println("findzhihou-----jiba")



	contacts, err := instance1.DoFindNode(&(instance2.SelfContact), key)


	
	if contacts == nil || len(contacts) == 0 {
		t.Error("No contacts were found")
	}

	fmt.Println("hurongrong =======")
	fmt.Println("Contact NodeID: " + instance2.NodeID.AsString())
	for _, con := range contacts {
		fmt.Println("Node ID: " + con.NodeID.AsString())
	}

	fmt.Println("tree NodeID =========")
	for _, con := range tree_node {
		fmt.Println("Node ID: " + con.NodeID.AsString())
	}

	fmt.Println("Instance1 Node ID: " + instance1.NodeID.AsString())
	fmt.Println("Instance2 Node ID: " + instance2.NodeID.AsString())


	_, err = instance1.FindContact(instance2.NodeID)
	if err != nil {
			t.Error("tree node 1 returned not found in Instance 1's contact list")
			return
	}


	// _, err = instance1.FindContact(tree_node[9].NodeID)
	// if err != nil {
	// 		t.Error("tree node 0 returned not found in Instance 1's contact list")
	// 		return
	// }



	for i := 0; i < 3; i++ {
		_, err := instance1.FindContact(tree_node[i].NodeID)
		if err != nil {
			fmt.Println(i)
			t.Error("Instance returned not found in Instance 1's contact list" + tree_node[i].NodeID.AsString() + " index: ")
			return
		}
		// if returnedContact.NodeID != tree_node[i].NodeID {
		// 	t.Error("Returned ID incorrectly stored in Instance 1's contact list")
		// }
	}
	return
}

func Connect(t *testing.T, list []*Kademlia, kNum int) {
	count := 0
	for i := 0; i < kNum; i++ {
		for j := 0; j < kNum; j++ {
			if j != i {
				list[i].DoPing(list[j].SelfContact.Host, list[j].SelfContact.Port)
			}
			count++
			t.Log(list[i].NodeID.AsString())
		}
	}
}








func TestIterativeFindNode(t *testing.T) {
	//Find the node when it's in local list
	kNum := 40
	targetIdx := kNum - 10
	instance2 := NewKademlia("localhost:8305")
	host2, port2, _ := StringToIpPort("localhost:8305")
	//	instance2.DoPing(host2, port2)
	tree_node := make([]*Kademlia, kNum)
	//t.Log("Before loop")
	for i := 0; i < kNum; i++ {
		address := "localhost:" + strconv.Itoa(8306+i)
		tree_node[i] = NewKademlia(address)
		tree_node[i].DoPing(host2, port2)
		t.Log("ID:" + tree_node[i].SelfContact.NodeID.AsString())
	}
	for i := 0; i < kNum; i++ {
		if i != targetIdx {
			tree_node[targetIdx].DoPing(tree_node[i].SelfContact.Host, tree_node[i].SelfContact.Port)
		}
	}
	SearchKey := tree_node[targetIdx].SelfContact.NodeID

	//time.Sleep(100 * time.Millisecond)

	res, err := tree_node[2].DoIterativeFindNode(SearchKey)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("SearchKey:" + SearchKey.AsString())
	if res == nil || len(res) == 0 {
		t.Error("No contacts were found")
	}
	find := false
	fmt.Print("# of results:  ")
	fmt.Println(len(res))
	for _, value := range res {
		t.Log(value.NodeID.AsString())
		if value.NodeID.Equals(SearchKey) {
			find = true
		}
	}
	if !find {
		t.Log("Instance2:" + instance2.NodeID.AsString())
		t.Error("Find wrong id")
	}
	return
}

func TestIterativeFindValue(t *testing.T) {
	// tree structure;
	// A->B->tree->tree2
	/*
		                F
			  /
		          C --G
		         /    \
		       /        H
		   A-B -- D
		       \
		          E
	*/

	instance1 := NewKademlia("localhost:7406")
	instance2 := NewKademlia("localhost:7407")
	host2, port2, _ := StringToIpPort("localhost:7407")
	instance1.DoPing(host2, port2)

	//Build the  A->B->Tree structure
	tree_node := make([]*Kademlia, 20)
	for i := 0; i < 20; i++ {
		address := "localhost:" + strconv.Itoa(7408+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}
	//Build the A->B->Tree->Tree2 structure
	tree_node2 := make([]*Kademlia, 20)
	for j := 20; j < 40; j++ {
		address := "localhost:" + strconv.Itoa(7408+j)
		tree_node2[j-20] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		for i := 0; i < 20; i++ {
			tree_node[i].DoPing(host_number, port_number)
		}
	}

	//Store value into nodes
	value := []byte("Hello world")
	key := NewRandomID()
	contacts, err := instance1.DoIterativeStore(key, value)
	if err != nil || len(contacts) != 20 {
		t.Error("Error doing DoIterativeStore")
	}

	//After Store, check out the correctness of DoIterativeFindValue
	result, err := instance1.DoIterativeFindValue(key)
	if err != nil || result == nil {
		t.Error("Error doing DoIterativeFindValue")
	}

	//Check the correctness of the value we find
	res := string(result[:])
	fmt.Println(res)
	//t.Error("Finish")
}

func TestIterativeFindNode1(t *testing.T) {
tree_node := make([]*Kademlia, 30)
address := make([]string, 30)
for i := 0; i < 30; i++ {
	address[i] = "localhost:" + strconv.Itoa(7696+i)
	tree_node[i] = NewKademlia(address[i])
}

//30 nodes ping each other
for i := 0; i < 30 ; i++ {
	for j := 0; j < 30; j++ {
		host_number, port_number, _ := StringToIpPort(address[j])
		tree_node[i].DoPing(host_number, port_number)
	}
}

//find node[19], start from node 0
contacts, _ := tree_node[0].DoIterativeFindNode(tree_node[19].SelfContact.NodeID)
count := 0
//check the result
for i := 0; i < len(contacts); i++ {
	if(contacts[i].NodeID.Equals(tree_node[19].SelfContact.NodeID)) {
		count ++
	}
	fmt.Print(contacts[i].NodeID)
}
if(count != 1) {
	t.Error("the result is not true")
}
}




func TestIterativeStore(t *testing.T) {
	// tree structure;
	// A->B->tree->tree2
	/*
	          C
	      /
	   A-B — D
	       \
	          E
	*/
	fmt.Println("test")
	instance1 := NewKademlia("localhost:7506")
	instance2 := NewKademlia("localhost:7507")
	host2, port2, _ := StringToIpPort("localhost:7507")
	instance1.DoPing(host2, port2)

	//Build the  A->B->Tree structure
	tree_node := make([]*Kademlia, 20)
	for i := 0; i < 20; i++ {
		address := "localhost:" + strconv.Itoa(7508+i)
		tree_node[i] = NewKademlia(address)
		host_number, port_number, _ := StringToIpPort(address)
		instance2.DoPing(host_number, port_number)
	}
	//implement DoIterativeStore, and get the the result
	value := []byte("Hello world")
	key := NewRandomID()
	contacts, err := instance1.DoIterativeStore(key, value)
	//the number of contacts store the value should be 20
	if err != nil || len(contacts) != 20 {
		t.Error("Error doing DoIterativeStore")
	}
	//Check all the 22 nodes,
	//find out the number of nodes that contains the value
	count := 0
	// check tree_nodes[0~19]
	for i := 0; i < 20; i++ {
		result, err := tree_node[i].LocalFindValue(key)
		if result != nil && err == nil {
			count++
		}
	}
	//check instance2
	result, err := instance2.LocalFindValue(key)
	if result != nil && err == nil {
		count++
	}
	//check instance1
	result, err = instance1.LocalFindValue(key)
	if result != nil && err == nil {
		count++
	}
	//Within all 22 nodes
	//the number of nodes that store the value should be 20
	if count != 20 {
		t.Error("DoIterativeStore Failed")
	}
}



func TestIterativeFindNode2(t *testing.T) {
	//Find the node when it's in local list
	kNum := 20
	targetIdx := kNum - 10
	instance2 := NewKademlia("localhost:7305")
	host2, port2, _ := StringToIpPort("localhost:7305")
	//	instance2.DoPing(host2, port2)
	tree_node := make([]*Kademlia, kNum)
	//t.Log("Before loop")
	for i := 0; i < kNum; i++ {
		address := "localhost:" + strconv.Itoa(7306+i)
		tree_node[i] = NewKademlia(address)
		tree_node[i].DoPing(host2, port2)
		t.Log("ID:" + tree_node[i].SelfContact.NodeID.AsString())
	}
	for i := 0; i < kNum; i++ {
		if i != targetIdx {
			tree_node[targetIdx].DoPing(tree_node[i].SelfContact.Host, tree_node[i].SelfContact.Port)
		}
	}
	SearchKey := tree_node[targetIdx].SelfContact.NodeID

	time.Sleep(100 * time.Millisecond)

	res, err := tree_node[2].DoIterativeFindNode(SearchKey)
	if err != nil {
		t.Error(err.Error())
	}
	t.Log("SearchKey:" + SearchKey.AsString())
	if res == nil || len(res) == 0 {
		t.Error("No contacts were found")
	}
	find := false
	fmt.Print("# of results:  ")
	fmt.Println(len(res))
	for _, value := range res {
		t.Log(value.NodeID.AsString())
		if value.NodeID.Equals(SearchKey) {
			find = true
		}
	}
	if !find {
		t.Log("Instance2:" + instance2.NodeID.AsString())
		t.Error("Find wrong id")
	}
	return
}

func TestIterativeFindNode3(t *testing.T) {
	//find the node through a ajacent node
	tree_node := make([]*Kademlia, 20)
	address := make([]string, 20)
	for i := 0; i < 20; i++ {
		address[i] = "localhost:" + strconv.Itoa(11896+i)
		tree_node[i] = NewKademlia(address[i])
	}
	for i := 0; i < 20 ; i++ {
		//for j := 0; j < 20; j++ {
			host_number, port_number, _ := StringToIpPort(address[i])
			tree_node[1].DoPing(host_number, port_number)
		//}
	}

	contacts, _ := tree_node[0].DoIterativeFindNode(tree_node[19].SelfContact.NodeID)
	find := false
	for i := 0; i < len(contacts); i++ {
		if(contacts[i].NodeID.Equals(tree_node[19].SelfContact.NodeID)) {
			find = true
		}
	//fmt.Println(contacts[i].NodeID.AsString())
	}

	// res := tree_node[0].findCloseNodes(tree_node[19].SelfContact.NodeID, 3)
	// fmt.Println(tree_node[19].SelfContact.NodeID.AsString())

	if !find {
		t.Log("Finding Node Failed:" + tree_node[19].SelfContact.NodeID.AsString())
		t.Error("Find wrong id")
	}
	return
}

func TestIterativeStrore2(t *testing.T) {

	tree_node := make([]*Kademlia, 20)
	address := make([]string, 20)
	for i := 0; i < 20; i++ {
		address[i] = "localhost:" + strconv.Itoa(12196+i)
		tree_node[i] = NewKademlia(address[i])
	}

	msg := make([]byte, 20)
	for i := 0; i < 20; i++ {
		msg[i] = uint8(rand.Intn(256))
	}

	//Case 1, no node to store
	//fmt.Println(msg)
	_, err1 := tree_node[0].DoIterativeStore(tree_node[19].SelfContact.NodeID, msg)
	if err1 == nil {
		t.Error("IterativeStore Failed. Here shouldnt store anything")
	}

	//case 2 Correctly store the node
	for i := 2; i < 20 ; i++ {
			host_number, port_number, _ := StringToIpPort(address[i])
			tree_node[1].DoPing(host_number, port_number)
	}

	host_1, port_1, _ := StringToIpPort(address[1])
	tree_node[0].DoPing(host_1, port_1)

	_, err := tree_node[0].DoIterativeStore(tree_node[19].SelfContact.NodeID, msg)

	if err != nil {
		t.Error("IterativeStore Failed. Return error")
	}

	value, err2 := tree_node[19].LocalFindValue(tree_node[19].SelfContact.NodeID)
	if err2 != nil || !bytes.Equal(msg, value) {
		t.Error("IterativeStore Failed. Wrong Value")
	}

	return

}

func TestIterativeFindValue2(t *testing.T) {
	tree_node := make([]*Kademlia, 20)
	address := make([]string, 20)
	for i := 0; i < 20; i++ {
		address[i] = "localhost:" + strconv.Itoa(13196+i)
		tree_node[i] = NewKademlia(address[i])
	}

	msg := make([]byte, 20)
	for i := 0; i < 20; i++ {
		msg[i] = uint8(rand.Intn(256))
	}

	//Case 1, no node to store

	for i := 2; i < 20 ; i++ {
			host_number, port_number, _ := StringToIpPort(address[i])
			tree_node[1].DoPing(host_number, port_number)
	}

	host_1, port_1, _ := StringToIpPort(address[1])
	tree_node[0].DoPing(host_1, port_1)

	_, err1 := tree_node[0].DoIterativeFindValue(tree_node[19].SelfContact.NodeID)
	if err1 == nil {
		t.Error("IterativeFindValue Failed. Shouldnt find any value")
	}

	//case 2, value Found
	tree_node[0].DoIterativeStore(tree_node[19].SelfContact.NodeID, msg)
	value2, err2 := tree_node[0].DoIterativeFindValue(tree_node[19].SelfContact.NodeID)
	//value2, err2 := tree_node[19].LocalFindValue(tree_node[19].SelfContact.NodeID)
	if err2 != nil || !bytes.Equal(value2, msg) {
		fmt.Println(value2)
		fmt.Println(msg)
		t.Error("IterativeFindValue. Returned wrong value ")
	}

}









