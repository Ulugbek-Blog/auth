package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	pb "blog-auth/genproto/userservice"
	logger "blog-auth/internal/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UsersStorage interface {
	Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error)
	RegisterUser(ctx context.Context, req *pb.RegisterUserReq) (*pb.RegisterUserRes, error)
	ForgotPassword(ctx context.Context, req *pb.ForgotPasswordReq) (*pb.ForgotPasswordRes, error)
	UpdateUser(ctx context.Context, req *pb.UpdateUserReq) (*pb.UpdateUserRes, error)
	VerifyEmail(ctx context.Context, req *pb.VerifyEmailReq) (*pb.VerifyEmailRes, error)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

type userStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) UsersStorage {
	return &userStorage{db: db}
}

func (s *userStorage) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	query := `
		SELECT
			id,
			name,
			lastname,
			email,
			password,
			role,
			created_at,
			updated_at		
		FROM
			users
		WHERE
			email = $1 AND deleted_at = 0
	`
	res := pb.LoginRes{
		UserRes: &pb.UserModel{},
	}
	var password string
	err = s.db.QueryRow(query, req.Email).Scan(
		&res.UserRes.Id,
		&res.UserRes.Fname,
		&res.UserRes.Lname,
		&res.UserRes.Email,
		&password,
		&res.UserRes.Role,
		&res.UserRes.CreatedAt,
		&res.UserRes.UpdatedAt,
	)
	if err != nil {
		logs.Error("Error with scan:", zap.Error(err))
		return nil, err
	}
	ok := checkPasswordHash(req.Password, password)
	if !ok {
		return nil, fmt.Errorf("password not correct")
	}

	return &pb.LoginRes{UserRes: res.UserRes}, err

}

func (s *userStorage) RegisterUser(ctx context.Context, req *pb.RegisterUserReq) (*pb.RegisterUserRes, error) {
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO users (
			id, fname, lname, email, password
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, 
			CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, 0
		);`

	hashpass, err := hashPassword(req.Password)
	if err != nil {
		logs.Error("Error with create user")
		return nil, err
	}
	id := uuid.NewString()
	_, err = s.db.ExecContext(ctx, query, id, req.Fname, req.Lname, req.Email, hashpass)
	if err != nil {
		logs.Error("Error with create user")
		return nil, err
	}

	query1 := `select id, fname, lname, email,bio, phone, role, created_at, updated_at from users where id = $1`
	user := pb.UserModel{}
	err = s.db.QueryRowContext(ctx, query1, id).Scan(&user.Id, &user.Fname, &user.Lname, &user.Email,&user.Bio,&user.Phone, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logs.Error("Error getting user", zap.Error(err))
		return nil, err
	}
	if err != nil {
		logs.Error("Error with create user")
		return nil, err

	}
	userM := &pb.UserModel{
		Id:        user.Id,
		Fname:     user.Fname,
		Lname:     user.Lname,
		Email:     user.Email,
		Bio:       user.Bio,
		Phone:     user.Phone,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return &pb.RegisterUserRes{UserRes: userM}, nil
}

func (s *userStorage) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordReq) (*pb.ForgotPasswordRes, error) {
	//logs, err := logger.NewLogger()
	//if err != nil {
	//
	//	return nil, err
	//}
	return nil, nil
}

func (s *userStorage) UpdateUser(ctx context.Context, req *pb.UpdateUserReq) (*pb.UpdateUserRes, error) {
	query := "UPDATE users SET"
	var args []interface{}
	var updates []string
	argCounter := 1
	logs, err := logger.NewLogger()
	if err != nil {
		return nil, err
	}
	// Check for fields and build query
	if req.UserReq.Fname != "string" && req.UserReq.Fname != "" {
		updates = append(updates, " fname = $"+strconv.Itoa(argCounter))
		args = append(args, req.UserReq.Fname)
		argCounter++
	}

	if req.UserReq.Lname != "string" && req.UserReq.Lname != "" {
		updates = append(updates, " lname = $"+strconv.Itoa(argCounter))
		args = append(args, req.UserReq.Lname)
		argCounter++
	}

	if req.UserReq.Email != "string" && req.UserReq.Email != "" {
		updates = append(updates, " email = $"+strconv.Itoa(argCounter))
		args = append(args, req.UserReq.Email)
		argCounter++
	}

	if req.UserReq.Bio != "string" && req.UserReq.Bio != "" {
		updates = append(updates, " bio = $"+strconv.Itoa(argCounter))
		args = append(args, req.UserReq.Bio)
		argCounter++
	}

	if req.UserReq.Phone != "string" && req.UserReq.Phone != "" {
		updates = append(updates, " phone = $"+strconv.Itoa(argCounter))
		args = append(args, req.UserReq.Phone)
		argCounter++
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	query += " " + strings.Join(updates, ", ") + " WHERE id = $" + strconv.Itoa(argCounter)
	args = append(args, req.UserReq.Id)

	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		logs.Error("Error updating user: ", zap.Error(err))
		return nil, err
	}

	return &pb.UpdateUserRes{UserRes: req.UserReq}, nil
}

func (s *userStorage) VerifyEmail(ctx context.Context, req *pb.VerifyEmailReq) (*pb.VerifyEmailRes, error) {
	return nil, nil
}
