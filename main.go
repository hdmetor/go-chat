package main

import (
	"log"
	"net"
)

type ChatRoom struct {
}

func CreateChatRoom() *ChatRoom {
	return &ChatRoom{}
}

func (cr *ChatRoom) ListenForMessages() {}

func (cr *ChatRoom) Join(conn net.Conn)     {}


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
		chatroom.Join(conn)

	}

}
