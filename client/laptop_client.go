package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/trongtb88/secure-grpc-in-go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type LaptopClient struct {
	service pb.LaptopServiceClient
}

func NewLaptopClient(conn *grpc.ClientConn) *LaptopClient {
	service := pb.NewLaptopServiceClient(conn)
	return &LaptopClient{service: service}
}

func (client *LaptopClient) CreateLaptop(laptop *pb.Laptop) (*pb.CreateLaptopResponse, error) {
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.service.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			// not a big deal
			log.Print("laptop already exists")
		} else {
			log.Fatal("cannot create laptop: ", err)
		}
		return nil, err
	}
	log.Printf("created laptop with id: %s", resp.Id)
	return resp, nil
}

func (client *LaptopClient) SearchLaptop(filter *pb.Filter, found func(laptop *pb.SearchLaptopResponse)) error {
	req := &pb.SearchLaptopRequest{Filter: filter}

	// set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.SearchLaptop(ctx, req)
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		found(resp)
	}
}

func found(resp *pb.SearchLaptopResponse) {
	foundLaptop := resp.GetLaptop()
	log.Print("- found: ", foundLaptop.GetId())
	log.Print("  + brand: ", foundLaptop.GetBrand())
	log.Print("  + name: ", foundLaptop.GetName())
	log.Print("  + cpu cores: ", foundLaptop.GetCpu().GetNumberCores())
	log.Print("  + cpu min ghz: ", foundLaptop.GetCpu().GetMinGhz())
	log.Print("  + ram: ", foundLaptop.GetRam())
	log.Print("  + price: ", foundLaptop.GetPriceUsd())
}

// UploadImage calls upload image RPC
func (client *LaptopClient) UploadImage(laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info to server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err)
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot receive response: ", err)
	}

	log.Printf("image uploaded with id: %s, size: %d", res.GetId(), res.GetSize())
}

// RateLaptop calls rate laptop RPC
func (client *LaptopClient) RateLaptop(laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.service.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitResponse := make(chan error)
	// go routine to receive responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more responses")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}

			log.Print("received response: ", res)
		}
	}()

	// send requests
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Print("sent request: ", req)
	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send: %v", err)
	}

	err = <-waitResponse
	return err
}
