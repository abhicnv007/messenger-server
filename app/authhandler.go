package app

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	gorillacon "github.com/gorilla/context"

	"github.com/abhicnv007/whistle/user"
	"google.golang.org/appengine"
)

func getUIDContext(r *http.Request) int64 {
	return gorillacon.Get(r, "UID").(int64)
}

func setUIDContext(r *http.Request, val int64) {
	gorillacon.Set(r, "UID", val)
}

/*
	Checks the BasicAuth of the request and if not found, returns error

	Expected input: base64 encoded `{uid} : {secret}` in Authorization Header
*/
func authFn(w http.ResponseWriter, r *http.Request) error {

	u, secret, ok := r.BasicAuth()
	if ok != true {
		//w.Header().Add("WWW-Authenticate", "Basic realm=\"Whistle Api\"")
		return errors.New("Basic Auth")
	}

	//BasicAuth returns uid and secret as strings, so parse them
	uid, err := strconv.ParseInt(u, 10, 64)

	if err != nil {
		log.Println("Could not parse the uid in auth handler", err)
		return errors.New("Invalid uid")
	}

	c := appengine.NewContext(r)
	_, err = user.IsValidSecret(c, uid, secret)
	if err != nil {
		return err
	}

	setUIDContext(r, uid)

	return nil
}
