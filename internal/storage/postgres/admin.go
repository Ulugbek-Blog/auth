package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cast"

	pb "blog-auth/genproto/userservice"
	logger "blog-auth/internal/logger"

	"go.uber.org/zap"
)

type AdminStorage interface {
	CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error)
	GetUser(ctx context.Context, req *pb.GetUserByIDReq) (*pb.GetUserByIDRes, error)
	ForgetPassword(ctx context.Context, req *pb.ForgetPasswordReq) (*pb.ForgetPasswordRes, error)
	GetAllUsers(ctx context.Context, req *pb.GetAllUsersReq) (*pb.GetAllUsersRes, error)
	DeleteUser(ctx context.Context, req *pb.DeleteUserReq) (*pb.DeleteUserRes, error)
}
type AdminStorageImpl struct {
	db *sql.DB
}

func NewAdminStorage(db *sql.DB) AdminStorage {
	return &AdminStorageImpl{
		db: db,
	}
}

func (s *AdminStorageImpl) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserRes, error) {
	id := uuid.New().String()
	query := `INSERT INTO Users (id, fname, lname, email, password) VALUES ($1, $2, $3, $4, $5, $6)`
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	password, err := hashPassword(req.Password)
	if err != nil {
		logs.Error("Error hashing password", zap.Error(err))
	}
	_, err = s.db.ExecContext(ctx, query, id, req.Fname, req.Lname, req.Email, password)
	if err != nil {
		logs.Error("Error creating admin", zap.Error(err))
	}
	logs.Info("Successfully created user")
	user, err := s.GetUser(ctx, &pb.GetUserByIDReq{UserId: id})
	if err != nil {
		logs.Error("Error getting user", zap.Error(err))
	}
	logs.Info("Successfully got user")
	return &pb.CreateUserRes{AdminRes: user.UserRes}, nil
}

func (s *AdminStorageImpl) GetUser(ctx context.Context, req *pb.GetUserByIDReq) (*pb.GetUserByIDRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	query := `select id, fname, lname, email,bio, phone, role, created_at, updated_at from users where id = $1`
	user := pb.UserModel{}
	err = s.db.QueryRowContext(ctx, query, req.UserId).Scan(&user.Id, &user.Fname, &user.Lname, &user.Email, &user.Bio, &user.Phone, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logs.Error("Error getting user", zap.Error(err))
		return nil, err
	}
	logs.Info("Successfully got user")
	resp := pb.GetUserByIDRes{
		UserRes: &user,
	}
	return &resp, nil
}

func (s *AdminStorageImpl) ForgetPassword(ctx context.Context, req *pb.ForgetPasswordReq) (*pb.ForgetPasswordRes, error) {
	return nil, nil
}

func (s *AdminStorageImpl) GetAllUsers(ctx context.Context, req *pb.GetAllUsersReq) (*pb.GetAllUsersRes, error) {
	query := "SELECT id, name, lastname, email, O_CHAR(created_at, 'DD-MM-YYYY') AS created_at, updated_at FROM users WHERE deleted_at = 0"

	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	args := []interface{}{}
	argCounter := 1

	if req.User.Id != "string" && req.User.Id != "" {
		query += " AND id = $" + strconv.Itoa(argCounter)
		args = append(args, req.User.Id)
		argCounter++
	}

	if req.User.Email != "string" && req.User.Email != "" {
		query += " AND email = $" + strconv.Itoa(argCounter)
		args = append(args, req.User.Email)
		argCounter++
	}

	if req.User.Lname != "string" && req.User.Lname != "" {
		query += " AND lname = $" + strconv.Itoa(argCounter)
		args = append(args, req.User.Lname)
		argCounter++
	}

	if req.User.Fname != "string" && req.User.Fname != "" {
		query += " AND fname = $" + strconv.Itoa(argCounter)
		args = append(args, req.User.Fname)
		argCounter++
	}

	if req.User.CreatedAt != "string" && req.User.CreatedAt != "" {
		t1, err := time.Parse("01-02-2006", req.User.CreatedAt)
		if err != nil {
			logs.Error("Error parsing date")
			return nil, err
		}
		createdAtInSeconds := t1.Unix()
		query += " AND EXTRACT(EPOCH FROM created_at) > $" + strconv.Itoa(argCounter)

		args = append(args, createdAtInSeconds)
		argCounter++
	}

	// Add limit and offset to the query
	if req.Limit > 0 {
		query += " LIMIT $" + strconv.Itoa(argCounter)
		args = append(args, req.Limit)
		argCounter++
	}

	if req.Offset > 0 {
		query += " OFFSET $" + strconv.Itoa(argCounter)
		args = append(args, req.Offset)
		argCounter++
	}

	// Execute the query to get users
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		logs.Error("Error with get all users query")
		return nil, err
	}

	// Get total count of users
	countQuery := "SELECT COUNT(*) FROM users WHERE deleted_at = 0"
	var totalCount int
	err = s.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		logs.Error("Error getting total count of users")
		return nil, err
	}

	var admins []*pb.UserModel
	for rows.Next() {
		admin := pb.UserModel{}

		err = rows.Scan(&admin.Id, &admin.Fname, &admin.Lname, &admin.Email, &admin.Bio,
			&admin.Phone, &admin.CreatedAt, &admin.UpdatedAt, &admin.Role)
		if err != nil {
			logs.Error("Error with get all users")
		}
		admins = append(admins, &admin)
	}

	resp := pb.GetAllUsersRes{
		Users:      admins,
		TotalCount: cast.ToInt32(totalCount),
	}

	logs.Info("Successfully get all users")
	return &resp, nil
}

func (s *AdminStorageImpl) DeleteUser(ctx context.Context, req *pb.DeleteUserReq) (*pb.DeleteUserRes, error) {
	query := `update users set deleted_at = $1 where id = $2`
	t := time.Now().Unix()
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, query, t, req.UserId)
	if err != nil {
		logs.Error("Error while deleting admin")
	}

	logs.Info("Successfully deleted admin")
	return &pb.DeleteUserRes{Message: "Success"}, nil

}
