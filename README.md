# URL encode
> encode struct/map/slice/... to url query string

## Usage

```go
package main

import "fmt"
import "github.com/Equationzhao/urlencode"

func main() {
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
	// output: 
	// =10.001s
	// =2023-05-28T23:06:31+08:00
}

```

