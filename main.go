// Написать мини сервис с разделением слоев в одном main.go файле. Сервис должен уметь:
// 1. Подключаться к базе данных
// 2. Использовать кэш c применением Proxy паттерна
// 3. Принимать http запросы REST like API

// 4. Регистрировать пользователя в базе данных
// 5. Выводить список всех пользователей

// 6. У пользователя следующие данные email, password, name, age

// 7. Запретить регистрацию пользователей с одинаковым email и возрастом меньше 18 лет

package main

import (
	"database/sql"

	"github.com/go-chi/chi/v5"

	"log"
	"net/http"
)

// internal/models
type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Age      int    `json:"age"`
}

type UserCache map[string]User

// cmd
func main() {
	app := Injection()
	app.Run()

}

// internal/repo
// db
func CreateTableINE(db *sql.DB) error {
	_, err := db.Exec(`
  CREATE TABLE IF NOT EXISTS users (
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    age INT NOT NULL
  );
  `)
	return err
}

func DBConnect() (DBController, error) {
	db, err := sql.Open("postgress", "host=db port=5432 user=p password=pp5 dbname=pp sslmode=disable")
	if err != nil {
		log.Println(err)
		return DBController{}, err
	}
	err = CreateTableINE(db)
	if err != nil {
		log.Println(err)
		return DBController{}, err
	}
	return DBController{db}, nil
}

// other

type DBController struct {
	Db *sql.DB
}

func (db *DBController) RAddUser(user User) error

func (db *DBController) RListUser() ([]User, error)

// internal/service
type RepoIface interface {
	RAddUser(user User) error
	RListUser() ([]User, error)
}

type ServiceUser struct {
	RepoUSer RepoIface
}

func NewServiceUser(r RepoIface) *ServiceUser {
	return &ServiceUser{r}
}

func (s *ServiceUser) SAddUser(user User) error

func (s *ServiceUser) SListUser() ([]User, error)

// internal/controller
// handlers.go
type ServiceIface interface {
	SAddUser(user User) error
	SListUser() ([]User, error)
}

type ControllerUser struct {
	ServiceUser ServiceIface
}

func NewControllerUser(s ServiceIface) *ControllerUser {
	return &ControllerUser{s}
}

func (c *ControllerUser) SAddUser() http.HandlerFunc

func (c *ControllerUser) SListUser() http.HandlerFunc

// proxy.go
type ProxyUser struct {
	C     *ControllerUser
	cache *UserCache
}

func NewProxyUser(c *ControllerUser) *ProxyUser {
	users := make(UserCache)
	return &ProxyUser{c, &users}
}

func (p *ProxyUser) SAddUser() http.HandlerFunc

/*
реализуем проверку пользователей по эмейлу и возрасту
эмэйл делаем ключем кэша, возраст - значением
если возраст не подходит то не добавляем, если подходит
берем значение из кэша и если оно есть, то не добавляем юзера,
если нет, то кэшируем и добавляем
*/

func (p *ProxyUser) SListUser() http.HandlerFunc

/*
получаем юзеров и возвращаем,
сверяем результат с кэшем,
если в кэше есть юзеры которых нет в базе
добавляем в базу
*/

// routes.go
func InitRoutes(p *ProxyUser) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/user", p.SAddUser())
	r.Get("/user", p.SListUser())
	return r
}

// internal/controller/app
// app.go
type App struct {
	ProxyUser *ProxyUser
}

func (a *App) Run() {
	routs := InitRoutes(a.ProxyUser)
	http.ListenAndServe(":8080", routs)
}

func Injection() *App {
	db, err := DBConnect()
	if err != nil {
		log.Println(err)
		return nil
	}
	return &App{
		ProxyUser: NewProxyUser(NewControllerUser(NewServiceUser(&db))),
	}
}
