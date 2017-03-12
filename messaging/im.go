package messaging

import (
	"log"

	"errors"

	"golang.org/x/net/context"

	"google.golang.org/appengine/datastore"
)

//Message Type for storing IMs
type Message struct {
	ParentThread int64  `json:"parentthread"`
	MessageID    int64  `json:"messageid"`
	From         int64  `json:"from"`
	Content      string `json:"content"`
	Time         string `json:"time"`
}

//Thread represents a conversation
type Thread struct {
	ThreadID     int64   `json:"threadid"`
	Participants []int64 `json:"participants"`
}

//CheckIfParticipant Returns true if i in the participants
func (t *Thread) CheckIfParticipant(i int64) bool {
	for _, tid := range t.Participants {
		if i == tid {
			return true
		}
	}
	return false
}

//AllThreads represents all threads of conversation a user has
type AllThreads struct {
	UserID  int64   `json:"userid"`
	Threads []int64 `json:"threads"`
}

//NewThread Creates a new thread
func NewThread(c context.Context, t Thread) (Thread, error) {

	l, _, err := datastore.AllocateIDs(c, "Thread", nil, 1)
	if err != nil {
		log.Println(err)
		return t, err
	}
	//Add the thread to the datastore
	k := datastore.NewKey(c, "Thread", "", l, nil)
	t.ThreadID = l
	key, err := datastore.Put(c, k, &t)
	if err != nil {
		log.Print("Put thread failed ", err)
		return t, err
	}

	var keys []*datastore.Key
	var lc []*AllThreads
	//Add the thread to all Conversations maintained by the Participants
	for _, p := range t.Participants {
		var con AllThreads

		k = datastore.NewKey(c, "AllThreads", "", p, nil)

		if err := datastore.Get(c, k, &con); err != nil {
			con.UserID = p
			con.Threads = []int64{key.IntID()}
		} else {
			con.Threads = append(con.Threads, key.IntID())
		}

		keys = append(keys, k)
		lc = append(lc, &con)

	}

	if _, err := datastore.PutMulti(c, keys, lc); err != nil {
		log.Print("Put failed", err)
		return t, err
	}

	return t, nil

}

//GetAllThreads gets all threads that the user participates in
func GetAllThreads(c context.Context, participant int64) ([]int64, error) {
	k := datastore.NewKey(c, "AllThreads", "", participant, nil)
	var con AllThreads
	if err := datastore.Get(c, k, &con); err == datastore.ErrNoSuchEntity {
		return nil, err
	}

	return con.Threads, nil

}

//GetThread gets the thread with the given threadID
func GetThread(c context.Context, tid int64) (Thread, error) {
	//q := datastore.NewQuery("Thread").Filter("ThreadID =", tid)
	k := datastore.NewKey(c, "Thread", "", tid, nil)
	//datastore.AllocateIDs(c, kind, parent, n)
	var t Thread
	datastore.Get(c, k, &t)
	return t, nil
}

//GetMessage gets single message
func GetMessage(c context.Context, tid int64, mid int64) (Message, error) {
	key := datastore.NewKey(c, "Message", "", mid, getThreadKey(c, tid))
	var msg Message
	if err := datastore.Get(c, key, &msg); err != nil {
		return msg, errors.New("No such message found")
	}
	return msg, nil
}

//GetMessages Gets messages from storage
func GetMessages(c context.Context, tid int64, t string, num int) ([]Message, error) {
	if num == 0 && t == "" {
		var m []Message
		_, err := datastore.NewQuery("Message").Ancestor(getThreadKey(c, tid)).
			Limit(10).Order("-Time").GetAll(c, &m)
		return m, err
	} else if num == 0 {
		var m []Message
		_, err := datastore.NewQuery("Message").Ancestor(getThreadKey(c, tid)).
			Filter("Time >", t).Limit(10).GetAll(c, &m)
		return m, err
	} else if t == "" {
		var m []Message
		_, err := datastore.NewQuery("Message").Ancestor(getThreadKey(c, tid)).
			Limit(num).Order("-Time").GetAll(c, &m)
		return m, err
	}
	return nil, errors.New("Invalid Input")
}

func getThreadKey(c context.Context, tid int64) *datastore.Key {
	return datastore.NewKey(c, "Thread", "", tid, nil)
}

//InsertMessage inserts message
func InsertMessage(c context.Context, m *Message) error {
	//Get the parent thread first, then add the message
	k := datastore.NewKey(c, "Thread", "", m.ParentThread, nil)
	var t Thread

	//[TODO Return errors to correspond with the HTTP Error codes]
	if err := datastore.Get(c, k, &t); err == datastore.ErrNoSuchEntity {
		return err
	}
	l, _, err := datastore.AllocateIDs(c, "Message", k, 1)

	if err != nil {
		return err
	}

	//[TODO Hopefully change this and make the message ids linear for a thread]
	key := datastore.NewKey(c, "Message", "", l, k)
	m.MessageID = l
	//Set the parent key k
	datastore.Put(c, key, m)

	return nil
}
