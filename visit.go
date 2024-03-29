package configs

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// visitor is a function which acts on struct leaf properties.
type visitor func(environment string, value reflect.Value) *visitError

// visitError is an error which can be returned by Visitors if something
// went wrong while running the function.
type visitError struct {
	error
	// Key describes the leaf node. In general, this can just be the
	// "environment" argument.
	Key string
}

// visit calls the visitor function on each property on container,
// unless that property is a struct itself. It will recurse through any
// any structs until it eventually gets finds the leaves.
func visit(container interface{}, v visitor) error {
	return doVisit("", reflect.ValueOf(container), v, nil)
}

var s struct{}
var terminalTypes = map[string]struct{}{
	"big.Int":  s,
	"*big.Int": s,
}

func doVisit(environmentSoFar string, theValue reflect.Value, v visitor, errs error) error {
	theType := theValue.Type().Elem()

	for i := 0; i < theType.NumField(); i++ {
		thisField := theType.Field(i)
		thisFieldValue := theValue.Elem().Field(i)
		environment := environmentSoFar + "_" + envFromName(thisField.Name)
		switch thisField.Type.Kind() {
		case reflect.Ptr:
			if _, ok := terminalTypes[thisField.Type.String()]; ok {
				if err := v(environment, thisFieldValue); err != nil {
					errs = appendError(errs, err.Key, err)
				}
			} else {
				errs = doVisit(environment, thisFieldValue, v, errs)
			}
		default:
			if err := v(environment, thisFieldValue); err != nil {
				errs = appendError(errs, err.Key, err)
			}
		}
	}
	return errs
}

func envFromName(name string) string {
	var sb strings.Builder

	thisRune, width := utf8.DecodeRuneInString(name[0:])
	sb.WriteRune(unicode.ToUpper(thisRune))
	var rest = name[width:]
	for len(rest) > 0 {
		lastRune := thisRune
		thisRune, width = utf8.DecodeRuneInString(rest[0:])
		if unicode.IsUpper(thisRune) && unicode.IsLower(lastRune) {
			sb.WriteString("_")
		}
		sb.WriteRune(unicode.ToUpper(thisRune))
		rest = rest[width:]
	}
	return sb.String()
}

// traversalError is returned by visit() if the visitor returned any errors
type traversalError struct {
	summary     string
	invalidKeys map[string]error
}

// Ensure adds custom error messagse to the error returned by LoadWithPrefix().
//
// Normal usage looks like this:
//
//	err := configs.LoadWithPrefix(&cfg, "MYAPP")
//	err = configs.Ensure(err, "MYAPP_PORT", cfg.Port > 0, "must be a positive integer")
//	err = configs.Ensure(err, "MYAPP_ENV", isValid(cfg.ENV), "must be one of: %v", validEnvs)
//
// If predicate is true, err is returned unchanged.
// If predicate is false and err is nil, a new error will be returned.
//
// In all cases the returned error will "pretty print" your validation error alongside
// any errors generated by the LoadWithPrefix() call.
func Ensure(err error, key string, predicate bool, msgFormat string, msgArgs ...interface{}) error {
	if predicate {
		return err
	}
	return appendError(err, key, fmt.Errorf(msgFormat, msgArgs...))
}

func appendError(err error, key string, msg error) error {
	if err == nil {
		return &traversalError{
			invalidKeys: map[string]error{
				key: msg,
			},
		}
	}

	if casted, ok := err.(*traversalError); ok {
		// Don't overwrite old error messages. This makes sure that type errors like
		// "must be an int" get printed over post-parse errors like "must be positive"
		if existing, ok := casted.invalidKeys[key]; ok {
			casted.invalidKeys[key] = fmt.Errorf("%v: %v", existing, msg)
		} else {
			casted.invalidKeys[key] = msg
		}
		return casted
	}

	panic("Ensure() only works on errors returend by this library")
}

// Error returns an error message describing all the invalid environment variables.
func (p *traversalError) Error() string {
	if p == nil {
		return ""
	}

	msg := strings.Builder{}
	msg.WriteString("Errors occurred while acting on the struct:\n")
	for env, err := range p.invalidKeys {
		// May be overkill... but playing it a little safe. Someone might mis-type a password,
		// call Ensure() after a failed login, and then this library would print a password
		// that's only off by one character.
		if strings.Contains(strings.ToLower(env), "password") {
			msg.WriteString(fmt.Sprintf("  %s %v\n", env, err))
		} else {
			msg.WriteString(fmt.Sprintf("  %s %v: got \"%s\"\n", env, err, os.Getenv(env)))
		}
	}
	return msg.String()
}
