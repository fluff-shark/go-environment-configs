package configs_test

import (
	"bytes"
	"log"
	"math/big"
	"os"
	"strings"
	"testing"

	configs "github.com/wikisophia/go-environment-configs"
)

type Config struct {
	Boolean      bool
	Int          int
	UINT_8       uint8
	UINT_16      uint16
	UINT_32      uint32
	UINT_64      uint64
	BigInt       big.Int
	String       string
	IntSlice     []int
	StringSlice  []string
	Nested       *Nested
	SomePassword string

	Ά本語 string
}

type Nested struct {
	Value         int
	BigIntPointer *big.Int
}

func TestWellFormedValues(t *testing.T) {
	defer setEnv(t, "MY_BOOLEAN", "true")()
	defer setEnv(t, "MY_INT", "10")()
	defer setEnv(t, "MY_UINT_8", "6")()
	defer setEnv(t, "MY_UINT_16", "7")()
	defer setEnv(t, "MY_UINT_32", "8")()
	defer setEnv(t, "MY_UINT_64", "9")()
	defer setEnv(t, "MY_BIG_INT", "9571")()
	defer setEnv(t, "MY_STRING", "someString")()
	defer setEnv(t, "MY_INT_SLICE", "1,-2")()
	defer setEnv(t, "MY_STRING_SLICE", "abc,def")()
	defer setEnv(t, "MY_NESTED_VALUE", "20")()
	defer setEnv(t, "MY_NESTED_BIG_INT_POINTER", "112")()
	defer setEnv(t, "MY_SOME_PASSWORD", "secret")()
	defer setEnv(t, "MY_Ά本語", "日本語")()

	cfg := Config{
		Nested: &Nested{},
	}
	if err := configs.LoadWithPrefix(&cfg, "MY"); err != nil {
		t.Errorf("Got unexpected Load() error: %v", err)
		return
	}
	assertBoolsEqual(t, true, cfg.Boolean)
	assertStringsEqual(t, "someString", cfg.String)
	assertIntsEqual(t, 10, cfg.Int)
	assertIntsEqual(t, 6, int(cfg.UINT_8))
	assertIntsEqual(t, 7, int(cfg.UINT_16))
	assertIntsEqual(t, 8, int(cfg.UINT_32))
	assertIntsEqual(t, 9, int(cfg.UINT_64))
	assertBigIntsEqual(t, big.NewInt(9571), &cfg.BigInt)
	assertIntSlicesEqual(t, []int{1, -2}, cfg.IntSlice)
	assertStringSlicesEqual(t, []string{"abc", "def"}, cfg.StringSlice)
	assertIntsEqual(t, 20, cfg.Nested.Value)
	assertBigIntsEqual(t, big.NewInt(112), cfg.Nested.BigIntPointer)
	assertStringsEqual(t, "日本語", cfg.Ά本語)

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	configs.LogWithPrefix(&cfg, "MY")
	logged := buf.String()
	assertStringContains(t, logged, "MY_BOOLEAN: true")
	assertStringContains(t, logged, "MY_INT: 10")
	assertStringContains(t, logged, "MY_UINT_8: 6")
	assertStringContains(t, logged, "MY_UINT_16: 7")
	assertStringContains(t, logged, "MY_UINT_32: 8")
	assertStringContains(t, logged, "MY_UINT_64: 9")
	assertStringContains(t, logged, "MY_BIG_INT: big.Int{neg:false, abs:big.nat{0x2563}}")
	assertStringContains(t, logged, "MY_STRING: \"someString\"")
	assertStringContains(t, logged, "MY_INT_SLICE: []int{1, -2}")
	assertStringContains(t, logged, "MY_STRING_SLICE: []string{\"abc\", \"def\"}")
	assertStringContains(t, logged, "MY_NESTED_VALUE: 20")
	assertStringContains(t, logged, "MY_NESTED_BIG_INT_POINTER: 112")
	assertStringContains(t, logged, "MY_SOME_PASSWORD: <redacted>")
	assertNotStringContains(t, logged, "secret")
	assertStringContains(t, logged, "MY_Ά本語: \"日本語\"")
}

