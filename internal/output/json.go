package output

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintJSON(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		PrintErrorJSON(err.Error())
		return
	}
	fmt.Println(string(data))
}

func PrintErrorJSON(msg string) {
	data, _ := json.MarshalIndent(map[string]string{"error": msg}, "", "  ")
	fmt.Fprintln(os.Stderr, string(data))
}
