package user

import (
	"log"
	"math/rand"
	"time"

	"golang.org/x/net/context"

	"errors"

	"google.golang.org/appengine/datastore"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString generates random strings of length 64 [a-zA-Z0-9]
func generateRandomString() string {
	b := make([]byte, 64)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

//User Defines a user
type User struct {
	UID       int64  `json:"uid"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	SecretKey string
}

//InsertNew inserts the given user into the database and returns it
func InsertNew(c context.Context, u User) (User, error) {

	//[TODO] Make the username key
	if c, err := datastore.NewQuery("User").Filter("Name =", u.Name).
		Count(c); err != nil {
		return u, err
	} else if c != 0 {
		return u, errors.New("Username already exists")
	}

	l, _, err := datastore.AllocateIDs(c, "User", nil, 1)

	if err != nil {
		log.Println(err)
		return u, err
	}

	key := datastore.NewKey(c, "User", "", l, nil)

	u.SecretKey = generateRandomString()
	u.UID = l
	_, err = datastore.Put(c, key, &u)
	if err != nil {
		log.Println(err)
	}

	return u, nil

}

//Check takes name and password
func Check(c context.Context, name string, pass string) (User, error) {
	var u []User
	datastore.NewQuery("User").Filter("Name =", name).GetAll(c, &u)

	if len(u) != 1 {
		return User{}, errors.New("User does not exist")
	}
	if u[0].Password != pass {
		return User{}, errors.New("Username or password invalid")
	}
	return u[0], nil
}

//IsValidSecret checks if the user is valid, if is then send the uid
func IsValidSecret(c context.Context, uid int64, sec string) (User, error) {

	var nu User
	k := datastore.NewKey(c, "User", "", uid, nil)
	if err := datastore.Get(c, k, &nu); err != nil {
		return User{}, errors.New("Invalid user")
	}

	//Compare all fields
	if sec != nu.SecretKey {
		return User{}, errors.New("Could not authenticate")
	}

	return nu, nil

}

//IsValidUser checks if the user is valid, if is then sends the user object back
func IsValidUser(c context.Context, u User) (User, error) {

	var nu User
	k := datastore.NewKey(c, "User", "", u.UID, nil)
	if err := datastore.Get(c, k, &nu); err != nil {
		return User{}, errors.New("Invalid user")
	}

	if nu.Name != u.Name || nu.SecretKey != u.SecretKey {
		return User{}, errors.New("Incorrect user credentials")
	}

	return nu, nil

}

//Get gets the user details from the the uid
func Get(c context.Context, uid int64) (User, error) {
	var nu User
	k := datastore.NewKey(c, "User", "", uid, nil)
	if err := datastore.Get(c, k, &nu); err == datastore.ErrNoSuchEntity {
		return User{}, errors.New("Invalid user")
	} else if err != nil {
		return User{}, errors.New("Could not fetch the details")
	}

	return nu, nil
}
