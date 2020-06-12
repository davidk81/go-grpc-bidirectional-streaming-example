package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"

	pb "github.com/pahanini/go-grpc-bidirectional-streaming-example/src/proto"

	"time"

	"google.golang.org/grpc"
)

func main() {
	rand.Seed(time.Now().Unix())

	// dail server
	conn, err := grpc.Dial(":50005", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewMathClient(conn)
	stream, err := client.Max(context.Background())
	if err != nil {
		log.Fatalf("openn stream error %v", err)
	}

	var size int32
	ctx := stream.Context()
	done := make(chan bool)
	timerStart := time.Now()

	// first goroutine sends random increasing numbers to stream
	// and closes int after 10 iterations
	go func() {
		for i := 1; i <= 1000; i++ {
			// generate random nummber and send it to stream
			buffer := make([]byte, 1000000)
			rand.Read(buffer)
			req := pb.Request{Data: buffer}
			if err := stream.Send(&req); err != nil {
				log.Fatalf("can not send %v", err)
			}
		}
		if err := stream.CloseSend(); err != nil {
			log.Println(err)
		}
	}()

	// second goroutine receives data from stream
	// and saves result in max variable
	//
	// if stream is finished it closes done channel
	go func() {
		totalReceived := float64(0)
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Fatalf("can not receive %v", err)
			}
			size = resp.Size
			totalReceived += float64(size)
			elapsed := time.Since(timerStart).Seconds()
			if elapsed > 0 {
				log.Printf("rate: %vbps", humanReadableCountSI(int64(totalReceived*8/elapsed)))
			}
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
	log.Printf("finished")
}

// converts number (typically bytes or bits) to human readable SI string
// eg: bytes=1200 -> 1.2k
// 	   bytes=1200000 0> 1.2M
// ported from https://stackoverflow.com/questions/3758606/how-to-convert-byte-size-into-human-readable-format-in-java
func humanReadableCountSI(bytes int64) string {
	if -1000 < bytes && bytes < 1000 {
		return fmt.Sprintf("%d ", bytes)
	}
	ci := "kMGTPE"
	idx := 0
	for bytes <= -999950 || bytes >= 999950 {
		bytes /= 1000
		idx++
	}
	return fmt.Sprintf("%.1f %c", float64(bytes)/1000.0, ci[idx])
}
