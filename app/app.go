package app

//[TODO] DO a check if the body is very large and if it is, ignore it

import (
	"io/ioutil"
	"log"
	"net/http"

	"google.golang.org/appengine"

	"encoding/json"

	"github.com/abhicnv007/messenger-server/handler"
	"github.com/abhicnv007/messenger-server/messaging"
	"github.com/abhicnv007/messenger-server/parse"
	"github.com/abhicnv007/messenger-server/response"
	"github.com/abhicnv007/messenger-server/user"
	"github.com/gorilla/mux"
)

const (
	threadsURI       = "/threads"
	singleThreadURI  = "/threads/{threadID}"
	allMessagesURI   = "/threads/{threadID}/messages"
	singleMessageURI = "/threads/{threadID}/messages/{messageID}"
	userURI          = "/users"

	userDetailsURI = "/users/{userID}"

	loginURI = "/login"
)

func init() {
	r := mux.NewRouter()

	handler.SetGlobalAuthFunc(authFn)

	r.Handle(userURI, handler.New(addUser).NoAuth()).Methods("POST")
	r.Handle(userURI, handler.New(login).NoAuth()).Methods("GET")

	r.Handle(userDetailsURI, handler.New(getUserDetails)).Methods("GET")

	r.Handle(threadsURI, handler.New(getAllThreads)).Methods("GET")
	r.Handle(threadsURI, handler.New(addThread)).Methods("POST")

	r.Handle(singleThreadURI, handler.New(getThread)).Methods("GET")

	r.Handle(allMessagesURI, handler.New(getAllMessages)).Methods("GET")
	r.Handle(allMessagesURI, handler.New(addMessage)).Methods("POST")

	r.Handle(singleMessageURI, handler.New(getMessage)).Methods("GET")

	http.Handle("/", r)

}

/*

addUser add the the user with details sent in JSON format to "/users" in POST

Request: POST at "/users" with JSON body containing name and password

Response, if successful, is a JSON object that contains the UID and the secret
that has to be set in Authorization header in every future requests

Testing-->

curl -i --request POST \
--header 'content-type: application/json' \
--url "localhost:8080/users" \
--data '{
    "name" : "Abhicnv002",
	"password" : "2344Hf"
  }'
*/
func addUser(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Could not read the user object")

		return
	}

	//User has to send password, so cannot use response.User object directly

	var ru struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if err = json.Unmarshal(b, &ru); err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Could not read the user object")

		return
	}

	//[TODO] Add limit checking

	var u user.User
	u.Name = ru.Name
	u.Password = ru.Password

	log.Println(u)
	if u, err = user.InsertNew(c, u); err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Could not create a new user")

		return
	}

	response.New(w).WithCode(http.StatusCreated).WithData(encodeUser(u))
}

/*

login checks the validity of user

Request: GET request to "/users" with name and password as quesry params

Response: if successful, (status 200/ StatusOK), is a JSON object that contains
the UID and the secret that has to be set in Authorization header in every future
requests
	If unsuccessful, StatusBadRequest/400

Testing-->

curl -i --request GET \
--header 'content-type: application/json' \
--url "localhost:8080/users?name=Abhicnv002&password=2344Hf"
*/
func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	r.ParseForm()
	name := r.FormValue("name")
	password := r.FormValue("password")

	u, err := user.Check(c, name, password)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error(err.Error())
		return
	}

	//log.Println(password, " ", u.Password)

	response.New(w).WithCode(http.StatusOK).
		WithData(encodeUser(u))
}

/*
getUser gets the detail of the user

Request: GET request at "/users"

Response: if successful, is a JSON object of the user

Testing-->

curl -i --request GET \
--header 'content-type: application/json' \
--header 'Authorization: Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--url "localhost:8080/users/2002"
*/
func getUserDetails(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	/*[TODO] Make a public and private profile, if auth uid matches the path uid,
	respond with private profile, else public profile
	*/
	//u := getUIDContext(r)
	uid, err := parse.MustGetInt64(mux.Vars(r)["userID"])
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).Error("Invalid UserID")
		return
	}
	u, err := user.Get(c, uid)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).Error(err.Error())
		return
	}

	response.New(w).WithCode(http.StatusOK).
		WithData(encodeUser(u))
}

/*
addThread used to create new threads of conversation

Request:  Is a POST with JSON body specifying all participants,
the caller needn't specify his/her own ID in the participants, it will be added
automatically and the Authorization header must be set

Response:  if successful, is a JSON body of the created thread with the threadid

Testing-->

curl -i --request POST \
--header 'Authorization : Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--header 'content-type: application/json' \
--url "localhost:8080/threads" \
--data '{
    "participants": [
	{"href" : "localhost:8080/users/2001"},
	{"href" : "localhost:8080/users/2002"},
	]
  }'
*/
func addThread(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	d, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Println("addThread could't read the body", err)
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Could not read the body")
		return
	}

	var rt response.Thread
	json.Unmarshal(d, &rt)

	t, err := decodeThread(rt, true)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).Error(err.Error())
		return
	}

	uid := getUIDContext(r)

	t.Participants = append(t.Participants, uid)
	t.Participants = removeDuplicates(t.Participants)

	//Check if every user mentioned is a valid user
	for _, pid := range t.Participants {
		if _, err = user.Get(c, pid); err != nil {
			response.New(w).WithCode(http.StatusBadRequest).
				Error("Invalid user given")
			return
		}
	}

	t, err = messaging.NewThread(c, t)
	if err != nil {
		response.New(w).WithCode(http.StatusInternalServerError).
			Error("Could not create a new thread")

		return
	}

	response.New(w).WithCode(http.StatusCreated).WithData(encodeThread(t))
}

