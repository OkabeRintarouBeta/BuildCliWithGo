package todo_test

import (
	"reflect"
	"testing"

	"github.com/okaberintaroubeta/todo"
)

func TestAdd(t *testing.T) {
	l := todo.List{}
	taskName := "New Task"
	l.Add(taskName)
	if l[0].Task != taskName {
		t.Errorf("Expected %q, got %q instead.", taskName, l[0].Task)
	}
	if l[0].Done != false {
		t.Errorf("Task shouldn't be done when created")
	}
}

func TestComplete(t *testing.T) {
	l := todo.List{}
	taskName := "New Task"
	l.Add(taskName)
	err := l.Complete(0)
	if err != nil {
		t.Errorf("Got error: %v", err)
	}
	if !l[0].Done {
		t.Errorf("This task should be completed")
	}
}

func TestDelete(t *testing.T) {
	l := todo.List{}
	tasks := []string{
		"New Task 1",
		"New Task 2",
		"New Task 3",
	}
	for _, v := range tasks {
		l.Add(v)
	}
	if l[0].Task != tasks[0] {
		t.Errorf("Expected %q, got %q instead.", tasks[0], l[0].Task)
	}
	l.Delete(2)
	if len(l) != 2 {
		t.Errorf("Expected list length %d, got %d instead", 2, len(l))
	}
	if l[1].Task != tasks[2] {
		t.Errorf("Expected %q, got %q instead.", tasks[2], l[1].Task)
	}
}

func TestSaveGet(t *testing.T) {
	l1 := todo.List{}
	l2 := todo.List{}

	taskName1 := "New Task 1"
	taskName2 := "New Task 2"
	l1.Add(taskName1)
	l1.Add(taskName2)

	filename := "file1"
	if err := l1.Save(filename); err != nil {
		t.Errorf("Error while saving the file: %s", err)
	}
	if err := l2.Get(filename); err != nil {
		t.Errorf("Error while getting the file: %s", err)
	}
	if !reflect.DeepEqual(l1, l2) {
		t.Errorf("File before Save is not the same as file after Get")
	}

}
