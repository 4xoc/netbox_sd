// MIT License
//
// Copyright (c) 2024 WIIT AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
// documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the
// Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
// WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR
// OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package netbox

import "log"

// Logger implements log related methods to forward messages to the appropriate recipient.
type Logger interface {
	Infof(format string, val ...interface{})
	Errorf(format string, val ...interface{})
	Debugf(format string, val ...interface{})
	Tracef(format string, val ...interface{})
}

// HTTPTracing enables or disables HTTP tracing. When enabled, the Logger's Tracef function is called and contains all
// HTTP request and response headers and payload as well as timing information. Use with care, this will expose secrets
// in plain text and affects performance.
func (client *Client) HTTPTracing(val bool) {
	client.httpTracing = val
}

// SetLogger updates the Logger interface used by this Client for sending log messages. NOTE: there is no check in place
// that validates the correctness of the interface. Logging messages might cause panics when a specific interface method
// is not defined. Make sure all methods exist for an instance of Logger.
func (client *Client) SetLogger(logger Logger) {
	client.log = logger
}

// defaultLogger is the default implementation of the Logger interface used by this package. It only logs Info and Error
// messages.
//
// NOTE: Be aware that this package updates the log.SetFlags globally.
type defaultLogger int

// Infof is the default implementation of the Logger interface.
func (logger defaultLogger) Infof(format string, val ...interface{}) {
	log.Printf("[netbox-go] "+format, val...)
}

// Errorf is the default implementation of the Logger interface.
func (logger defaultLogger) Errorf(format string, val ...interface{}) {
	log.Printf("[netbox-go] "+format, val...)
}

// Debugf is the default implementation of the Logger interface.
func (logger defaultLogger) Debugf(format string, val ...interface{}) {
	log.Printf("[netbox-go] "+format, val...)
}

// Tracef is the default implementation of the Logger interface.
func (logger defaultLogger) Tracef(format string, val ...interface{}) {
	log.Printf("[netbox-go] "+format, val...)
}
