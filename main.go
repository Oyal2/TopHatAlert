package main

import (
	"encoding/json"
	"fmt"
	"github.com/oyal2/TopHatAlert/TopHat"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

func main(){
	var err error
	jar,_ := cookiejar.New(nil)
	client := http.Client{
		Jar: jar,
	}
	th := TopHat.TopHatInfo{
		Webhook:   "",
		CourseID: 0,
		Client:   &client,
		Now: time.Time{},
		Omit: make(map[interface{}]struct{}),
	}

	fmt.Print("Enter Course ID: ")
	_, err = fmt.Scanln(&th.PublicCode)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Print("Enter Webhook: ")
	_, err = fmt.Scanln(&th.Webhook)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if ok := scrapeInfo(&th); !ok {
		err = th.Login()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	fmt.Println("Logged in Successfully!")

	th.Now = time.Now().UTC()

	className,err := th.GrabClasses()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Monitoring Class:",className)

	for true {
		err = th.Monitor()
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(4000 * time.Millisecond)
	}
}

func scrapeInfo(th *TopHat.TopHatInfo) bool {
	f, err := os.OpenFile("settings.json",os.O_CREATE|os.O_RDONLY,0777)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer f.Close()
	fInfo,err := f.Stat()
	dat := make([]byte,fInfo.Size())

	_, err = f.Read(dat)
	if err != nil {
		return false
	}

	if len(dat) == 0 {
		return false
	}

	var data TopHat.SettingsInfo
	err = json.Unmarshal(dat, &data)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if len(data.SessionID) == 0 {
		return false
	}

	u,err := url.Parse("https://app.tophat.com/")
	var c []*http.Cookie
	c = append(c, &http.Cookie{Name: "sessionid", Value: data.SessionID, Domain: "app.tophat.com"})
	th.Client.Jar.SetCookies(u,c)

	return true
}