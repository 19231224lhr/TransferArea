package main

import (
	"context"
	"log"
	pb "transfer/grpc/proto"

	"google.golang.org/grpc"
)

const (
	address = "localhost:1145"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure()) // 建立与 gRPC 服务器的连接
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTransferGRPCClient(conn) // 创建了一个 gRPC 客户端实例 client，用于与服务器进行通信
	// Contact the server and print out its response.
	FromAddress := make([]byte, 5) // 创建一个长度为5的字节切片
	FromAddress[0] = 'H'
	FromAddress[1] = 'e'
	FromAddress[2] = 'l'
	FromAddress[3] = 'l'
	FromAddress[4] = 'o'

	BAddress := make([]byte, 5) // 创建一个长度为5的字节切片
	BAddress[0] = 'W'
	BAddress[1] = 'o'
	BAddress[2] = 'r'
	BAddress[3] = 'l'
	BAddress[4] = 'd'
	r, err := c.ToTransferCommit(context.Background(), &pb.ToTransferRequest{FromAddress: FromAddress, BAddress: BAddress, Amount: 3}) // 调用 client.Send 方法向服务器发送请求
	if err != nil {
		log.Fatalf("连接轻计算区grpc接口失败: %v", err)
	}
	log.Printf("返回结果: %s", r.GetResult()) // 不用管返回值，加返回值是因为返回空值需要下载一个包
}
