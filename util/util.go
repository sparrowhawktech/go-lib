package util

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"time"
	"unsafe"
)

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ProcessPanic(intf interface{}) {
	Logger("error").Printf("%v", intf)
	stackTrace := string(debug.Stack())
	Logger("error").Println(stackTrace)
}

func Unmarshal(bytes []byte, object interface{}) {
	err := json.Unmarshal(bytes, object)
	if err != nil {
		panic(err)
	}
}

func Marshal(object interface{}) []byte {
	bytes, err := json.Marshal(object)
	if err != nil {
		panic(err)
	}
	return bytes
}

func ScanAll(rows *sql.Rows) []interface{} {
	result := make([]interface{}, 0)
	columns, err := rows.Columns()
	CheckErr(err)
	n := len(columns)
	r := 0
	references := make([]interface{}, n)
	pointers := make([]interface{}, n)
	for i := range references {
		pointers[i] = &references[i]
	}
	for rows.Next() {
		CheckErr(rows.Scan(pointers...))
		values := make([]interface{}, n)
		for i := range pointers {
			values[i] = *pointers[i].(*interface{})
		}
		result = append(result, values)
		r = r + 1
	}
	return result
}

func TruncDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func DateMoveToPrev(date time.Time, day time.Weekday) time.Time {
	for date.Weekday() != day {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

func DateMoveToNext(date time.Time, day time.Weekday) time.Time {
	for date.Weekday() != day {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func EndOfDay(date time.Time) time.Time {
	y, m, d := date.Date()
	return time.Date(y, m, d, 23, 59, 59, 0, date.Location())
}

func Today() time.Time {
	return TruncDate(time.Now())
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		panic(err)
	}
}

func LoadConfig(path string, config interface{}) {
	abs, err := filepath.Abs(path)
	CheckErr(err)
	bytes, err := ioutil.ReadFile(abs)
	CheckErr(err)
	err = json.Unmarshal(bytes, config)
	CheckErr(err)
}

func ParseInt(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	CheckErr(err)
	return i
}

func JsonDecode(i interface{}, r io.Reader) interface{} {
	err := json.NewDecoder(r).Decode(i)
	CheckErr(err)
	return i
}

func JsonEncode(i interface{}, w io.Writer) {
	err := json.NewEncoder(w).Encode(i)
	CheckErr(err)
}

func JsonPretty(i interface{}, w io.Writer) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(i)
	CheckErr(err)
}

func Float64bits(f float64) uint64     { return *(*uint64)(unsafe.Pointer(&f)) }
func Float64frombits(b uint64) float64 { return *(*float64)(unsafe.Pointer(&b)) }
func Round(x float64) float64 {
	const (
		uvone    = 0x3FF0000000000000
		mask     = 0x7FF
		shift    = 64 - 11 - 1
		bias     = 1023
		signMask = 1 << 63
		fracMask = 1<<shift - 1
	)

	bits := Float64bits(x)
	e := uint(bits>>shift) & mask
	if e < bias {
		bits &= signMask
		if e == bias-1 {
			bits |= uvone
		}
	} else if e < bias+shift {
		const half = 1 << (shift - 1)
		e -= bias
		bits += half >> e
		bits &^= fracMask >> e
	}
	return Float64frombits(bits)
}

func RoundTo2Dec(value float32) float32 {
	value64 := float64(value)
	return float32(Round(value64*100) / 100)
}

func PString(s string) *string {
	return &s
}

func PStringf(s string, values ...interface{}) *string {
	return PString(fmt.Sprintf(s, values...))
}

func PInt64(i int64) *int64 {
	return &i
}

func PInt(i int) *int {
	return &i
}

func PFloat32(f float32) *float32 {
	return &f
}

func PFloat64(f float64) *float64 {
	return &f
}

func PTime(t time.Time) *time.Time {
	return &t
}

func PBool(b bool) *bool {
	return &b
}

func PJson(b []byte) *json.RawMessage {
	rm := json.RawMessage(b)
	return &rm
}
