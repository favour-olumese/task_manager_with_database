// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify

package mocks

import (
	"context"

	mock "github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
)

// NewMockUserUsecase creates a new instance of MockUserUsecase. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockUserUsecase(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockUserUsecase {
	mock := &MockUserUsecase{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockUserUsecase is an autogenerated mock type for the UserUsecase type
type MockUserUsecase struct {
	mock.Mock
}

type MockUserUsecase_Expecter struct {
	mock *mock.Mock
}

func (_m *MockUserUsecase) EXPECT() *MockUserUsecase_Expecter {
	return &MockUserUsecase_Expecter{mock: &_m.Mock}
}

// Login provides a mock function for the type MockUserUsecase
func (_mock *MockUserUsecase) Login(ctx context.Context, username string, password string) (string, error) {
	ret := _mock.Called(ctx, username, password)

	if len(ret) == 0 {
		panic("no return value specified for Login")
	}

	var r0 string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) (string, error)); ok {
		return returnFunc(ctx, username, password)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = returnFunc(ctx, username, password)
	} else {
		r0 = ret.Get(0).(string)
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = returnFunc(ctx, username, password)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserUsecase_Login_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Login'
type MockUserUsecase_Login_Call struct {
	*mock.Call
}

// Login is a helper method to define mock.On call
//   - ctx
//   - username
//   - password
func (_e *MockUserUsecase_Expecter) Login(ctx interface{}, username interface{}, password interface{}) *MockUserUsecase_Login_Call {
	return &MockUserUsecase_Login_Call{Call: _e.mock.On("Login", ctx, username, password)}
}

func (_c *MockUserUsecase_Login_Call) Run(run func(ctx context.Context, username string, password string)) *MockUserUsecase_Login_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockUserUsecase_Login_Call) Return(s string, err error) *MockUserUsecase_Login_Call {
	_c.Call.Return(s, err)
	return _c
}

func (_c *MockUserUsecase_Login_Call) RunAndReturn(run func(ctx context.Context, username string, password string) (string, error)) *MockUserUsecase_Login_Call {
	_c.Call.Return(run)
	return _c
}

// Register provides a mock function for the type MockUserUsecase
func (_mock *MockUserUsecase) Register(ctx context.Context, username string, password string) (*mongo.InsertOneResult, error) {
	ret := _mock.Called(ctx, username, password)

	if len(ret) == 0 {
		panic("no return value specified for Register")
	}

	var r0 *mongo.InsertOneResult
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) (*mongo.InsertOneResult, error)); ok {
		return returnFunc(ctx, username, password)
	}
	if returnFunc, ok := ret.Get(0).(func(context.Context, string, string) *mongo.InsertOneResult); ok {
		r0 = returnFunc(ctx, username, password)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*mongo.InsertOneResult)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = returnFunc(ctx, username, password)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// MockUserUsecase_Register_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Register'
type MockUserUsecase_Register_Call struct {
	*mock.Call
}

// Register is a helper method to define mock.On call
//   - ctx
//   - username
//   - password
func (_e *MockUserUsecase_Expecter) Register(ctx interface{}, username interface{}, password interface{}) *MockUserUsecase_Register_Call {
	return &MockUserUsecase_Register_Call{Call: _e.mock.On("Register", ctx, username, password)}
}

func (_c *MockUserUsecase_Register_Call) Run(run func(ctx context.Context, username string, password string)) *MockUserUsecase_Register_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *MockUserUsecase_Register_Call) Return(insertOneResult *mongo.InsertOneResult, err error) *MockUserUsecase_Register_Call {
	_c.Call.Return(insertOneResult, err)
	return _c
}

func (_c *MockUserUsecase_Register_Call) RunAndReturn(run func(ctx context.Context, username string, password string) (*mongo.InsertOneResult, error)) *MockUserUsecase_Register_Call {
	_c.Call.Return(run)
	return _c
}
