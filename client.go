package main

import (
	"context"
	"fmt"
	"log"
	"micro/proto"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func main() {
	start := time.Now()

	// Create a gRPC client connection
	conn, err := grpc.Dial(":50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	// Create a new MyService client
	client := proto.NewMyServiceClient(conn)

	// Define the number of tasks to send
	numTasks := 1000000

	// Define the number of workers
	numWorkers := 100000

	// Create channels for tasks and results
	taskChan := make(chan *proto.MyRequest, numTasks)
	resultChan := make(chan *proto.MyResponse, numTasks)

	// Start the workers
	var wg sync.WaitGroup
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			for task := range taskChan {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := client.MyMethod(ctx, task)
				if err != nil {
					log.Fatalf("failed to call MyMethod: %v", err)
				}

				resultChan <- resp
			}

			wg.Done()
		}()
	}

	// Send the tasks
	for i := 0; i < numTasks; i++ {
		taskChan <- &proto.MyRequest{Name: "Alice" + strconv.Itoa(i)}
	}

	// Close the task channel to signal that all tasks have been sent
	close(taskChan)

	// Collect the results
	go func() {
		for resp := range resultChan {
			// Print the response message
			fmt.Println(resp.Message)
		}
	}()

	// Wait for all workers to complete
	wg.Wait()

	// Close the result channel to signal that all results have been collected
	close(resultChan)

	fmt.Printf("Elapsed time: %s\n", time.Since(start))
}
