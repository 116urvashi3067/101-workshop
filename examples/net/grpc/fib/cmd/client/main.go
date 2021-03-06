package main

import (
	"context"
	"io"
	"sync"
	"time"

	pb "algogrit.com/fib-grpc/fibonacci"
	empty "github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func getFirstN(c pb.FibonacciClient) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	log.Info("First 10 fibonacci numbers...")
	r, err := c.GetFirstN(ctx, &pb.FirstNQuery{N: 10})

	if err != nil {
		log.Fatalf("could not get first 10 fib numbers: %v", err)
	}

	// log.Debug("Numbers: ", r)
	for _, el := range r.Values {
		log.Info(el.Value)
	}
}

func getNext(c pb.FibonacciClient) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	nextVal, err := c.GetNext(ctx, &empty.Empty{})

	if err != nil {
		log.Fatalf("could not get next number: %v", err)
	}

	log.Info("Next number: ", nextVal.Value)
}

func streamValues(c pb.FibonacciClient, wg *sync.WaitGroup) {
	defer wg.Done()
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	stream, _ := c.Stream(ctx, &empty.Empty{})

	for {
		nextVal, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Errorf("Error streaming: %v\n", err)
			break
		}

		log.Info("Received next number: ", nextVal.Value)
	}
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewFibonacciClient(conn)

	var wg sync.WaitGroup
	wg.Add(1)

	go streamValues(c, &wg)

	go getFirstN(c)
	go getNext(c)

	wg.Wait()
}
