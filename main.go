package main

import (
	"fmt"
	"blog-auth/internal/config"
	"net"

	pb "blog-auth/genproto/userservice"
	logger "blog-auth/internal/logger"
	"blog-auth/internal/service"
	"blog-auth/internal/storage/postgres"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.Load()
	logs, err := logger.NewLogger()
	if err != nil {
		logs.Error("Error while initializing logger")
		return
	}
	db, err := postgres.ConnectPostgres()
	if err != nil {
		logs.Error("Error while initializing postgres connection")
	}
	defer db.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.AUTHHOST, cfg.AUTHPORT))
	if err != nil {
		logs.Error("Error while initializing listener")
	}

	defer listener.Close()
	logs.Info(fmt.Sprintf("Server start on port: %d", cfg.AUTHPORT))

	userStorage := postgres.NewUserStorage(db)
	userService := service.NewUserService(userStorage)

	adminStorage := postgres.NewAdminStorage(db)
	adminService := service.NewAdminService(adminStorage)

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, userService)
	pb.RegisterAdminServiceServer(s, adminService)

	if err := s.Serve(listener); err != nil {
		logs.Error("Error while initializing server")
	}
}
