package app

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/abhicnv007/messenger-server/messaging"
	"github.com/abhicnv007/messenger-server/response"
	"github.com/abhicnv007/messenger-server/user"
)

func encodeThread(t messaging.Thread) response.Thread {
	var rt response.Thread

	rt.Href = threadsURI + "/" + strconv.FormatInt(t.ThreadID, 10)

	for _, i := range t.Participants {
		l := response.Link{
			Href: userURI + "/" + strconv.FormatInt(i, 10),
		}
		rt.Participants = append(rt.Participants, l)
	}
	return rt
}

func decodeThread(rt response.Thread, new bool) (messaging.Thread, error) {
	var t messaging.Thread

	var err error
	if t.ThreadID, err = getIDFromLink(rt.Href); err != nil && !new {
		return t, errors.New("Invalid threadid")
	}

	for _, i := range rt.Participants {
		p, err := getIDFromLink(i.Href)
		if err != nil {
			return t, errors.New("Invalid Participant")
		}
		t.Participants = append(t.Participants, p)
	}

	return t, nil
}

func encodeMessage(m messaging.Message) response.Message {
	var rm response.Message

	rm.ParentThread.Href = threadsURI + "/" + strconv.FormatInt(m.ParentThread, 10)
	rm.Href = rm.ParentThread.Href + "/" + "messages" + "/" + strconv.FormatInt(m.MessageID, 10)
	rm.From.Href = userURI + "/" + strconv.FormatInt(m.From, 10)
	rm.Content = m.Content
	rm.Time = m.Time

	return rm
}

func decodeMessage(rm response.Message, new bool) (messaging.Message, error) {
	var m messaging.Message
	var err error

	if !new {
		if m.ParentThread, err = getIDFromLink(rm.ParentThread.Href); err != nil {
			return m, errors.New("Invalid parentthread")
		}

		if m.From, err = getIDFromLink(rm.From.Href); err != nil {
			return m, errors.New("Invalid from")
		}

		if m.MessageID, err = getIDFromLink(rm.From.Href); err != nil {
			return m, errors.New("Invalid messageid")
		}
	}

	m.Content = rm.Content
	//Just making sure that the time sent is not bogus
	if _, err := time.Parse(time.RFC3339, rm.Time); err != nil {
		return m, errors.New("Invalid time")
	}
	m.Time = rm.Time

	return m, nil
}

func encodeUser(u user.User) response.User {
	var ru response.User
	ru.Href = userURI + "/" + strconv.FormatInt(u.UID, 10)
	ru.Name = u.Name
	ru.Secret = u.SecretKey
	return ru
}

func decodeUser(ru response.User, new bool) (user.User, error) {
	var u user.User
	var err error
	if u.UID, err = getIDFromLink(ru.Href); err != nil && !new {
		return u, errors.New("Invalid uid")
	}
	u.Name = ru.Name
	u.SecretKey = ru.Secret
	return u, nil
}

func encodeAllThreads(k []int64) response.AllThreads {
	var rt response.AllThreads

	for _, i := range k {
		t := response.Link{
			Href: threadsURI + "/" + strconv.FormatInt(i, 10),
		}
		rt.Threads = append(rt.Threads, t)
	}
	return rt
}

func getIDFromLink(link string) (int64, error) {
	s := strings.Split(link, "/")
	i, err := strconv.ParseInt(s[len(s)-1], 10, 64)
	if err != nil {
		return i, errors.New("Could not parse id")
	}
	return i, nil
}

func removeDuplicates(elements []int64) []int64 {
	// Use map to record duplicates as we find them.
	encountered := map[int64]bool{}
	result := []int64{}

	for v := range elements {
		if encountered[elements[v]] != true {
			// Record this element as an encountered element.
			encountered[elements[v]] = true
			// Append to result slice.
			result = append(result, elements[v])
		}
	}
	// Return the new slice.
	return result
}
