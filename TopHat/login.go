package TopHat

import (
	"encoding/json"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"net/http"
	"net/url"
	"os"
	"time"
)

func (th *TopHatInfo) Login() error {
	launch := launcher.
		New().
		Devtools(false).
		NoSandbox(true).
		Headless(false)

	uri := launch.MustLaunch()
	browser := rod.
		New().
		ControlURL(uri)

	err := browser.Connect()
	if err != nil {
		return err
	}
	defer browser.Close()

	page := browser.MustPage().Timeout(1 * time.Minute)

	page.MustNavigate("https://app.tophat.com/login")
	fmt.Println("Waiting for you to login...")

	cont := true
	for cont {
		time.Sleep(500 * time.Millisecond)
		messages := browser.Context(page.GetContext()).Event()
		for msg := range messages {
			t := proto.TargetTargetInfoChanged{}
			if msg.Load(&t) {
				if t.TargetInfo.URL == "https://app.tophat.com/e" {
					cont = false
					break
				}
			}
		}
	}

	cookies, err := page.Cookies([]string{})
	var c []*http.Cookie
	var sessionID string
	for _, v := range cookies {
		if v.Name == "sessionid"{
			sessionID = v.Value
		}
		c = append(c, &http.Cookie{Name: v.Name, Value: v.Value, Domain: v.Domain})
	}
	u,err := url.Parse("https://app.tophat.com/")
	if err != nil {
		return err
	}
	th.Client.Jar.SetCookies(u,c)

	err = saveSettings(sessionID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func saveSettings(sessionID string) error {
	var settings SettingsInfo
	settings.SessionID = sessionID
	bData, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile("settings.json", bData, 0755)
	if err != nil {
		return err
	}
	return nil
}