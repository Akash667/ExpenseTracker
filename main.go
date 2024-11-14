package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

var counter = 0
var dbFile *os.File

func init() {
	file, err := os.Open("expenses.json")
	if err != nil {
		fmt.Println("init() unable to read file", err)
	}
	buffer, size := fileToBuffer(*file)

	jsonData := BufferToJson(buffer, size)

	for _, value := range jsonData {
		if value.ExpenseID == counter {
			counter = value.ExpenseID + 1

		}
	}

}

type ExpenseList []Expense
type Expense struct {
	ExpenseID   int    `json:"id"`
	Description string `json:"description"`
	Value       int    `json:"value"`
	DateCreated string `json:"dateCreated"`
}

type DBStruct struct {
	ListOfExpenses []Expense
}

func fileToBuffer(file os.File) ([]byte, int64) {
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()

	buffer := make([]byte, fileSize)
	_, err := file.Read(buffer)
	if err != nil {
		fmt.Println("Reading file to buffer failed", err)
		return []byte{}, 0
	}
	return buffer, fileSize
}

func BufferToJson(buffer []byte, n int64) ExpenseList {

	var jsonData = ExpenseList{}
	if n == 0 {
		return ExpenseList{}
	}

	err := json.Unmarshal(buffer, &jsonData)
	if err != nil {
		fmt.Println("Unmarshaling failed", err)
	}

	return jsonData
}

func main() {
	argsWithoutFile := os.Args[1:]
	file, err := os.OpenFile("expenses.json", os.O_RDWR|os.O_CREATE, 0644)
	dbFile = file
	if err != nil {
		println("unable to open file")
		os.Exit(1)
	}
	defer file.Close()

	buffer, filesize := fileToBuffer(*file)

	jsonData := BufferToJson(buffer, filesize)

	action := argsWithoutFile[0]

	switch action {
	case "add":
		fmt.Println("adding")
		jsonData.addExpense(argsWithoutFile[1:])
	case "list":
		jsonData.listExpenses()
	case "summary":
		jsonData.getSummary(argsWithoutFile[1:])
	case "delete":
		jsonData.deleteExpense(argsWithoutFile[1:])
	default:
		fmt.Println("nothing")
	}
}

func (dataFile ExpenseList) addExpense(args []string) {
	expenseData := Expense{}
	currentTime := time.Now()
	year, month, day := currentTime.Date()
	currentDate := fmt.Sprintf("%d-%d-%d", year, month, day)
	fmt.Println(currentDate)
	// newListOfExpenses := &dataFile.ListOfExpenses
	if args[0] == "--description" {
		expenseData.ExpenseID = counter
		expenseData.Description = args[1]
		expenseData.DateCreated = currentDate
		if args[2] == "--amount" {
			expenseData.Value, _ = strconv.Atoi(args[3])
		}
	}
	dataFile = append(dataFile, expenseData)

	updateDatabase(dataFile)
}

func updateDatabase(dataFile ExpenseList) {
	bufferData, err := json.Marshal(dataFile)
	if err != nil {
		fmt.Println("error while marshalling for adding entry", err)
	}
	dbFile.Truncate(0)
	dbFile.Seek(0, 0)
	dbFile.Write(bufferData)

}
func (dataFile ExpenseList) listExpenses() {
	fmt.Println("id", "Expense", "Value")
	for _, value := range dataFile {
		fmt.Println(value.ExpenseID, value.Description, value.Value)
	}

}

func (dataFile ExpenseList) getSummary(monthInput []string) {

	fmt.Println("size of monthInput ", len(monthInput))
	var dayOfWeek int
	monthGiven := false

	if len(monthInput) != 0 && monthInput[0] == "--month" {
		fmt.Println("monthGiven set to true!")
		monthGiven = true
		if len(monthInput) < 2 {
			fmt.Println("month value not provided")
			os.Exit(1)
		}

	}
	sum := 0

	for _, value := range dataFile {
		var monthInInt int
		if monthGiven {
			currTime, _ := time.Parse(time.DateOnly, value.DateCreated)
			temp := currTime.Month()
			dayOfWeek = int(temp)
			var err error
			monthInInt, err = strconv.Atoi(monthInput[1])
			if err != nil {
				fmt.Println("Invalid month value")
				os.Exit(1)
			}
		}
		if !monthGiven {
			fmt.Println("Adding to sum")
			sum += value.Value
		} else if dayOfWeek == monthInInt {
			sum += value.Value

		}
	}
	if monthGiven {
		fmt.Printf("The summary for given month %d is $%d\n", dayOfWeek, sum)
		return
	}

	fmt.Printf("The sum of the amount is $%d\n", sum)

}

func (dataFile ExpenseList) deleteExpense(idInput []string) {

	if idInput[0] != "--id" {
		fmt.Println("Invalid Delete syntax")
		os.Exit(1)
	}

	expenseId, err := strconv.Atoi(idInput[1])
	if err != nil {
		fmt.Println("Invalid id")
		os.Exit(1)
	}
	for index, value := range dataFile {
		if value.ExpenseID == expenseId {
			dataFile = append(dataFile[:index], dataFile[index+1:]...)
			break
		}
	}

	updateDatabase(dataFile)
	fmt.Printf("Data Updated %+v", dataFile)
}
