package v1

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
	v1.UnimplementedTodoServiceServer
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

func (s *todoServiceServer) Read(ctx context.Context, req *v1.ReadRequest) (*v1.ReadResponse, error) {
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

	// Query Todo by ID
	rows, err := c.QueryContext(ctx, "SELECT id, title, description, reminder FROM todo WHERE id = ?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from todo-> "+err.Error())
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve data from todo-> "+err.Error())
		}
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with ID='%d' is not found",
			req.Id))
	}

	// Get Todo data
	var td v1.Todo
	var reminder time.Time
	if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve field values from Todo row-> "+err.Error())
	}

	td.Reminder, err = ptypes.TimestampProto(reminder)
	if err != nil {
		return nil, status.Error(codes.Unknown, "reminder field has invalid format-> "+err.Error())
	}

	if rows.Next() {
		return nil, status.Error(codes.Unknown, fmt.Sprintf("found multiple Todo rows with ID='%d'",
			req.Id))
	}

	return &v1.ReadResponse{
		Api:  API_VERSION,
		Todo: &td,
	}, nil
}

// Update todo task
func (s *todoServiceServer) Update(ctx context.Context, req *v1.UpdateRequest) (*v1.UpdateResponse, error) {
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

	// Update todo
	res, err := c.ExecContext(ctx, "UPDATE todo SET title=?, description=?, reminder=? WHERE id=?",
		req.Todo.Title, req.Todo.Description, reminder, req.Todo.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to update Todo-> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value-> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with ID='%d' is not found",
			req.Todo.Id))
	}

	return &v1.UpdateResponse{
		Api:     API_VERSION,
		Updated: rows,
	}, nil
}

// Delete todo task
func (s *todoServiceServer) Delete(ctx context.Context, req *v1.DeleteRequest) (*v1.DeleteResponse, error) {
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

	// Delete Todo
	res, err := c.ExecContext(ctx, "DELETE FROM todo WHERE id=?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to delete Todo-> "+err.Error())
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve rows affected value-> "+err.Error())
	}

	if rows == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("Todo with ID='%d' is not found",
			req.Id))
	}

	return &v1.DeleteResponse{
		Api:     API_VERSION,
		Deleted: rows,
	}, nil
}

// Read all todo tasks
func (s *todoServiceServer) ReadAll(ctx context.Context, req *v1.ReadAllRequest) (*v1.ReadAllResponse, error) {
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

	// Get Todo list
	rows, err := c.QueryContext(ctx, "SELECT id, title, description, reminder FROM todo")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from Todo-> "+err.Error())
	}
	defer rows.Close()

	var reminder time.Time
	list := []*v1.Todo{}
	for rows.Next() {
		td := new(v1.Todo)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Todo row-> "+err.Error())
		}
		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder field has invalid format-> "+err.Error())
		}
		list = append(list, td)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from Todo-> "+err.Error())
	}

	return &v1.ReadAllResponse{
		Api:   API_VERSION,
		Todos: list,
	}, nil
}

// Read all todo tasks by title
func (s *todoServiceServer) ReadByTitle(ctx context.Context, req *v1.ReadByTitleRequest) (*v1.ReadByTitleResponse, error) {
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

	// Get Todo list
	rows, err := c.QueryContext(ctx, "SELECT id, title, description, reminder FROM todo WHERE title LIKE ?", "%"+req.Title+"%")
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to select from Todo-> "+err.Error())
	}
	defer rows.Close()

	var reminder time.Time
	list := []*v1.Todo{}
	for rows.Next() {
		td := new(v1.Todo)
		if err := rows.Scan(&td.Id, &td.Title, &td.Description, &reminder); err != nil {
			return nil, status.Error(codes.Unknown, "failed to retrieve field values from Todo row-> "+err.Error())
		}
		td.Reminder, err = ptypes.TimestampProto(reminder)
		if err != nil {
			return nil, status.Error(codes.Unknown, "reminder field has invalid format-> "+err.Error())
		}
		list = append(list, td)
	}

	if err := rows.Err(); err != nil {
		return nil, status.Error(codes.Unknown, "failed to retrieve data from Todo-> "+err.Error())
	}

	return &v1.ReadByTitleResponse{
		Api:   API_VERSION,
		Todos: list,
	}, nil
}
