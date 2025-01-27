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

// internal/entities
type User struct {
	Email    string
	Password string
	Name     string
	Age      int
}

type UserCache map[string]User

//cmd

// internal//repo
// db.go
type DBController struct {
	Db *sql.DB
}

func CreateTableINE(db *sql.DB) error {
	_, err := db.Exec(`
  CREATE TABLE IF NOT EXIST users (
    email VARCHAR NOT NULL,
    password VARCHAR NOT NULL,
    name VARCHAR NOT NULL,
    age INT NOT NULL
  );
  `)
	return err
}

func DBConnection() *DBController {
	db, err := sql.Open("postgress", "host=db port=5432 user=p password=pp5 dbname=pp sslmode=disable")
	if err != nil {
		log.Println(err)
		return nil
	}
	err = CreateTableINE(db)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &DBController{db}
}

func (db *DBController) RAddUser(user User) error

// реализация добавления юзера через инсерт инто

func (db *DBController) RListUser() ([]User, error)

// реализация собирания всех юзеров из бд в массив через селект

// proxy.go
type ProxyUser struct {
	DbContr *DBController
	Cache   *UserCache
}

func NewProxyUser(ui *DBController) *ProxyUser {
	users := make(UserCache)
	return &ProxyUser{ui, &users}
}

func (p *ProxyUser) StartCache() error

//берем юзеров из базы данных и кладем в кэш

func (p *ProxyUser) EndCache() error

//берем юзеров из кэша и возвращаем в базу данных

func (p *ProxyUser) RAddUser(user User) error

// реализовываем апроверку пользователя по параметрам если подходит кэшируем и добавляем в базу данных

func (p *ProxyUser) RListUser() ([]User, error)

//берем юзеров из кэша и возвращаем

// internal/service
type RepoIface interface {
	RAddUser(user User) error
	RListUser() ([]User, error)
}

type ServiceUser struct {
	PRepoUser *ProxyUser
}

func NewServiceUser(r *ProxyUser) *ServiceUser {
	return &ServiceUser{r}
}

func (s *ServiceUser) SAddUser(user User) error

func (s *ServiceUser) SListUser() ([]User, error)

// internal/controller
type ServiceIface interface {
	SAddUser(user User) error
	SListUser() ([]User, error)
}

type ControllerUser struct {
	ServiceUser *ServiceUser
}

func NewControllerUser(s *ServiceUser) *ControllerUser {
	return &ControllerUser{s}
}

func (c *ControllerUser) AddUSer() http.HandlerFunc
func (c *ControllerUser) ListUSer() http.HandlerFunc

type UserIface interface {
	AddUser() http.HandlerFunc
	ListUSer() http.HandlerFunc
}

//internal/controller/app
//ap.go

type App struct {
	UserIface *ControllerUser
}

func Inject() *App {
	db := DBConnection()
	proxy := NewProxyUser(db)
	proxy.StartCache()
	service := NewServiceUser(proxy)
	controller := NewControllerUser(service)

	return &App{
		UserIface: controller,
	}
}

func (a *App) Run() {

	routs := InitRoutes(a.UserIface)
	http.ListenAndServe(":8080", routs)
}

func (a *App) Stop() {
	a.UserIface.ServiceUser.PRepoUser.EndCache()
	//реализация корректного завершения работы

}

func InitRoutes(p *ControllerUser) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/user", p.AddUSer())
	r.Get("/user", p.ListUSer())
	return r
}

func main() {
	app := Inject()
	go app.Run()
	app.Stop()
}
