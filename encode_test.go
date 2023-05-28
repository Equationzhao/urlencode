package urlencode

import (
	"fmt"
	"testing"
	"time"
)

func TestEncode(t *testing.T) {
	type A struct {
		Device       string `urlencoded:"device" json:"device"`
		IP           string `json:"ip"`
		Type         string
		NotEmpty     string        `urlencoded:"notempty,omitempty"`
		Empty0       string        `urlencoded:"empty0,omitempty"`
		Empty1       string        `urlencoded:",omitempty"`
		Timeduration time.Duration `urlencoded:"timeduration,omitempty" time_duration_format:"day"`
		Map          map[string]string
		Time         time.Time `urlencoded:"time,omitempty" time_format:"2006-01-02-15-04-05"`
		unexported   string
	}
	a := A{
		Device: "device",
		IP:     "ip", Type: "type",
		NotEmpty:     "notempty",
		Timeduration: time.Second * 10,
		Empty1:       "empty1",
		Map:          map[string]string{"map1": "map1", "map2": "map2", "map3": "map3"},
		Time:         time.Now(),
	}

	fmt.Println(Convert2Urlencoded(a))

	type B struct {
		X string
		x string
	}

	ab := struct {
		A
		B
	}{A: a, B: B{X: "123", x: "321"}}

	fmt.Println(Convert2Urlencoded(ab))

	m := map[string]string{"device": "device", "ip": "ip", "Type": "type"}
	fmt.Println(Convert2Urlencoded(m))

	s := []string{"device", "ip", "type"}
	fmt.Println(Convert2Urlencoded(s))

	c := time.Second * 10
	fmt.Println(Convert2Urlencoded(c))
}

func TestExample(t *testing.T) {
	// encode struct
	type User struct {
		Name string    `urlencoded:"name,omitempty"`
		Age  int       `urlencoded:"age,omitempty"`
		Born time.Time `urlencoded:"born,omitempty" time_format:"20060102"`
	}
	user := User{
		Name: "equation",
		Age:  18,
		Born: time.Date(2002, 5, 31, 0, 0, 0, 0, time.Local),
	}
	fmt.Println(Convert2Urlencoded(user))
	// output: name=equation&age=18&born=20020531

	// encode map
	m := map[string]any{
		"name": "equation",
		"age":  18,
		"born": time.Date(2002, 5, 31, 0, 0, 0, 0, time.Local),
	}
	fmt.Println(Convert2Urlencoded(m))
	// output: born=2002-05-31T00%3A00%3A00%2B08%3A00&name=equation&age=18 // random order

	// encode slice
	s := []any{
		"equation",
		18,
		time.Date(2002, 5, 31, 0, 0, 0, 0, time.Local),
	}
	fmt.Println(Convert2Urlencoded(s))
	// output: =equation&=18&=2002-05-31T00%3A00%3A00%2B08%3A00

	// encode array
	a := [10]any{
		"equation",
		"18",
		time.Date(2002, 5, 31, 0, 0, 0, 0, time.Local),
	}
	fmt.Println(Convert2Urlencoded(a))
	// output: =equation&=18&=2002-05-31T00%3A00%3A00%2B08%3A00&

	// encode time.Duration time.Time
	td := 10*time.Second + 1*time.Millisecond
	fmt.Println(Convert2Urlencoded(td))
	tt := time.Now()
	fmt.Println(Convert2Urlencoded(tt))
	// output: =10.001s
	// 		   =2023-05-28T23:06:31+08:00
}
