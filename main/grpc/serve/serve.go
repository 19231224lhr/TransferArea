package main

import (
	"context"
	"log"
	"net"

	pb "transfer/grpc/proto"
	"transfer/interconnected"

	"github.com/ethereum/go-ethereum/common"
	"google.golang.org/grpc"
)

const (
	port = ":1145"
)

type server struct {
	*pb.UnimplementedTransferGRPCServer
}

func (s *server) ToTransferCommit(ctx context.Context, in *pb.ToTransferRequest) (*pb.ToTransferReply, error) { // 实现具体方法
	log.Println("收到了一个调用请求")
	// 调用相关函数
	interconnected.ToTransfer(common.BytesToAddress(in.FromAddress), common.BytesToAddress(in.BAddress), int(in.Amount))
	return &pb.ToTransferReply{Result: true}, nil
}

func main() {
	lis, err := net.Listen("tcp", port) // 监听器
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()                       // 服务器实例
	pb.RegisterTransferGRPCServer(s, &server{}) // 将服务器实例注册到服务器上
	if err := s.Serve(lis); err != nil {        // 启动服务器并监听
		log.Fatalf("failed to serve: %v", err)
	}
}
