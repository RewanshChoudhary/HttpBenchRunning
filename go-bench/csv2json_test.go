package main

import (
	"flag"
	"fmt"
	
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_getFileData(t *testing.T) {
	tests := []struct {
		name    string
		want    inputFile
		wantErr bool
		osArgs  []string
	}{
		{"Default parameters", inputFile{"test.csv", "comma", false}, false, []string{"zsh", "test.csv"}},
		{"No Parameter", inputFile{}, true, []string{"zsh"}},
		{"Semicolon", inputFile{"test.csv", "semicolon", false}, false, []string{"zsh", "--separator=semicolon", "test.csv"}},
		{"Pretty Enabled", inputFile{"test.csv", "comma", true}, false, []string{"zsh", "--pretty", "test.csv"}},
		{"Pretty and Semicolon enabled", inputFile{"test.csv", "semicolon", true}, false, []string{"zsh", "--pretty", "--separator=semicolon", "test.csv"}},
		{"Invalid Separator", inputFile{}, true, []string{"zsh", "--separator=pipe"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualArgs := os.Args
			defer func() {
				os.Args = actualArgs
				flag.CommandLine = flag.NewFlagSet(tt.osArgs[0], flag.ContinueOnError)
			}()

			os.Args = tt.osArgs
			got, err := getFileData()

			if (err != nil) != tt.wantErr {
				t.Errorf("getFileData() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFileData() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_checkValidFile(t *testing.T) {
	validFile, err := os.CreateTemp("", "*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(validFile.Name())

	tests := []struct {
		name     string
		fileName string
		want     bool
		wantErr  bool
	}{
		{"File does not exist", "test.csv", false, true},
		{"File does exist", validFile.Name(), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkValidFile(tt.fileName)

			if (err != nil) != tt.wantErr {
				t.Errorf("checkValidFile() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("checkValidFile() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func Test_processCSVFile(t *testing.T) {
	expectedRecords := []map[string]string{
		{"COL1": "1", "COL2": "2", "COL3": "3"},
		{"COL1": "4", "COL2": "5", "COL3": "6"},
	}

	tests := []struct {
		name      string
		csvString string
		separator string
	}{
		{"Comma separated", "COL1,COL2,COL3\n1,2,3\n4,5,6\n", "comma"},
		{"Semicolon separated", "COL1;COL2;COL3\n1;2;3\n4;5;6\n", "semicolon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test.csv")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tt.csvString)
			if err != nil {
				t.Fatal(err)
			}
			tmpFile.Close()

			testInput := inputFile{
				filepath:  tmpFile.Name(),
				pretty:    false,
				separator: tt.separator,
			}

			recordChan := make(chan map[string]string)
			go processCSVFile(testInput, recordChan)

			for _, expected := range expectedRecords {
				got := <-recordChan
				if !reflect.DeepEqual(got, expected) {
					t.Errorf("Expected: %v, Got: %v", expected, got)
				}
			}
		})
	}
}


func Test_getJson (t *testing.T){
	dataMap := []map[string]string{
		{"COL1": "1", "COL2": "2", "COL3": "3"},
		{"COL1": "4", "COL2": "5", "COL3": "6"},
	}
	// Defining our test cases
	tests := []struct {
		csvPath  string // The "fake" csv path.
		jsonPath string // The existing JSON file with the expected data
		pretty   bool // Whether the output is formatted or not
		name     string // The name of the test
	}{
		{"compact.csv", "compact.json", false, "Compact JSON"}, 
		{"pretty.csv", "pretty.json", true, "Pretty JSON"},
	}

	for _,tt:= range tests {
		t.Run (tt.name,func(t *testing.T){
            writerChannel:=make (chan map[string ]string )
			done :=make (chan bool)

			go func (){
				for _,record:=range dataMap{
					writerChannel<-record
				}
				close(writerChannel)

			}()
			//Sending the data in csvPath to json path 
			go writeJsonFile(tt.csvPath,writerChannel,done,tt.pretty)

				<-done


			testOutput,err:=os.ReadFile(tt.jsonPath)

			if (err!=nil){
				fmt.Errorf("The error we get the from the writeJsonPath():= %v",err)

			}
			defer os.Remove(tt.jsonPath)

			wantOutput,err:=os.ReadFile(filepath.Join("testJsonFiles",tt.jsonPath))



			if (string (testOutput)!=string(wantOutput)){
				fmt.Errorf("writeJson():= threw us %v but we want %v",testOutput,wantOutput)

			}







			

		})

	}
}