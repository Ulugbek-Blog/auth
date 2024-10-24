package service

import (
	"context"

	pb "blog-auth/genproto/userservice"
	logger "blog-auth/internal/logger"
	"blog-auth/internal/storage/postgres"
)

type AdminService interface {
	CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error)
	GetUser(ctx context.Context, req *pb.GetUserByIDReq) (*pb.GetUserByIDRes, error)
	ForgetPassword(ctx context.Context, req *pb.ForgetPasswordReq) (*pb.ForgetPasswordRes, error)
	GetAllUsers(ctx context.Context, req *pb.GetAllUsersReq) (*pb.GetAllUsersRes, error)
	DeleteUser(ctx context.Context, req *pb.DeleteUserReq) (*pb.DeleteUserRes, error)
}
type AdminServiceImpl struct {
	admin postgres.AdminStorage
	pb.UnimplementedAdminServiceServer
}

func NewAdminService(admin AdminService) *AdminServiceImpl {
	return &AdminServiceImpl{
		admin: admin,
	}
}

func (s *AdminServiceImpl) CreateAdmin(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	resp , err := s.admin.CreateUser(ctx, req)
	if err != nil {
		logs.Error("Error while calling CreateAdmin")
	}
	logs.Info("Successfully create admin")
	return resp, nil
}

func (s *AdminServiceImpl) GetUserByID(ctx context.Context, req *pb.GetUserByIDReq) (*pb.GetUserByIDRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	
	resp , err := s.admin.GetUser(ctx,req)
	if err != nil {
		logs.Error("Error while calling Get User")
	}
	logs.Info("Successfully get user")
	return resp, nil
}

func (s *AdminServiceImpl) ForgetPassword(ctx context.Context, req *pb.ForgetPasswordReq) (*pb.ForgetPasswordRes, error) {
	return nil, nil
}

func (s *AdminServiceImpl) GetAllAdmins(ctx context.Context, req *pb.GetAllUsersReq) (*pb.GetAllUsersRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	resp, err := s.admin.GetAllUsers(ctx, req)
	if err != nil {
		logs.Error("Error while calling GetAllUsers")
	}
	logs.Info("Successfully get all users")
	return resp, nil
}

func (s *AdminServiceImpl) DeleteUser(ctx context.Context, req *pb.DeleteUserReq) (*pb.DeleteUserRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	resp, err := s.admin.DeleteUser(ctx, req)
	if err != nil {
		logs.Error("Error while deleting admin")
	}
	logs.Info("Successfully delete admin")
	return resp,nil
}