/*
getThread used to get the details of a thread

Request: GET request at the uri without any query parameters and the Authorization
	header must be set

Response: Is a JSON body consisting of all participants in the thread

Testing -->

curl -i --request GET \
--header 'content-type: application/json' \
--header 'Authorization: Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--url "localhost:8080/threads/2006"
*/
func getThread(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	tid, err := parse.MustGetInt64(mux.Vars(r)["threadID"])
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid threadID")

		return
	}

	t, err := messaging.GetThread(c, tid)
	if err != nil {
		response.New(w).WithCode(http.StatusInternalServerError).
			Error("Could Not Get Thread")

		return
	}

	uid := getUIDContext(r)
	//If not a participant, not authorised
	if t.CheckIfParticipant(uid) == false {
		response.New(w).WithCode(http.StatusUnauthorized).
			Error("Not authorized as not a participant")

		return
	}

	response.New(w).WithCode(http.StatusOK).WithData(encodeThread(t))
}

/*
GetAllThreads used to get all threads of conversation a user is involved in

Request: GET request at the uri without any parameters and the Authorization
	header must be set

Response: Is a JSON body of all threads with their threadid

Testing -->

curl -i --request GET \
--header 'content-type: application/json' \
--header 'Authorization: Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--url "localhost:8080/threads"
*/
func getAllThreads(w http.ResponseWriter, r *http.Request) {

	c := appengine.NewContext(r)

	uid := getUIDContext(r)

	//k is nil if no threads present for the uid
	k, _ := messaging.GetAllThreads(c, uid)

	response.New(w).WithCode(http.StatusOK).WithData(encodeAllThreads(k))
}

/*
addMessage adds the given message to a threads

Request: POST request with JSON body with content and time (RFC3339)

Response: If successful, an 201 status is sent

Testing-->

curl -i --request POST \
--header 'Authorization : Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--header 'content-type: application/json' \
--url "localhost:8080/threads/2006/messages" \
--data '{
    "content": "Hey Man!!",
	"time": "2017-03-10T14:43:28+05:30"
  }'
*/
func addMessage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vals := mux.Vars(r)
	tid, err := parse.MustGetInt64(vals["threadID"])
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid threadid")
		return
	}

	//Body is a JSON Message
	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Add message body read error : ", err)
	}

	var rm response.Message
	if err = json.Unmarshal(d, &rm); err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Cannot parse body")
		return
	}

	m, err := decodeMessage(rm, true)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).Error(err.Error())
		return
	}

	m.From = getUIDContext(r)
	m.ParentThread = tid

	//log.Println(m)

	err = messaging.InsertMessage(c, &m)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("No thread with given id")
		return
	}

	response.New(w).WithCode(http.StatusCreated)
}

/*
getAllMessages used to get a bunch of messages based on a query

Request: GET request to "/threads/{threadID}/messages"
	Could include query parameters "limit" and "time"

Response: A JSON body with a list of uri of the messages that satisfy the query params

Testing -->

curl -i --request GET \
--header 'content-type: application/json' \
--header 'Authorization: Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--url "localhost:8080/threads/2006/messages?limit=10"
*/
func getAllMessages(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vals := mux.Vars(r)
	tid, err := parse.MustGetInt64(vals["threadID"])
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid threadid")
		return
	}

	r.ParseForm()

	num, err := parse.GetInt64(r.FormValue("limit"))
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid limit")

		return
	}

	t := r.FormValue("time")
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid time")

		return
	}

	m, err := messaging.GetMessages(c, tid, t, int(num))
	if err != nil {
		log.Println(err)
		response.New(w).WithCode(http.StatusInternalServerError).
			Error("Could not get message")

		return
	}

	//[TODO] Send just the MessageID href
	var rm []response.Message
	for _, mess := range m {
		rm = append(rm, encodeMessage(mess))
	}
	response.New(w).WithCode(http.StatusOK).WithData(rm)
}

/*
getMessage used to get a bunch of messages based on a query

Request: GET request to "/threads/{threadID}/messages/{messageID}"

Response: A JSON body of the message

Testing -->

curl -i --request GET \
--header 'content-type: application/json' \
--header 'Authorization: Basic MTAwNjp3U2ZxMUlFUk1qb0M4b0s0Vk9qM1NmQ2RteTA4Q1VXRTV6VHpyRGVja2F2VHptdkh5bnIxQ2U3bENxWkhGbFlH' \
--url "localhost:8080/threads/2006/messages/3002"
*/
func getMessage(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	vals := mux.Vars(r)
	tid, err := parse.MustGetInt64(vals["threadID"])
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid ThreadID")
		return
	}

	mid, err := parse.MustGetInt64(vals["messageID"])
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error("Invalid MessageID")

		return
	}

	msg, err := messaging.GetMessage(c, tid, mid)
	if err != nil {
		response.New(w).WithCode(http.StatusBadRequest).
			Error(err.Error())

		return
	}

	response.New(w).WithCode(http.StatusOK).WithData(encodeMessage(msg))
	return
}