func TestBadValues(t *testing.T) {
	defer setEnv(t, "MY_UINT_8", "a")()
	defer setEnv(t, "MY_UINT_16", "b")()
	defer setEnv(t, "MY_UINT_32", "c")()
	defer setEnv(t, "MY_UINT_64", "d")()
	defer setEnv(t, "MY_INT", "foo")()
	defer setEnv(t, "MY_BIG_INT", "99abc")()
	defer setEnv(t, "MY_NESTED_BIG_INT_POINTER", "a34k")()
	defer setEnv(t, "MY_BOOLEAN", "3")()
	defer setEnv(t, "MY_INT_SLICE", "1,foo,2")()
	defer setEnv(t, "MY_NESTED_VALUE", "bar")()
	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.LoadWithPrefix(&cfg, "MY")
	if err == nil {
		t.Errorf("Missing expected Load() error: %v", err)
		return
	}

	msg := err.Error()
	assertStringContains(t, msg, `MY_BOOLEAN must be "true" or "false": got "3"`)
	assertStringContains(t, msg, `MY_UINT_8 must be a uint8: got "a"`)
	assertStringContains(t, msg, `MY_UINT_16 must be a uint16: got "b"`)
	assertStringContains(t, msg, `MY_UINT_32 must be a uint32: got "c"`)
	assertStringContains(t, msg, `MY_UINT_64 must be a uint64: got "d"`)
	assertStringContains(t, msg, `MY_INT must be an int: got "foo"`)
	assertStringContains(t, msg, `MY_BIG_INT must be a base-10 big.Int: got "99abc"`)
	assertStringContains(t, msg, `MY_NESTED_BIG_INT_POINTER must be a base-10 big.Int: got "a34k"`)
	assertStringContains(t, msg, `MY_INT_SLICE must be a comma-separated list of ints: index 1 is invalid: got "1,foo,2"`)
	assertStringContains(t, msg, `MY_NESTED_VALUE must be an int: got "bar"`)
}

func TestUndeflowingInts(t *testing.T) {
	defer setEnv(t, "MY_UINT_8", "-1")()
	defer setEnv(t, "MY_UINT_16", "-1")()
	defer setEnv(t, "MY_UINT_32", "-1")()
	defer setEnv(t, "MY_UINT_64", "-1")()
	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.LoadWithPrefix(&cfg, "MY")
	if err == nil {
		t.Errorf("Missing expected Load() error: %v", err)
		return
	}
	msg := err.Error()
	assertStringContains(t, msg, `MY_UINT_8 has a min value of 0: got "-1"`)
	assertStringContains(t, msg, `MY_UINT_16 has a min value of 0: got "-1"`)
	assertStringContains(t, msg, `MY_UINT_32 has a min value of 0: got "-1"`)
	assertStringContains(t, msg, `MY_UINT_64 has a min value of 0: got "-1"`)
}

func TestOverflowingInts(t *testing.T) {
	defer setEnv(t, "MY_UINT_8", "256")()
	defer setEnv(t, "MY_UINT_16", "65536")()
	defer setEnv(t, "MY_UINT_32", "4294967296")()
	defer setEnv(t, "MY_UINT_64", "18446744073709551616")()
	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.LoadWithPrefix(&cfg, "MY")
	if err == nil {
		t.Errorf("Missing expected Load() error: %v", err)
		return
	}
	msg := err.Error()
	assertStringContains(t, msg, `MY_UINT_8 has a max value of 255: got "256"`)
	assertStringContains(t, msg, `MY_UINT_16 has a max value of 65535: got "65536"`)
	assertStringContains(t, msg, `MY_UINT_32 has a max value of 4294967295: got "4294967296"`)
	assertStringContains(t, msg, `MY_UINT_64 has a max value of 18446744073709551615: got "18446744073709551616"`)
}

func TestMultipleErrors(t *testing.T) {
	defer setEnv(t, "MY_INT", "foo")()
	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.LoadWithPrefix(&cfg, "MY")
	err = configs.Ensure(err, "MY_INT", false, "must be %s", "positive")
	if err == nil {
		t.Errorf("Missing expected Load() error: %v", err)
		return
	}
	assertStringContains(t, err.Error(), `MY_INT must be an int: must be positive: got "foo"`)
}

