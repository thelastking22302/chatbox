package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	chat "thelastking/gRPC/chatbox/chatpb"
	"thelastking/gRPC/chatbox/client/login"

	tui "github.com/marcusolsson/tui-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var client chat.BroadcastClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *chat.User, ui tui.UI, newMessage *tui.Box) error {
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &chat.Connect{
		User:   user,
		Active: true,
	})

	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	wait.Add(1)
	go func(str chat.Broadcast_CreateStreamClient) {
		defer wait.Done()

		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("error reading message: %v", err)
				break
			}
			ui.Update(func() {
				usernameText := tui.NewLabel(msg.Name.Name)
				if user.Id == msg.Name.Id {
					newMessage.Append(tui.NewHBox(
						tui.NewPadder(1, 0, tui.NewLabel("You")),
						tui.NewLabel(" : "),
						tui.NewLabel(msg.Content),
						tui.NewSpacer(),
					))
				} else {
					newMessage.Append(tui.NewHBox(
						tui.NewPadder(1, 0, usernameText),
						tui.NewLabel(" : "),
						tui.NewLabel(msg.Content),
						tui.NewSpacer(),
					))
				}
			})
		}
	}(stream)

	return streamerror
}

func main() {
	username := strings.Title(login.GetUserName())

	newMessage := tui.NewVBox()

	chatBoxScroll := tui.NewScrollArea(newMessage)
	chatBoxScroll.SetAutoscrollToBottom(true)

	chatbox := tui.NewVBox(chatBoxScroll)
	chatbox.SetTitle(username)
	chatbox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	msgInputBox := tui.NewVBox(input)
	msgInputBox.SetTitle("Enter Message")
	msgInputBox.SetBorder(true)
	msgInputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	root := tui.NewVBox(chatbox, msgInputBox)
	root.SetSizePolicy(tui.Expanding, tui.Maximum)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() {
		ui.Quit()
		os.Exit(1)
	})

	timestamp := time.Now()
	done := make(chan struct{})

	id := sha256.Sum256([]byte(timestamp.String() + username))

	conn, err := grpc.Dial("localhost:50069", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn't connect to service: %v", err)
	}

	client = chat.NewBroadcastClient(conn)
	user := &chat.User{
		Id:   hex.EncodeToString(id[:]),
		Name: username,
	}

	connect(user, ui, newMessage)

	wait.Add(1)
	go func() {
		defer wait.Done()

		input.OnSubmit(func(e *tui.Entry) {
			m := e.Text()
			if m != "" {
				input.SetText("")
				msg := &chat.Message{
					Name:      user,
					Content:   m,
					Timestamp: timestamp.String(),
				}
				_, err := client.BroadcastMessage(context.Background(), msg)
				if err != nil {
					fmt.Printf("Error Sending Message: %v", err)
				}
			}
		})

	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}

	<-done
}
