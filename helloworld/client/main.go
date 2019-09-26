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

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "world"
)

// 启动客户端
// $ go run main.go
func main() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "10") // 输出10以下的log
	flag.Parse()
	defer glog.Flush()

	// 第一步：创建连接GRPC Server的信道
	// 还可以调用DialOptions设置认证信息再传入Dial
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		glog.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// 第二步：创建GRPC调用的客户端
	c := pb.NewGreeterClient(conn)

	// 第三步：通过GRPC客户端调用服务接口
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		glog.Fatalf("could not greet: %v", err)
	}
	glog.V(8).Infof("Greeting: %s", r.GetMessage())
}
