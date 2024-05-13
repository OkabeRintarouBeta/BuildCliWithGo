package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type item struct {
	Task        string    `json:"task"`
	Done        bool      `json:"done"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
}

type List []item

func (l *List) Add(task string) {
	t := item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now().Round(0),
		CompletedAt: time.Time{}.Round(0),
	}
	*l = append(*l, t)
}

func (l *List) Complete(idx int) error {
	if idx >= len(*l) || idx < 0 {
		return errors.New("item not found")
	}
	(*l)[idx-1].Done = true
	(*l)[idx-1].CompletedAt = time.Now().Round(0)
	return nil
}

func (l *List) Delete(idx int) error {
	if idx >= len(*l) || idx < 0 {
		return errors.New("item not found")
	}
	*l = append((*l)[:idx-1], (*l)[idx:]...)
	return nil
}

// save the list into a json file
func (l *List) Save(filename string) error {
	js, err := json.Marshal(l)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, js, 0644)
}

// Get method opens the provided file name,
// Decodes the JSON data and parses it into a List
func (l *List) Get(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(file) == 0 {
		return nil
	}
	return json.Unmarshal(file, l)
}

func (l *List) String() string {
	formatted := ""

	for k, t := range *l {
		prefix := "  "
		if t.Done {
			prefix = "X "
		}
		formatted += fmt.Sprintf("%s%d: %s\n", prefix, k+1, t.Task)
	}
	// formatted = formatted[:len(formatted)-1]
	return formatted
}

func (l *List) PrintFlexible(verbose, hideCompleted bool) string {
	formatted := ""
	for k, t := range *l {
		prefix := "  "
		if t.Done {
			if hideCompleted {
				continue
			} else {
				prefix = "X "
			}
		}
		if verbose {

			formatted += fmt.Sprintf("%s%d: %s  %s  %s\n", prefix, k+1, t.Task, t.CreatedAt.Format("01-02-2006 15:04:05"), t.CompletedAt.Format("01-02-2006 15:04:05"))
		} else {
			formatted += fmt.Sprintf("%s%d: %s\n", prefix, k+1, t.Task)
		}
	}
	// formatted = formatted[:len(formatted)-1]
	return formatted
}
