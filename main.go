package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type Question struct {
	ID      int      `json:"id"`
	Text    string   `json:"text"`
	Options []string `json:"options"`
	Correct []int    `json:"correct"`
	Image   string   `json:"image"` // Path to image (optional)
}

type ExamResponse struct {
	QuestionID int   `json:"question_id"`
	Selected   []int `json:"selected"` // IDs of selected options
}

var questions []Question

// Load the questions from a JSON file
func loadQuestions() {
	file, err := ioutil.ReadFile("questions.json")
	if err != nil {
		fmt.Println("Error reading questions file:", err)
		return
	}

	err = json.Unmarshal(file, &questions)
	if err != nil {
		fmt.Println("Error parsing questions JSON:", err)
		return
	}
}

// Save the questions back to the JSON file
func saveQuestions() {
	file, err := json.MarshalIndent(questions, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling questions:", err)
		return
	}

	err = ioutil.WriteFile("questions.json", file, 0644)
	if err != nil {
		fmt.Println("Error writing questions file:", err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func examHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/exam.html"))
	tmpl.Execute(w, questions)
}

func resultsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var responses []ExamResponse
	r.ParseForm()

	// Parse the responses from the form
	for _, question := range questions {
		selected := []int{}
		for i := range question.Options {
			key := fmt.Sprintf("q%d_%d", question.ID, i)
			if r.FormValue(key) == "on" {
				selected = append(selected, i)
			}
		}
		responses = append(responses, ExamResponse{
			QuestionID: question.ID,
			Selected:   selected,
		})
	}

	// Calculate the score
	score := 0
	total := len(questions)

	for _, response := range responses {
		for _, question := range questions {
			if response.QuestionID == question.ID {
				if len(response.Selected) == len(question.Correct) {
					correct := true
					for i, val := range response.Selected {
						if val != question.Correct[i] {
							correct = false
						}
					}
					if correct {
						score++
					}
				}
			}
		}
	}

	// Save the results temporarily to a JSON file
	file, _ := json.MarshalIndent(responses, "", "  ")
	_ = ioutil.WriteFile("responses.json", file, 0644)

	tmpl := template.Must(template.ParseFiles("templates/results.html"))
	tmpl.Execute(w, map[string]interface{}{
		"Score": score,
		"Total": total,
	})
}

func manageQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/questions.html"))
	tmpl.Execute(w, questions)
}

func addQuestionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		text := r.FormValue("text")
		options := r.Form["option"]
		correctStr := r.Form["correct"]
		image := r.FormValue("image")

		correct := []int{}
		for _, c := range correctStr {
			num, _ := strconv.Atoi(c)
			correct = append(correct, num)
		}

		newQuestion := Question{
			ID:      len(questions) + 1,
			Text:    text,
			Options: options,
			Correct: correct,
			Image:   image,
		}

		questions = append(questions, newQuestion)
		saveQuestions()
		http.Redirect(w, r, "/manage", http.StatusSeeOther)
	}
}

func main() {
	loadQuestions()

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/exam", examHandler)
	http.HandleFunc("/results", resultsHandler)
	http.HandleFunc("/manage", manageQuestionsHandler)
	http.HandleFunc("/add-question", addQuestionHandler)

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
