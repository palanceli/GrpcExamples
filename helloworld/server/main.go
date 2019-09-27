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
	"fmt"
	"net"
	"net/http"

	"github.com/golang/glog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// ================================
// Prometheus 支持

// PrometheusServer 封装prometheus服务
type PrometheusServer struct {
	Registry                *prometheus.Registry
	GrpcMetrics             *grpc_prometheus.ServerMetrics
	CustomizedCounterMetric *prometheus.CounterVec
	HTTPServer              *http.Server
}

// CreatePrometheusServer 工厂方法
func CreatePrometheusServer() (obj *PrometheusServer) {
	// 1. 创建metrics registry
	registry := prometheus.NewRegistry()
	// 2. 创建标准server metrics
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	// 3. 创建自定义metric
	customizedCounterMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "demo_server_say_hello_method_handle_count",
			Help: "Total number of RPCs handled on the server.",
		}, []string{"name"})
	// 4. 注册标准server metrics和自定义metric
	registry.MustRegister(grpcMetrics, customizedCounterMetric)
	customizedCounterMetric.WithLabelValues("Test")

	// 5. 创建HTTP server
	httpServer := &http.Server{
		Handler: promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
		Addr:    fmt.Sprintf("0.0.0.0:%d", 50052),
	}

	return &PrometheusServer{
		Registry:                registry,
		GrpcMetrics:             grpcMetrics,
		CustomizedCounterMetric: customizedCounterMetric,
		HTTPServer:              httpServer,
	}
}

// InitializeMetrics .
func (s *PrometheusServer) InitializeMetrics(grpcServer *grpc.Server) {
	s.GrpcMetrics.InitializeMetrics(grpcServer)
}

// Run .
func (s *PrometheusServer) Run() {
	go func() {
		if err := s.HTTPServer.ListenAndServe(); err != nil {
			glog.Fatalf("Unable to start a http server. err=%v", err)
		}
	}()
}

// ================================

// ================================
// gRPC 服务实现
const (
	port = ":50051"
)

// GRPCServer 我的GRPC服务
type GRPCServer struct {
	PrometheusServer *PrometheusServer
}

// SayHello .
func (s *GRPCServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	glog.V(8).Infof("Received: %v", in.GetName())
	s.PrometheusServer.CustomizedCounterMetric.WithLabelValues(in.Name).Inc()
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

	prometheusServer := CreatePrometheusServer()

	opts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(
			LoggingInterceptor,
			prometheusServer.GrpcMetrics.UnaryServerInterceptor(),
		),
	}

	s := grpc.NewServer(opts...)

	pb.RegisterGreeterServer(s, &GRPCServer{PrometheusServer: prometheusServer})

	prometheusServer.InitializeMetrics(s)
	prometheusServer.Run()

	if err := s.Serve(lis); err != nil {
		glog.Fatalf("failed to serve: %v", err)
	}
}
