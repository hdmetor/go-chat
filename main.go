package main

import (
	"bufio"
	"log"
	"net"
)

type ChatRoom struct {
	// map to keep track of connected users
	users map[string]*ChatUser
	// incoming messages
	incoming chan string
	// users joining
	joins chan *ChatUser
	// users disconnecting
	disconnets chan string
}

func CreateChatRoom() *ChatRoom {
	return &ChatRoom{
		users:      make(map[string]*ChatUser),
		incoming:   make(chan string),
		joins:      make(chan *ChatUser),
		disconnets: make(chan string),
	}
}

func (cr *ChatRoom) ListenForMessages() {
	go func() {
		for {
			select {
			case user := <-cr.joins:
				cr.users[user.username] = user
				cr.Broadcast(" --- " + user.username + " joined")
			}
		}
	}()
}

func (cr *ChatRoom) Join(conn net.Conn) {
	user := CreateChatUser(conn)
	if user.Login(cr) == nil {
		cr.joins <- user
	} else {
		log.Fatal("Could not log in user ", cr)
	}
}

func (cr *ChatRoom) Broadcast(msg string) {
	for _, user := range cr.users {
		user.Send(msg)
	}

}

type ChatUser struct {
	username    string
	connection  net.Conn
	isConnected bool
	sending     chan string
	writer      *bufio.Writer
	reader      *bufio.Reader
}

func CreateChatUser(conn net.Conn) *ChatUser {

	return &ChatUser{

		connection:  conn,
		isConnected: false,
		sending:     make(chan string),
		writer:      bufio.NewWriter(conn),
		reader:      bufio.NewReader(conn),
	}

}

func (cu *ChatUser) Login(chatroom *ChatRoom) error {

	cu.WriteString("Welcome to the chat\n")
	cu.WriteString("Your username is: ")
	username, err := cu.ReadLine()
	cu.username = username
	if err != nil {
		return err
	}

	log.Println(cu.username, " logged in")

	cu.WriteString("You are succesfully signed in as " + cu.username + "\n")
	return nil

}

func (cu *ChatUser) WriteString(msg string) error {
	_, err := cu.writer.WriteString(msg)
	if err != nil {
		return err
	}
	return cu.writer.Flush()
}

func (cu *ChatUser) ReadLine() (string, error) {
	// ReadLine return (line []byte, isPrefix bookl, err error)
	bytes, _, err := cu.reader.ReadLine()
	message := string(bytes)
	return message, err
}

func (cu *ChatUser) Send(msg string) {
	cu.sending <- msg
}

func (cu *ChatUser) WriteOutgoingMessages(room *ChatRoom) {
	go func() {
		for {
			data := <-cu.sending
			data = data + "\n"
			cu.WriteString(data)
		}
	}()
}

func main() {
	log.Println("Starting chat server...")
	chatroom := CreateChatRoom()

	// let's listen to port 6677
	listener, err := net.Listen("tcp", ":6677")
	if err != nil {
		log.Fatal("Error while binding to port ", err)
	}
	defer listener.Close()
	chatroom.ListenForMessages()

	// when accepting a connection, let's print the
	// address of who has connected

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal(err)
		}

		log.Println(conn.RemoteAddr(), " joined!")
		go chatroom.Join(conn)

	}

}
