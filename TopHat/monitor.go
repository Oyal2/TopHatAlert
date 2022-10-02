package TopHat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (th *TopHatInfo) GrabClasses() (string,error) {
	res, err := th.Client.Get("https://app.tophat.com/enrollments")
	if err != nil {
		return "",err
	}
	defer res.Body.Close()

	bData,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "",err
	}

	if len(bData) == 0 || string(bData) == "{\"detail\":\"Authentication credentials were not provided.\"}"{
		return "",errors.New("Could not fetch class information")
	}

	var classes []struct {
		Available            bool   `json:"available"`
		CourseCode           string `json:"course_code"`
		CourseID             int    `json:"course_id"`
		CourseName           string `json:"course_name"`
		IsMultiSectionCourse bool   `json:"is_multi_section_course"`
		NumberOfSections     int    `json:"number_of_sections"`
		OrgID                int    `json:"org_id"`
		OrgName              string `json:"org_name"`
		Profs                []struct {
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Name      string `json:"name"`
			UserID    int    `json:"user_id"`
		} `json:"profs"`
		PublicCode string        `json:"public_code"`
		Role       string        `json:"role"`
		Sections   []interface{} `json:"sections"`
	}
	err = json.Unmarshal(bData, &classes)
	if err != nil {
		return "",err
	}

	for _,class := range classes {
		//Check if the class is avail
		if class.PublicCode == th.PublicCode{
			th.CourseID = class.CourseID
			return class.CourseName,nil
		}
	}

	return "",errors.New("Could not find class with code")
}

