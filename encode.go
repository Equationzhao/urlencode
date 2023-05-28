package urlencode

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"
)

const bufferSize = 1024

const (
	StrBuilderPoolEnable = true
)

var StrBuilderPool = sync.Pool{
	New: func() any {
		a := strings.Builder{}
		a.Grow(bufferSize)
		return &a
	},
}

// Urlencoded Convert any type to x-www-form-urlencoded format
type Urlencoded interface {
	Convert2Urlencoded() string
}

// Convert2Urlencoded Convert to x-www-form-urlencoded format
// if the type implements Urlencoded, use its method
// else use reflection
// if the type has a field named "urlencoded", use it as key
// else if the type has a field named "json", use it as key
// else use the field name as key
//
// // example:
//
//		type A struct {
//			Device     string `urlencoded:"device" json:"device"`
//			IP         string `json:"ip"`
//			Type       string
//			NotEmpty   string `urlencoded:"notempty,omitempty"`
//			Empty0     string `urlencoded:"empty0,omitempty"`
//			Empty1     string `urlencoded:",omitempty"`
//			unexported string
//		}
//		a := A{Device: "device", IP: "ip", Type: "type", NotEmpty: "notempty"}
//		fmt.Println(Convert2Urlencoded(a))
//
//			output:device=device&ip=ip&Type=type&notempty=notempty
//
//		type B struct {
//		        X string
//		        x string
//		}
//
//		ab:=struct {
//			A
//			B
//		}{A: a, B: B{X: "123", x: "321"}}
//
//		fmt.Println(Convert2Urlencoded(ab))
//
//
//	 	output:device=device&ip=ip&Type=type&notempty=notempty&X=123
//
//		m := map[string]string{"device": "device", "ip": "ip", "Type": "type"}
//		fmt.Println(Convert2Urlencoded(m))
//		output:
//		device=device&ip=ip&Type=type
//
//		s := []string{"device", "ip", "type"}
//		fmt.Println(Convert2Urlencoded(s))
//
//	   	output:=device&=ip&=type
//
// ! [need test]
func Convert2Urlencoded(i any) string {
	return convert2urlencoded(i, true)
}