func TestPanics(t *testing.T) {
	recovered := false
	defer func() {
		if r := recover(); r != nil {
			recovered = true
		}
		if !recovered {
			t.Error("MustLoadWithPrefix should panic on invalid inputs, but didn't")
		}
	}()
	defer setEnv(t, "MY_INT", "foo")()
	cfg := Config{
		Nested: &Nested{},
	}
	configs.MustLoadWithPrefix(&cfg, "MY")
}

func TestExtraErrors(t *testing.T) {
	defer setEnv(t, "MY_INT", "-1")()

	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.LoadWithPrefix(&cfg, "MY")
	if err != nil {
		t.Errorf("Got unexpected Load() error: %v", err)
		return
	}
	err = configs.Ensure(err, "MY_INT", true, "must be a negative integer")
	if err != nil {
		t.Errorf("Ensure() shouldn't produce an error if the predicate is true. Got: %v", err)
		return
	}

	err = configs.Ensure(err, "MY_INT", false, "must be a positive integer")
	if err == nil {
		t.Error("Ensure() should have returned a real error")
		return
	}
	assertStringContains(t, err.Error(), `MY_INT must be a positive integer: got "-1"`)
}

func TestPasswordPrinting(t *testing.T) {
	defer setEnv(t, "MY_SOME_PASSWORD", "secret")()
	cfg := Config{
		Nested: &Nested{},
	}
	err := configs.LoadWithPrefix(&cfg, "MY")
	err = configs.Ensure(err, "MY_SOME_PASSWORD", false, "is invalid")
	assertStringContains(t, err.Error(), `MY_SOME_PASSWORD is invalid`)
	assertNotStringContains(t, err.Error(), "secret")
}

func assertStringsEqual(t *testing.T, expected string, actual string) {
	t.Helper()
	if expected != actual {
		t.Errorf(`Expected "%s" does not match actual "%s"`, expected, actual)
	}
}

func assertStringContains(t *testing.T, whole string, fragment string) {
	t.Helper()
	if !strings.Contains(whole, fragment) {
		t.Errorf(`Expected "%s" to contain fragment "%s"`, whole, fragment)
	}
}

func assertNotStringContains(t *testing.T, whole string, fragment string) {
	t.Helper()
	if strings.Contains(whole, fragment) {
		t.Errorf(`Expected "%s" NOT to contain fragment "%s"`, whole, fragment)
	}
}

func assertStringSlicesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf(`Expected "%v" does not match actual "%v". The number of elements differ`, expected, actual)
		return
	}
	for i := 0; i < len(expected); i++ {
		assertStringsEqual(t, expected[i], actual[i])
	}
}

func assertIntSlicesEqual(t *testing.T, expected []int, actual []int) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Errorf(`Expected "%v" does not match actual "%v". The number of elements differ`, expected, actual)
		return
	}
	for i := 0; i < len(expected); i++ {
		assertIntsEqual(t, expected[i], actual[i])
	}
}

func assertIntsEqual(t *testing.T, expected int, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf(`Expected "%d" does not match actual "%d"`, expected, actual)
	}
}

func assertBigIntsEqual(t *testing.T, expected *big.Int, actual *big.Int) {
	t.Helper()
	if expected.Cmp(actual) != 0 {
		t.Errorf(`Expected "%s" does not match actual "%s"`, expected.String(), actual.String())
	}
}

func assertBoolsEqual(t *testing.T, expected bool, actual bool) {
	t.Helper()
	if expected != actual {
		t.Errorf(`Expected "%t" does not match actual "%t"`, expected, actual)
	}
}

// setEnv acts as a wrapper around os.Setenv, returning a function that resets the environment
// back to its original value. This prevents tests from setting environment variables as a side-effect.
func setEnv(t *testing.T, key string, val string) func() {
	t.Helper()
	orig, set := os.LookupEnv(key)
	if err := os.Setenv(key, val); err != nil {
		t.Errorf("Error setting environment value %s: %v", key, err)
		return func() {}
	}

	if set {
		return func() {
			os.Setenv(key, orig)
		}
	} else {
		return func() {
			os.Unsetenv(key)
		}
	}
}
