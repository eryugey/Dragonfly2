// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: pkg/rpc/dfproxy/dfproxy.proto

package dfproxy

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
)

// Validate checks the field values on DfDaemonReq with the rules defined in
// the proto definition for this message. If any rules are violated, an error
// is returned.
func (m *DfDaemonReq) Validate() error {
	if m == nil {
		return nil
	}

	if v, ok := interface{}(m.GetStatTask()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DfDaemonReqValidationError{
				field:  "StatTask",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetImportTask()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DfDaemonReqValidationError{
				field:  "ImportTask",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetExportTask()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DfDaemonReqValidationError{
				field:  "ExportTask",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if v, ok := interface{}(m.GetDeleteTask()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DfDaemonReqValidationError{
				field:  "DeleteTask",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// DfDaemonReqValidationError is the validation error returned by
// DfDaemonReq.Validate if the designated constraints aren't met.
type DfDaemonReqValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DfDaemonReqValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DfDaemonReqValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DfDaemonReqValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DfDaemonReqValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DfDaemonReqValidationError) ErrorName() string { return "DfDaemonReqValidationError" }

// Error satisfies the builtin error interface
func (e DfDaemonReqValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDfDaemonReq.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DfDaemonReqValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DfDaemonReqValidationError{}

// Validate checks the field values on DaemonProxyClientPacket with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *DaemonProxyClientPacket) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Type

	if v, ok := interface{}(m.GetError()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DaemonProxyClientPacketValidationError{
				field:  "Error",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// DaemonProxyClientPacketValidationError is the validation error returned by
// DaemonProxyClientPacket.Validate if the designated constraints aren't met.
type DaemonProxyClientPacketValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DaemonProxyClientPacketValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DaemonProxyClientPacketValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DaemonProxyClientPacketValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DaemonProxyClientPacketValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DaemonProxyClientPacketValidationError) ErrorName() string {
	return "DaemonProxyClientPacketValidationError"
}

// Error satisfies the builtin error interface
func (e DaemonProxyClientPacketValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDaemonProxyClientPacket.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DaemonProxyClientPacketValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DaemonProxyClientPacketValidationError{}

// Validate checks the field values on DaemonProxyServerPacket with the rules
// defined in the proto definition for this message. If any rules are
// violated, an error is returned.
func (m *DaemonProxyServerPacket) Validate() error {
	if m == nil {
		return nil
	}

	// no validation rules for Type

	if v, ok := interface{}(m.GetDaemonReq()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return DaemonProxyServerPacketValidationError{
				field:  "DaemonReq",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	return nil
}

// DaemonProxyServerPacketValidationError is the validation error returned by
// DaemonProxyServerPacket.Validate if the designated constraints aren't met.
type DaemonProxyServerPacketValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e DaemonProxyServerPacketValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e DaemonProxyServerPacketValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e DaemonProxyServerPacketValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e DaemonProxyServerPacketValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e DaemonProxyServerPacketValidationError) ErrorName() string {
	return "DaemonProxyServerPacketValidationError"
}

// Error satisfies the builtin error interface
func (e DaemonProxyServerPacketValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sDaemonProxyServerPacket.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = DaemonProxyServerPacketValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = DaemonProxyServerPacketValidationError{}
