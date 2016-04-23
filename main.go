package main

import (
	"bufio"
	"log"
	"net"
)

type Room struct {
	// map to keep track of connected users
	users map[string]*User
	// incoming messages
	incoming chan string
	// users joining
	joins chan *User
	// users disconnecting
	disconnets chan string
}

func CreateChatRoom() *Room {
	return &Room{
		users:      make(map[string]*User),
		incoming:   make(chan string),
		joins:      make(chan *User),
		disconnets: make(chan string),
	}
}

func (r *Room) ListenForMessages() {
	go func() {
		for {
			select {
				case message := <- r.incoming:
				r.Broadcast(message)
			case user := <-r.joins:
				r.users[user.name] = user
				r.Broadcast(" --- " + user.name + " joined")
			}
		}
	}()
}

func (r *Room) Join(conn net.Conn) {
	user := CreateUser(conn)
	if user.Login(r) == nil {
		r.joins <- user
	} else {
		log.Fatal("Could not log in user ", r)

	}
}

func (r *Room) Broadcast(msg string) {
	for _, user := range r.users {
		user.Send(msg)
	}

}

type User struct {
	name        string
	connection  net.Conn
	isConnected bool
	sending     chan string
	writer      *bufio.Writer
	reader      *bufio.Reader
}

func CreateUser(conn net.Conn) *User {

	return &User{

		connection:  conn,
		isConnected: false,
		sending:     make(chan string),
		writer:      bufio.NewWriter(conn),
		reader:      bufio.NewReader(conn),
	}

}

func (u *User) Login(room *Room) error {

	u.WriteString("Welcome to the chat\n")
	u.WriteString("Your name is: ")
	name, err := u.ReadLine()
	u.name = name
	if err != nil {
		return err
	}

	log.Println(u.name, " logged in")

	u.WriteString("You are succesfully signed in as " + u.name + "\n")
	u.WriteOutgoingMessages(room)
	u.ReadInMessages(room)

	return nil

}

func (u *User) WriteString(msg string) error {
	_, err := u.writer.WriteString(msg)
	if err != nil {
		return err
	}
	return u.writer.Flush()
}

func (u *User) ReadLine() (string, error) {
	// ReadLine return (line []byte, isPrefix bookl, err error)
	bytes, _, err := u.reader.ReadLine()
	message := string(bytes)
	return message, err
}

func (u *User) Send(msg string) {
	u.sending <- msg
}

func (u *User) WriteOutgoingMessages(room *Room) {
	go func() {
		for {
			data := <-u.sending
			data = data + "\n"
			u.WriteString(data)
		}
	}()
}

func (u *User) ReadInMessages(room *Room) {
	go func() {
		for {
			message, err := u.ReadLine()
			if err != nil {
				log.Fatal(err)
			}

			if message != "" {
				room.incoming <- ("[" + u.name + "]: " + message)
			}

		}
	}()
}

func main() {
	log.Println("Starting chat server...")
	room := CreateChatRoom()

	// let's listen to port 6677
	listener, err := net.Listen("tcp", ":6677")
	if err != nil {
		log.Fatal("Error while binding to port ", err)
	}
	defer listener.Close()
	room.ListenForMessages()

	// when accepting a connection, let's print the
	// address of who has connected

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Fatal(err)
		}

		log.Println(conn.RemoteAddr(), " joined!")
		go room.Join(conn)

	}

}
