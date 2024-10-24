package orm

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type Date struct {
	Date Time `json:"date"`
}

func TestUnmarshalJSON(t *testing.T) {
	{
		b, _ := json.Marshal(map[string]any{"date": 0})
		fmt.Println(string(b))
		var date Date
		if err := json.Unmarshal(b, &date); err != nil {
			t.Fatal(err)
		}
		fmt.Println(date)
	}
	{
		b, _ := json.Marshal(map[string]any{"date": nil})
		fmt.Println(string(b))
		var date Date
		if err := json.Unmarshal(b, &date); err != nil {
			t.Fatal(err)
		}
		fmt.Println(date)
	}
	{
		b, _ := json.Marshal(map[string]any{"date": time.Now().Unix()})
		fmt.Println(string(b))
		var date Date
		if err := json.Unmarshal(b, &date); err != nil {
			t.Fatal(err)
		}
		fmt.Println(date)
	}
	{
		b, _ := json.Marshal(map[string]any{"date": time.Now().UnixMilli()})
		fmt.Println(string(b))

		var date Date
		if err := json.Unmarshal(b, &date); err != nil {
			t.Fatal(err)
		}
		fmt.Println(date)
	}
	{
		b, _ := json.Marshal(map[string]any{"date": time.Now().Format(time.DateTime)})
		fmt.Println(string(b))

		var date Date
		if err := json.Unmarshal(b, &date); err != nil {
			t.Fatal(err)
		}
		fmt.Println(date)
	}
}