func (th *TopHatInfo) Monitor() error {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://app.tophat.com/api/v3/course/%d/student_viewable_module_item/",th.CourseID), nil)

	req.Header.Add("authority", "app.tophat.com")
	req.Header.Add("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("sec-gpc", "1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")

	res, err := th.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return errors.New("Monitor Network Error: " + res.Status)
	}

	bData,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var questions struct {
		Meta struct {
			Next     interface{} `json:"next"`
			Previous interface{} `json:"previous"`
			Offset   int         `json:"offset"`
		} `json:"meta"`
		TotalResults int `json:"total_results"`
		Objects      []struct {
			DisplayName     string      `json:"display_name"`
			ID              int         `json:"id"`
			LastActivatedAt string      `json:"last_activated_at"`
			ModuleID        string      `json:"module_id"`
			ResourceURI     string      `json:"resource_uri"`
			Scheduled       bool        `json:"scheduled"`
			Status          string      `json:"status"`
			ParentFileID    interface{} `json:"parent_file_id"`
		} `json:"objects"`
	}
	err = json.Unmarshal(bData, &questions)
	if err != nil {
		return err
	}

	for _,question := range questions.Objects {
		tm, err := time.Parse("2006-01-02T15:04:05+0000",question.LastActivatedAt)
		_,ok := th.Omit[strconv.Itoa(question.ID)]
		if err == nil && th.Now.Before(tm) && (question.ModuleID == "question" || question.ModuleID == "learning_tool") && !ok {
			if question.ModuleID == "question"{
				err := th.questionParse(strconv.Itoa(question.ID))
				if err != nil {
					return err
				}
			} else {
				err := th.learningToolParse(strconv.Itoa(question.ID))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (th *TopHatInfo) questionParse(id string) (err error) {
	defer func(err error) {
		if err != nil {
			panic(err)
		}
	}(err)


	req, _ := http.NewRequest("GET", "https://app.tophat.com/api/v1/question/" + id, nil)

	req.Header.Add("authority", "app.tophat.com")
	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("sec-gpc", "1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")

	res, err := th.Client.Do(req)

	if res.StatusCode >= 400 {
		return errors.New("Question Network Error: " + res.Status)
	}

	bData,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var data struct {
		AllowPartialGrading bool        `json:"allow_partial_grading"`
		Choices             []string    `json:"choices"`
		ContainerModuleItem interface{} `json:"container_module_item"`
		Course              string      `json:"course"`
		CustomData          struct {
			AllCorrect bool `json:"all_correct"`
			IsSegment  bool `json:"is_segment"`
			NumericBlanks []struct {
				Keyword   string  `json:"keyword"`
				Tolerance float64 `json:"tolerance"`
			} `json:"numeric_blanks"`
			WordBlanks []struct {
				CaseSensitive bool   `json:"case_sensitive"`
				Keyword       string `json:"keyword"`
			} `json:"word_blanks"`
			ImageURL      string `json:"image_url"`
			LimitAttempts bool   `json:"limit_attempts"`
			NumAttempts   int    `json:"num_attempts"`
		} `json:"custom_data"`
		CustomImageWidth         interface{} `json:"custom_image_width"`
		HasCorrectAnswer         bool        `json:"has_correct_answer"`
		ID                       string      `json:"id"`
		ImageAltText             string      `json:"image_alt_text"`
		ImageFullTextDescription string      `json:"image_full_text_description"`
		ImageThumbnailURL        interface{} `json:"image_thumbnail_url"`
		ImageURL                 string 	 `json:"image_url"`
		IsAnonymous              bool        `json:"is_anonymous"`
		LastActivatedAt          string      `json:"last_activated_at"`
		LastDeactivatedAt        time.Time   `json:"last_deactivated_at"`
		LayoutType               string      `json:"layout_type"`
		Learning                 bool        `json:"learning"`
		Profile                  struct {
			IsTeamItem        bool        `json:"is_team_item"`
			IsTimed           bool        `json:"is_timed"`
			MaxAttempts       interface{} `json:"max_attempts"`
			ShowCorrectAnswer bool        `json:"show_correct_answer"`
		} `json:"profile"`
		Question    string `json:"question"`
		ResourceURI string `json:"resource_uri"`
		Status      string `json:"status"`
		Title       string `json:"title"`
		Type        string `json:"type"`
	}
	err = json.Unmarshal(bData, &data)
	if err != nil {
		return err
	}
	th.Omit[id] = struct{}{}

	title := data.Title
	question := CleanText(data.Question)
	image := data.ImageURL
	fields := make(map[string]string)
	fmt.Printf("[%s] Found Question: %s",time.Now().Format("01/02/06  3:04:05 PM"),question)

	switch data.Type {
	case "wa": {
		fields["Question"] = question
		return th.SendWebhook(title,"Word Answer",fields,image)
	}
	case "mc":{
		fields["Question"] = question
		for i,choice := range data.Choices {
			fields[fmt.Sprintf("%d)",i+1)] = CleanText(choice)
		}

		return th.SendWebhook(title,"Multiple Choice",fields,image)
	}
	case "na": {
		fields["Question"] = question
		return th.SendWebhook(title,"Numeric Answer",fields,image)
	}
	case "fitbq": {
		fields["Question"] = CleanBlanks(question)
		return th.SendWebhook(title,"Fill in Blank",fields,image)
	}
	case "match": {
		fields["Question"] = CleanBlanks(question)
		for i,choice := range data.Choices {
			fields[fmt.Sprintf("%d)",i+1)] = CleanText(choice)
		}
		return th.SendWebhook(title,"Fill in Blank",fields,image)
	}
	case "target": {
		fields["Question"] = CleanBlanks(question)
		image = data.CustomData.ImageURL
		fields["Number of Attempts"] = strconv.Itoa(data.CustomData.NumAttempts)
		return th.SendWebhook(title,"Click on Target",fields,image)
	}
	case "sort": {
		fields["Question"] = CleanBlanks(question)
		for i,choice := range data.Choices {
			fields[fmt.Sprintf("%d)",i+1)] = CleanText(choice)
		}
		return th.SendWebhook(title,"Sorting",fields,image)
	}

	}

	return nil
}

func (th *TopHatInfo) learningToolParse(id string) (err error){
	defer func(err error) {
		if err != nil {
			panic(err)
		}
	}(err)


	req, _ := http.NewRequest("GET", "https://app.tophat.com/learning_tool/api/v1/learning_tool/" + id, nil)

	req.Header.Add("authority", "app.tophat.com")
	req.Header.Add("accept", "application/json, text/plain, */*")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("cache-control", "max-age=0")
	req.Header.Add("sec-fetch-dest", "document")
	req.Header.Add("sec-fetch-mode", "navigate")
	req.Header.Add("sec-fetch-site", "none")
	req.Header.Add("sec-fetch-user", "?1")
	req.Header.Add("sec-gpc", "1")
	req.Header.Add("upgrade-insecure-requests", "1")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")

	res, err := th.Client.Do(req)

	if res.StatusCode >= 400 {
		return errors.New("Question Network Error: " + res.Status)
	}

	bData,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var data struct {
		ID                         string      `json:"id"`
		CanDisplayAnswerFeedback   bool        `json:"can_display_answer_feedback"`
		CanShowCorrectAnswer       bool        `json:"can_show_correct_answer"`
		ContainerModuleItem        interface{} `json:"container_module_item"`
		CourseID                   int         `json:"course_id"`
		ExternalID                 string      `json:"external_id"`
		HasConfigureView           bool        `json:"has_configure_view"`
		HasCorrectAnswer           bool        `json:"has_correct_answer"`
		IsAutoGraded               bool        `json:"is_auto_graded"`
		IsPresentable              bool        `json:"is_presentable"`
		IsQuestionService          bool        `json:"is_question_service"`
		IsReviewable               bool        `json:"is_reviewable"`
		IsSupportedInMobile        bool        `json:"is_supported_in_mobile"`
		IsSubmittedThroughTophat   bool        `json:"is_submitted_through_tophat"`
		IsFeCommunicationSupported bool        `json:"is_fe_communication_supported"`
		LastActivatedAt            string      `json:"last_activated_at"`
		LastDeactivatedAt          string      `json:"last_deactivated_at"`
		Learning                   bool        `json:"learning"`
		LearningToolID             int         `json:"learning_tool_id"`
		LearningToolType           string      `json:"learning_tool_type"`
		ProviderID                 int         `json:"provider_id"`
		ParameterValues            interface{} `json:"parameter_values"`
		Profile                    struct {
			MaxAttempts       interface{} `json:"max_attempts"`
			IsTeamItem        bool        `json:"is_team_item"`
			ShowCorrectAnswer bool        `json:"show_correct_answer"`
		} `json:"profile"`
		ShouldOpenOnLoad bool        `json:"should_open_on_load"`
		Status           string      `json:"status"`
		Title            string      `json:"title"`
		ToolURL          string      `json:"tool_url"`
		Tags             interface{} `json:"tags"`
	}
	err = json.Unmarshal(bData, &data)
	if err != nil {
		return err
	}
	th.Omit[id] = struct{}{}
	title := data.Title
	fmt.Printf("[%s] Found Learning Tool: %s",time.Now().Format("01/02/06  3:04:05 PM"),data.Title)

	switch data.LearningToolType {
		case "learnosity_chemistry_formula": {
			return th.SendWebhook(title,"Chemistry Response",nil,"")
		}
		case "learnosity_math_formula": {
			return th.SendWebhook(title,"Math Response",nil,"")
		}
		case "graded_calculation-gradedcalculationquestion": {
			return th.SendWebhook(title,"Math Response",nil,"")
		}
	}

	return nil
}

func (th *TopHatInfo) SendWebhook(title,t string, items map[string]string, image string) error {
	if len(th.Webhook) == 0 {
		return nil
	}

	fields := make([]Fields,len(items))
	if len(items) > 0 {
		fields[0].Name = "Question"
		fields[0].Value = items["Question"]
		fields[0].Inline = false
	}

	i := 1
	for k,v := range items{
		if k != "Question" {
			fields[i].Name = k
			fields[i].Value = v
			fields[i].Inline = true
			i++
		}
	}

	e := Embeds{
		Author: Author{},
		Title:     fmt.Sprintf("%s (%s)",title,t),
		URL:       fmt.Sprintf("https://app.tophat.com/e/%s/lecture/",th.PublicCode),
		Color:     4500823,
		Thumbnail: Thumbnail{},
		Fields:    fields,
		Footer:    Footer{IconURL: "https://tophat.com/wp-content/uploads/2017/05/tophat.png", Text: time.Now().Format("01/02/06  3:04:05 PM")},
	}

	if len(image) > 0 {
		e.Image.URL = image
	}

	webhook := Webhook{
		Username:  "TopHatAlert",
		AvatarURL: "",
		Embeds:    []Embeds{e},
	}

	b,err := json.Marshal(webhook)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest("POST", th.Webhook, bytes.NewReader(b))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/100.0.4896.127 Safari/537.36")
	res,err := th.Client.Do(req)
	bData,err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(bData))

	return nil
}

func CleanText(text string) string {
	fIndex := strings.Index(text,"\">")
	if fIndex == -1 {
		return text
	}
	lIndex := strings.Index(text, "</p>")
	if lIndex == -1 {
		return text
	}

	return text[fIndex + 2:lIndex]
}

func CleanBlanks(text string) string {
	s := strings.Split(text,"&lt;")
	result := s[0] + " "
	for i,blanks := range s {
		if i > 0 {
			blank := strings.Split(blanks,"&gt;")
			if len(blank) == 2 {
				for i := 0; i < len(blank[0]);i++ {
					result += "_"
				}
				result += " " + blank[1]
			}
		}
	}

	return result
}