package v1

import (
	"context"
	"database/sql"

	v1 "github.com/devararishivian/go-grpc/pkg/api/v1"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	API_VERSION = "v1"
)

// todoServiceServer is implementation of v1.TodoServiceServer proto interface
type todoServiceServer struct {
	db *sql.DB
}

func NewTodoServiceServer(db *sql.DB) v1.TodoServiceServer {
	return &todoServiceServer{db: db}
}

// checkAPI checks if the API version requested by client is supported by server
func (s *todoServiceServer) checkAPI(api string) error {
	// API version is "" means use the current version of the service
	if len(api) > 0 {
		if api != API_VERSION {
			return status.Errorf(codes.Unimplemented, "unsupported API version: service implements API version '%s', but asked for '%s'", API_VERSION, api)
		}
	}

	return nil
}

// connect returns SQL database connection from the pool
func (s *todoServiceServer) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := s.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
	}

	return c, nil
}

// Create new todo task
func (s *todoServiceServer) Create(ctx context.Context, req *v1.CreateRequest) (*v1.CreateResponse, error) {
	// check if the API version requested by client is supported by server
	if err := s.checkAPI(req.Api); err != nil {
		return nil, err
	}

	// Get SQL connection from pool
	c, err := s.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	reminder, err := ptypes.Timestamp(req.Todo.Reminder)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "reminder field has invalid format-> "+err.Error())
	}

	// Insert Todo entity data
	res, err := c.ExecContext(ctx, "INSERT INTO todo(title, description, reminder) VALUES(?,?,?)", req.Todo.Title, req.Todo.Description, reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to insert into todo-> "+err.Error())
	}

	// Get ID of created Todo
	id, err := res.LastInsertId()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve id for created Todo-> "+err.Error())
	}

	return &v1.CreateResponse{
		Api: API_VERSION,
		Id:  id,
	}, nil
}
