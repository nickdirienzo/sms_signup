package main

import (
	"fmt"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

type Signup struct {
	Name      string    `bson:"name" json:"name"`
	Email     string    `bson:"email" json:"email"`
	Raw       string    `bson:"raw" json:"raw"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

var mgoSession *mgo.Session

func SendSMS(w http.ResponseWriter, r *http.Request, message string) {
	w.Header().Set("Content-Type", "text/xml")
	s := fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?><Response><Sms>%s</Sms></Response>", message)
	fmt.Fprint(w, s)
}

func SMSHandler(w http.ResponseWriter, r *http.Request) {
	sms := r.URL.Query().Get("Body")
	sms = strings.Trim(sms, "8575-2838")
	if sms == "" {
		log.Println("Blank body.")
		SendSMS(w, r, "Please send a text with: \"Name\" \"email@address\"")
		return
	}
	info := strings.Fields(sms)
	name := strings.Join(info[:len(info)-1], " ")
	email := info[len(info)-1]
	name = strings.Trim(name, "\"")
	if name == "" {
		log.Println("Blank name.")
		SendSMS(w, r, "Please send a text with: \"Name\" \"email@address\"")
		return
	}
	email = strings.Trim(email, "\"")
	address := fmt.Sprintf("%s <%s>", name, email)
	_, err := mail.ParseAddress(address)
	if err != nil {
		SendSMS(w, r, "Please try again with a valid email address.")
		return
	}
	signup := Signup{Name: name, Email: email, Raw: sms, Timestamp: time.Now()}
	log.Println(signup)
	c := mgoSession.DB("sms_signup").C("signups")
	err = c.Insert(signup)
	if err != nil {
		log.Println(err.Error())
	}
	SendSMS(w, r, fmt.Sprintf("Thanks for signing up, %s!", info[0]))
}

func main() {
	var err error
	mgoSession, err = mgo.Dial("localhost")
	if err != nil {
		log.Println(err.Error())
	}
	log.Println("Connected to mongodb.")
	log.Println("Starting on :8080")
	http.HandleFunc("/sms", SMSHandler)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Println(err.Error())
	}
}
