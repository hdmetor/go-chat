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
	disconnects chan string
}

func CreateChatRoom() *Room {
	return &Room{
		users:       make(map[string]*User),
		incoming:    make(chan string),
		joins:       make(chan *User),
		disconnects: make(chan string),
	}
}

func (r *Room) ListenForMessages() {
	go func() {
		for {
			select {
			case message := <-r.incoming:
				r.Broadcast(message)
			case user := <-r.joins:
				r.users[user.name] = user
				r.Broadcast(" --- " + user.name + " joined")
			case user := <-r.disconnects:
				// remove the user from the mapping
				if r.users[user] != nil {
					r.users[user].Close()
					delete(r.users, user)
					r.Broadcast("--- " + user + " left the room")
				}
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

func (r *Room) Logout(user string) {
	r.disconnects <- user
}

func (r *Room) Broadcast(msg string) {
	for _, user := range r.users {
		user.Send(msg)
	}

}

type User struct {
	name       string
	connection net.Conn
	disconnect bool
	sending    chan string
	writer     *bufio.Writer
	reader     *bufio.Reader
}

func CreateUser(conn net.Conn) *User {

	return &User{

		connection: conn,
		disconnect: false,
		sending:    make(chan string),
		writer:     bufio.NewWriter(conn),
		reader:     bufio.NewReader(conn),
	}

}

func (u *User) Login(room *Room) error {

	u.WriteString("Welcome to the chat\n")

	ask := true
	for ask {
		u.WriteString("Please enter your name: ")
		name, err := u.ReadLine()
		if err != nil {
			log.Fatal("Trying to login as ", name, err)
			return err
		}
		if _, ok := room.users[name]; ok {
			log.Println("User ", name, " already exists")
			u.WriteString("The name is unavailable, please try again")
		} else {
			u.name = name
			ask = false
		}
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
			if u.disconnect {
				break
			}
			data = data + "\n"
			err := u.WriteString(data)
			if err != nil {
				room.Logout(u.name)
				break
			}
		}
	}()
}

func (u *User) ReadInMessages(room *Room) {
	go func() {
		for {
			message, err := u.ReadLine()
			if u.disconnect {
				break
			}
			// when a user disconnects err is going to be EOF
			if err != nil {
				room.Logout(u.name)
				break
			}

			if message != "" {
				room.incoming <- ("[" + u.name + "]: " + message)
			}

		}
	}()
}

func (u *User) Close() {
	u.disconnect = true
	u.connection.Close()
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
