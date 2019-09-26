/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

//go:generate protoc -I ../helloworld --go_out=plugins=grpc:../helloworld ../helloworld/helloworld.proto

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"net"

	"github.com/golang/glog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// ================================
// gRPC 服务实现
const (
	port = ":50051"
)

type server struct{}

func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	glog.V(8).Infof("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

// ================================

// LoggingInterceptor 实现unary拦截器
func LoggingInterceptor(ctx context.Context, req interface{},
	info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (
	interface{}, error) {
	glog.V(8).Infof("request: %s, %v", info.FullMethod, req)
	resp, err := handler(ctx, req)
	glog.V(8).Infof("response: %s, %v", info.FullMethod, resp)
	return resp, err
}

// 启动服务端
// $ go run main.go
func main() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "10") // 输出10以下的log
	flag.Parse()
	defer glog.Flush()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			LoggingInterceptor,
		),
	}

	s := grpc.NewServer(opts...)

	pb.RegisterGreeterServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		glog.Fatalf("failed to serve: %v", err)
	}
}
