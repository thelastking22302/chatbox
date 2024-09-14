package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	chat "thelastking/gRPC/chatbox/chatpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var grpcLog grpclog.LoggerV2

func init() {
	grpcLog = grpclog.NewLoggerV2(os.Stdout, os.Stdout, os.Stdout)
}

type ConnectServer struct {
	stream chat.Broadcast_CreateStreamServer
	id     string
	name   string
	active bool
	error  chan error
}

type Server struct {
	chat.UnimplementedBroadcastServer
	ConnectServer []*ConnectServer
}

func (s *Server) CreateStream(pconn *chat.Connect, stream chat.Broadcast_CreateStreamServer) error {
	conn := &ConnectServer{
		stream: stream,
		id:     pconn.User.Id,
		name:   pconn.User.Name,
		active: true,
		error:  make(chan error),
	}
	s.ConnectServer = append(s.ConnectServer, conn)
	return <-conn.error
}

func (s *Server) BroadcastMessage(ctx context.Context, req *chat.Message) (*chat.Close, error) {
	var wg sync.WaitGroup
	done := make(chan struct{})
	for _, conn := range s.ConnectServer {
		wg.Add(1)
		go func(conn *ConnectServer) {
			defer wg.Done()
			if conn.active {
				err := conn.stream.Send(req)
				grpcLog.Info("sending message to user :", conn.name)
				if err != nil {
					grpcLog.Errorf("Error with Stream: %v - Error: %v", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(conn)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	<-done
	return &chat.Close{}, nil
}

func main() {
	server := &Server{ConnectServer: []*ConnectServer{}}
	lis, err := net.Listen("tcp", "localhost:50069")
	if err != nil {
		log.Fatal(err)
	}
	defer lis.Close()
	newSer := grpc.NewServer()
	fmt.Println("Starting server...")
	chat.RegisterBroadcastServer(newSer, server)
	if err := newSer.Serve(lis); err != nil {
		fmt.Println("Connect server failed....")
		panic(err)
	}
}
