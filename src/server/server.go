package main

import (
	"io"
	"log"
	"net"
	"os"

	pb "github.com/pahanini/go-grpc-bidirectional-streaming-example/src/proto"

	"google.golang.org/grpc"
)

type server struct{}

func (s server) Max(srv pb.Math_MaxServer) error {

	log.Println("start new server")
	var size int
	ctx := srv.Context()

	for {

		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		req, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}

		// update max and send it to stream
		size = len(req.Data)
		resp := pb.Response{Size: int32(size)}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
		log.Printf("send size=%d", size)
	}
}

func main() {
	listenAddr := ":50005"
	if len(os.Args) > 1 {
		listenAddr = os.Args[1]
	}

	// create listiner
	log.Printf("listening on %v\n", listenAddr)
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create grpc server
	s := grpc.NewServer(
		grpc.InitialWindowSize(10000000),
		grpc.ReadBufferSize(10000000),
		grpc.MaxMsgSize(10000000),
		grpc.WriteBufferSize(10000000),
		grpc.InitialConnWindowSize(10000000))
	pb.RegisterMathServer(s, server{})

	// and start...
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
