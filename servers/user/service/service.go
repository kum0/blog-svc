package service

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Zhan9Yunhua/blog-svr/services/validator"

	"github.com/Zhan9Yunhua/blog-svr/common"
	"github.com/Zhan9Yunhua/blog-svr/services/email"
	"github.com/Zhan9Yunhua/blog-svr/utils"
	"github.com/gomodule/redigo/redis"
)

type UserServicer interface {
	Login(loginRequest) (string, error)
	GetUser(string) (string, error)
	SendCode() error
	Register(registerRequest) error
	Validate(interface{}) []error
}

func NewUserService(mdb *sql.DB, rd *redis.Pool, email *email.Email) *UserService {
	return &UserService{
		mdb,
		rd,
		email,
		nil,
	}
}

type UserService struct {
	mdb       *sql.DB
	rd        *redis.Pool
	email     *email.Email
	validator *validator.Validator
}

func (u *UserService) GetUser(s string) (string, error) {
	if s == "" {
		return "", common.ErrEmpty
	}
	return strings.ToUpper(s), nil
}

// 登录
func (u *UserService) Login(params loginRequest) (string, error) {
	return "", nil
}

// 注册
func (u *UserService) Register(params registerRequest) error {
	fmt.Printf("%+v\n", params)

	return nil
}

func (u *UserService) SendCode() error {
	uuid, err := utils.NewUUID()
	if err != nil {
		return nil
	}

	code := utils.NewRand(6)

	rc := u.rd.Get()
	defer rc.Close()

	ch := make(chan error)

	go func(c chan<- error) {
		if _, err := rc.Do("SET", uuid, code, "EX", 600); err != nil {
			c <- err
		}
		c <- nil
	}(ch)

	html := fmt.Sprintf(`
      <html>
      <body>
	  <h3>
      注册码: %d
      </h3>
      </body>
      </html>
      `, code)
	go func(c chan<- error) {
		c <- u.email.Send("zy.hua1122@outlook.com", "注册码", html)
	}(ch)

	n := 2
	for c := range ch {
		n--
		if c != nil {
			close(ch)
			return c
		}
		if n == 0 {
			close(ch)
		}
	}

	return nil
}

func (u *UserService) Validate(a interface{}) []error {
	if u.validator == nil {
		u.validator = validator.NewValidator()
	}

	return u.validator.LazyValidate(a)
}
