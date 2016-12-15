package main

import (
	"log"
	"fmt"
)

type User struct {
	Name  string
	Email string
}

type Admin struct {
	User
	Level string
}

func (u *User) Notify() error {
	log.Printf("User: Sending User Email To %s<%s>\n",
		u.Name,
		u.Email)

	return nil
}

//func (a *Admin) Notify() error {
//	log.Printf("User: Sending Admin Email To %s<%s>\n",
//		a.Name,
//		a.Email)
//
//	return nil
//}

type Notifier interface {
	Notify() error
}

func SendNotification(notify Notifier) error {
	return notify.Notify()
}

func main() {
	admin := &Admin{
		User: User{
			Name:  "john smith",
			Email: "john@email.com",
		},
		Level: "super",
	}

	admin.Notify()

	fmt.Println("1:", 1 << 3)
}
