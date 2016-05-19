# Chat server in golang

As a way to get my feet wet with `go` I decided to implement a chat service.

To try it:

- clone the repo

- in the repo folder do `go run main.go`. The chat server should start

- open a new terminal window and do `nc localhost 6677` and follow the instructions

- repeat the last step (once for each new user)

- ???

- Profit!

# How it works:

There are two main data structure, one for the chat room, and one for a user.

## The chat room

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

The Room contains global information about the chat room:

- `users` is a map that associates each username to its `User` object

- `incoming` is the channel for incoming messages (that are broadcast to the whole room)

- `joins` is a channel where each new `User` object is put when a physical user joins the room

- `disconnects` is the channel that contains the users disconnecting

## The user

    type User struct {
        name       string
        connection net.Conn
        disconnect bool
        sending    chan string
        writer     *bufio.Writer
        reader     *bufio.Reader
    }

The user `struct` contains information about each user:

- `name` contains the name that the user selected when it joined

- `connection` does the actual connection with the socket

- `disconnect` is set to true when the user disconnects

- `sending` is a `chan` with the user outgoing messages

- `writer` / `reader` are the object to write / read to and from the socket

## Special commands:

When typing `/users` as a chat message, the chat will list all the currently active users. Such list is not broadcast to other users.

Note that is not possible for a username to start with '/'.
