package response

import (
	"encoding/json"
	"net/http"
)

//Link is a link
type Link struct {
	Href string `json:"href"`
}

//AllThreads are all threads
type AllThreads struct {
	Threads []Link `json:"threads"`
}

//Message Type for storing IMs
type Message struct {
	Link
	ParentThread Link   `json:"threadid"`
	From         Link   `json:"from"`
	Content      string `json:"content"`
	Time         string `json:"time"`
}

//Thread represents a conversation
type Thread struct {
	Link
	Participants []Link `json:"participants"`
}

//User Defines a user
type User struct {
	Link
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

//Response is a  struct
type Response struct {
	w http.ResponseWriter
}

//New responds with json
func New(w http.ResponseWriter) *Response {
	var r Response
	w.Header().Set("content-type", "application/json")
	r.w = w
	return &r
}

/*WithCode returns response with code
[SUPER BIG THING] : ALWAYS CALL WithCode BEFORE WithData, else while writing data,
	it automatically sets header to StatusOK and starts sending the request
*/
func (r *Response) WithCode(code int) *Response {
	r.w.WriteHeader(code)
	return r
}

/*WithData returns with data
[SUPER BIG THING] : ALWAYS CALL WithCode BEFORE WithData, else while writing data,
	it automatically sets header to StatusOK and starts sending the request
*/
func (r *Response) WithData(data interface{}) *Response {
	json.NewEncoder(r.w).Encode(data)
	return r
}

//func (r *Response) WithMessage()

type errorJSON struct {
	Error string `json:"error"`
}

//Error Returns an error json
func (r *Response) Error(err string) *Response {
	json.NewEncoder(r.w).Encode(errorJSON{
		Error: err,
	})
	return r
}