func convert2urlencoded(i any, isTheLast bool) string {
	if i == nil {
		return ""
	}

	if c, ok := i.(Urlencoded); ok {
		return c.Convert2Urlencoded()
	}

	v := reflect.ValueOf(i)
	t := reflect.TypeOf(i)

	var content *strings.Builder
	if StrBuilderPoolEnable {
		content = StrBuilderPool.Get().(*strings.Builder)
		defer func() {
			content.Reset()
			StrBuilderPool.Put(content)
		}()
	} else {
		content = new(strings.Builder)
	}

	// recursively deference pointer
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			break
		}
		v = v.Elem()
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		if isTheLast {
			// the last element
			content.WriteByte('=')
			content.WriteString(url.QueryEscape(v.String()))
			return content.String()
		} else {
			// not the last element
			content.WriteByte('=')
			content.WriteString(url.QueryEscape(v.String()))
			content.WriteByte('&')
			return content.String()
		}
	case reflect.Map:
		iter := v.MapRange()
		l := v.Len()
		for iter.Next() {
			l--
			v := iter.Value()
			t := v.Kind()
			if t == reflect.Pointer {
				v = v.Elem()
				t = iter.Value().Elem().Kind()
			}
			// if v is not struct/map/pointer
			switch t {
			case reflect.Struct:
				fallthrough
			case reflect.Map:
				if l > 0 {
					// not the last element
					content.WriteString(convert2urlencoded(iter.Value(), false))
				} else {
					content.WriteString(convert2urlencoded(iter.Value(), true))
				}
			case reflect.Interface:
				key := iter.Key()
				switch v.Interface().(type) {
				// basic type
				case string, int, int8, int16, int32, int64,
					uint, uint8, uint16, uint32, uint64,
					float32, float64,
					complex64, complex128,
					bool, []byte, []rune, uintptr:
					if l > 0 {
						// not the last element
						// content.WriteString(fmt.Sprintf("%s=%s&", key, url.QueryEscape(fmt.Sprint(v))))
						content.WriteString(key.String())
						content.WriteByte('=')
						content.WriteString(url.QueryEscape(fmt.Sprint(v)))
						content.WriteByte('&')
					} else {
						// content.WriteString(fmt.Sprintf("%s=%s", key, url.QueryEscape(fmt.Sprint(v))))
						content.WriteString(key.String())
						content.WriteByte('=')
						content.WriteString(url.QueryEscape(fmt.Sprint(v)))
					}
				case time.Duration:
					if l > 0 {
						// not the last element
						content.WriteString(fmt.Sprintf("%s=%s&", key, url.QueryEscape(v.Interface().(time.Duration).String())))
					} else {
						content.WriteString(fmt.Sprintf("%s=%s", key, url.QueryEscape(v.Interface().(time.Duration).String())))
					}
				case time.Time:

					if l > 0 {
						// not the last element
						content.WriteString(fmt.Sprintf("%s=%s&", key, url.QueryEscape(v.Interface().(time.Time).Format(time.RFC3339))))
					} else {
						content.WriteString(fmt.Sprintf("%s=%s", key, url.QueryEscape(v.Interface().(time.Time).Format(time.RFC3339))))
					}
				default:
					if l > 0 {
						// not the last element
						content.WriteString(convert2urlencoded(v.Interface(), false))
					} else {
						content.WriteString(convert2urlencoded(v.Interface(), true))
					}
				}
			default:
				k := iter.Key()
				if l > 0 {
					// not the last element
					content.WriteString(fmt.Sprintf("%s=%s&", k, url.QueryEscape(fmt.Sprint(v))))
				} else {
					content.WriteString(fmt.Sprintf("%s=%s", k, url.QueryEscape(fmt.Sprint(v))))
				}
			}
		}
	case reflect.Struct:
		n := t.NumField()
		pieces := make([]string, 0, n)

		for i := 0; i < n; i++ {
			fieldi := t.Field(i)
			if fieldi.PkgPath != "" && !fieldi.Anonymous { // unexported
				continue
			}

			name := fieldi.Tag.Get(urlTag)
			var tags []string = nil
			if name == "" {
				// if there's no "urlencoded" Tag, use "json" instead
				name = fieldi.Tag.Get(jsonTag)

				if name == "-" {
					continue
				}

				// if there's no "json" Tag, use field name instead
				if name == "" {
					name = fieldi.Name
				} else {
					// not empty, split it
					if strings.Contains(name, ",") {
						tags = splitTags(name)
						name = tags[0] // remove ",omitempty,..."
					}
				}
			} else {
				// not empty, split it
				if strings.Contains(name, ",") {
					tags = splitTags(name)
					name = tags[0] // remove ",omitempty,..."
					if name == "" {
						name = fieldi.Name
					}
				}
			}

			fieldType := fieldi.Type
			if fieldType.Kind() == reflect.Struct || fieldType.Kind() == reflect.Map || fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {

				if tt, ok := v.Field(i).Interface().(time.Time); ok {
					// get tag "time_format"
					format := fieldi.Tag.Get(timeFormatTag)
					if format == "" {
						format = time.RFC3339
					}
					pieces = append(pieces, fmt.Sprintf("%s=%s", name, url.QueryEscape(tt.Format(format))))
				} else {
					pieces = append(pieces, convert2urlencoded(v.Field(i).Interface(), true))
				}

				continue
			}

			// if tags contains "omitempty" and the field is empty, skip it
			if !tagsContain(tags, omitemptyTag) || !isEmptyValue(v.Field(i)) {
				var value any

				switch v.Field(i).Interface().(type) {
				// basic type
				case string, int, int8, int16, int32, int64,
					uint, uint8, uint16, uint32, uint64,
					float32, float64,
					complex64, complex128,
					bool, []byte, []rune, uintptr:
					value = v.Field(i).Interface()

				case time.Duration:
					td := v.Field(i).Interface().(time.Duration)
					// get tag "time_duration_format"
					switch fieldi.Tag.Get(timeDurationFormatTag) {
					case "ns":
						value = fmt.Sprintf("%dns", td.Nanoseconds())
					case "us", "μs":
						value = fmt.Sprintf("%dus", td.Microseconds())
					case "ms":
						value = fmt.Sprintf("%dms", td.Milliseconds())
					case "s", "second":
						value = fmt.Sprintf("%fs", td.Seconds())
					case "m", "minute":
						value = fmt.Sprintf("%fm", td.Minutes())
					case "h", "hour":
						value = fmt.Sprintf("%fh", td.Hours())
					case "d", "day":
						value = fmt.Sprintf("%fd", td.Hours()/24)
					case humanReadableTag, "", "human":
						value = td.String()
					default:
						value = td.String()
					}

				default:
					value = v.Field(i).Interface()
				}

				pieces = append(pieces,
					fmt.Sprintf("%s=%s",
						url.QueryEscape(name),
						url.QueryEscape(fmt.Sprintf("%v", value))))
			}
		}

		if t == timeType {
			content.WriteByte('=')
			content.WriteString(url.QueryEscape(v.Interface().(time.Time).Format(time.RFC3339)))
		} else {
			content.WriteString(strings.Join(pieces, "&"))
		}

		if !isTheLast && content.Len() != 0 {
			content.WriteByte('&')
		}

	case reflect.Slice, reflect.Array:
		// if the slice is empty, return ""
		l := v.Len()
		if l == 0 {
			return ""
		}

		for i := 0; i < l-1; i++ {
			content.WriteString(convert2urlencoded(v.Index(i).Interface(), false))
		}
		content.WriteString(convert2urlencoded(v.Index(l-1).Interface(), isTheLast))
	case reflect.Interface:
		if v.IsNil() {
			return ""
		}
		switch v.Interface().(type) {
		// basic type
		case string, int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64,
			float32, float64,
			complex64, complex128,
			bool, []byte, []rune, uintptr:
			// not the last element
			// content.WriteString(fmt.Sprintf("%s=%s&", key, url.QueryEscape(fmt.Sprint(v))))
			if !isTheLast {
				content.WriteByte('=')
				content.WriteString(url.QueryEscape(fmt.Sprint(v)))
				content.WriteByte('&')
			} else {
				content.WriteByte('=')
				content.WriteString(url.QueryEscape(fmt.Sprint(v)))
			}

		case time.Duration:
			if !isTheLast {
				// not the last element
				content.WriteString(fmt.Sprintf("=%s&", url.QueryEscape(v.Interface().(time.Duration).String())))
			} else {
				content.WriteString(fmt.Sprintf("=%s", url.QueryEscape(v.Interface().(time.Duration).String())))
			}
		default:
			if !isTheLast {
				// not the last element
				content.WriteString(fmt.Sprintf("=%s&", url.QueryEscape(fmt.Sprint(v))))
			} else {
				content.WriteString(fmt.Sprintf("=%s", url.QueryEscape(fmt.Sprint(v))))
			}
		}
	default:
		if !isTheLast {
			// not the last element
			content.WriteString(fmt.Sprintf("=%s&", url.QueryEscape(fmt.Sprint(v))))
		} else {
			content.WriteString(fmt.Sprintf("=%s", url.QueryEscape(fmt.Sprint(v))))
		}
	}

	return content.String()
}

// isEmptyValue reports whether v is an empty value.
// https://github.com/google/go-querystring/blob/master/query/encode.go
// https://github.com/google/go-querystring/blob/3455b5313413da74e4b65662d161dde73d961cdd/query/encode.go#L316
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}

	type zeroable interface {
		IsZero() bool
	}

	if z, ok := v.Interface().(zeroable); ok {
		return z.IsZero()
	}

	return false
}

func splitTags(tags string) []string {
	return strings.Split(tags, ",")
}

func tagsContain(tags []string, tag string) (contain bool) {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

var timeType = reflect.TypeOf(time.Time{})

const (
	jsonTag               = "json"
	urlTag                = "urlencoded"
	timeFormatTag         = "time_format"
	omitemptyTag          = "omitempty"
	timeDurationFormatTag = "time_duration_format"
	timeDurationTag       = "time_duration"
	humanReadableTag      = "normal"
	unixTag               = "unix"
	unixmilliTag          = "unixmilli"
	unixnanoTag           = "unixnano"
)

// type Encoder struct {
//
// }
