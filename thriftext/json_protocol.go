/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements. See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership. The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License. You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package thriftext

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	. "github.com/DrWrong/monica/thrift"
)

type _ParseContext int

const (
	_CONTEXT_IN_TOPLEVEL          _ParseContext = 1
	_CONTEXT_IN_LIST_FIRST        _ParseContext = 2
	_CONTEXT_IN_LIST              _ParseContext = 3
	_CONTEXT_IN_OBJECT_FIRST      _ParseContext = 4
	_CONTEXT_IN_OBJECT_NEXT_KEY   _ParseContext = 5
	_CONTEXT_IN_OBJECT_NEXT_VALUE _ParseContext = 6
)

func (p _ParseContext) String() string {
	switch p {
	case _CONTEXT_IN_TOPLEVEL:
		return "TOPLEVEL"
	case _CONTEXT_IN_LIST_FIRST:
		return "LIST-FIRST"
	case _CONTEXT_IN_LIST:
		return "LIST"
	case _CONTEXT_IN_OBJECT_FIRST:
		return "OBJECT-FIRST"
	case _CONTEXT_IN_OBJECT_NEXT_KEY:
		return "OBJECT-NEXT-KEY"
	case _CONTEXT_IN_OBJECT_NEXT_VALUE:
		return "OBJECT-NEXT-VALUE"
	}
	return "UNKNOWN-PARSE-CONTEXT"
}

// JSON protocol implementation for thrift.
//
// This protocol produces/consumes a simple output format
// suitable for parsing by scripting languages.  It should not be
// confused with the full-featured TJSONProtocol.
//
type JSONProtocol struct {
	trans TTransport

	parseContextStack []int
	dumpContext       []int

	writer *bufio.Writer
}

// Constructor
func NewJSONProtocol(t TTransport) *JSONProtocol {
	v := &JSONProtocol{trans: t,
		writer: bufio.NewWriter(t),
	}
	v.parseContextStack = append(v.parseContextStack, int(_CONTEXT_IN_TOPLEVEL))
	v.dumpContext = append(v.dumpContext, int(_CONTEXT_IN_TOPLEVEL))
	return v
}

// Factory
type JSONProtocolFactory struct{}

func (p *JSONProtocolFactory) GetProtocol(trans TTransport) TProtocol {
	return NewJSONProtocol(trans)
}

func NewJSONProtocolFactory() *JSONProtocolFactory {
	return &JSONProtocolFactory{}
}

func jsonQuote(s string) string {
	b, _ := json.Marshal(s)
	s1 := string(b)
	return s1
}

func jsonUnquote(s string) (string, bool) {
	s1 := new(string)
	err := json.Unmarshal([]byte(s), s1)
	return *s1, err == nil
}

func mismatch(expected, actual string) error {
	return fmt.Errorf("Expected '%s' but found '%s' while parsing JSON.", expected, actual)
}

func (p *JSONProtocol) WriteMessageBegin(name string, typeId TMessageType, seqId int32) error {
	return nil
}

func (p *JSONProtocol) WriteMessageEnd() error {
	return p.OutputListEnd()
}

func (p *JSONProtocol) WriteStructBegin(name string) error {
	if e := p.OutputObjectBegin(); e != nil {
		return e
	}
	return nil
}

func (p *JSONProtocol) WriteStructEnd() error {
	return p.OutputObjectEnd()
}

func (p *JSONProtocol) WriteFieldBegin(name string, typeId TType, id int16) error {
	if e := p.WriteString(name); e != nil {
		return e
	}
	return nil
}

func (p *JSONProtocol) WriteFieldEnd() error {
	//return p.OutputListEnd()
	return nil
}

func (p *JSONProtocol) WriteFieldStop() error { return nil }

func (p *JSONProtocol) WriteMapBegin(keyType TType, valueType TType, size int) error {
	return nil
}

func (p *JSONProtocol) WriteMapEnd() error {
	return p.OutputListEnd()
}

func (p *JSONProtocol) WriteListBegin(elemType TType, size int) error {
	return p.OutputElemListBegin(elemType, size)
}

func (p *JSONProtocol) WriteListEnd() error {
	return p.OutputListEnd()
}

func (p *JSONProtocol) WriteSetBegin(elemType TType, size int) error {
	return p.OutputElemListBegin(elemType, size)
}

func (p *JSONProtocol) WriteSetEnd() error {
	return p.OutputListEnd()
}

func (p *JSONProtocol) WriteBool(b bool) error {
	return p.OutputBool(b)
}

func (p *JSONProtocol) WriteByte(b int8) error {
	return p.WriteI32(int32(b))
}

func (p *JSONProtocol) WriteI16(v int16) error {
	return p.WriteI32(int32(v))
}

func (p *JSONProtocol) WriteI32(v int32) error {
	return p.OutputI64(int64(v))
}

func (p *JSONProtocol) WriteI64(v int64) error {
	return p.OutputI64(int64(v))
}

func (p *JSONProtocol) WriteDouble(v float64) error {
	return p.OutputF64(v)
}

func (p *JSONProtocol) WriteString(v string) error {
	return p.OutputString(v)
}

func (p *JSONProtocol) WriteBinary(v []byte) error {
	// JSON library only takes in a string,
	// not an arbitrary byte array, to ensure bytes are transmitted
	// efficiently we must convert this into a valid JSON string
	// therefore we use base64 encoding to avoid excessive escaping/quoting
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	if _, e := p.writer.Write(JSON_QUOTE_BYTES); e != nil {
		return NewTProtocolException(e)
	}
	writer := base64.NewEncoder(base64.StdEncoding, p.writer)
	if _, e := writer.Write(v); e != nil {
		return NewTProtocolException(e)
	}
	if e := writer.Close(); e != nil {
		return NewTProtocolException(e)
	}
	if _, e := p.writer.Write(JSON_QUOTE_BYTES); e != nil {
		return NewTProtocolException(e)
	}
	return p.OutputPostValue()
}

// Reading methods.

func (p *JSONProtocol) ReadMessageBegin() (name string, typeId TMessageType, seqId int32, err error) {
	return
}

func (p *JSONProtocol) ReadMessageEnd() error {
	return nil
}

func (p *JSONProtocol) ReadStructBegin() (name string, err error) {
	return "", nil
}

func (p *JSONProtocol) ReadStructEnd() error {
	return nil
}

func (p *JSONProtocol) ReadFieldBegin() (string, TType, int16, error) {
	return "", STOP, 0, nil
}

func (p *JSONProtocol) ReadFieldEnd() error {
	return nil
}

func (p *JSONProtocol) ReadMapBegin() (keyType TType, valueType TType, size int, e error) {
	return VOID, VOID, 0, nil
}

func (p *JSONProtocol) ReadMapEnd() error {
	return nil
}

func (p *JSONProtocol) ReadListBegin() (elemType TType, size int, e error) {
	return VOID, 0, nil
}

func (p *JSONProtocol) ReadListEnd() error {
	return nil
}

func (p *JSONProtocol) ReadSetBegin() (elemType TType, size int, e error) {
	return VOID, 0, nil
}

func (p *JSONProtocol) ReadSetEnd() error {
	return nil
}

func (p *JSONProtocol) ReadBool() (bool, error) {
	return false, nil
}

func (p *JSONProtocol) ReadByte() (int8, error) {
	return 0, nil
}

func (p *JSONProtocol) ReadI16() (int16, error) {
	return 0, nil
}


func (p *JSONProtocol) ReadI32() (int32, error) {
	return 0, nil
}

func (p *JSONProtocol) ReadI64() (int64, error) {
	return 0, nil
}

func (p *JSONProtocol) ReadDouble() (float64, error) {
	return 0, nil
}

func (p *JSONProtocol) ReadString() (string, error) {
	return "", nil
}

func (p *JSONProtocol) ReadBinary() (v []byte, err error) {
	return
}

func (p *JSONProtocol) Flush() (err error) {
	return NewTProtocolException(p.writer.Flush())
}

func (p *JSONProtocol) Skip(fieldType TType) (err error) {
	return SkipDefaultDepth(p, fieldType)
}

func (p *JSONProtocol) Transport() TTransport {
	return p.trans
}

func (p *JSONProtocol) OutputPreValue() error {
	cxt := _ParseContext(p.dumpContext[len(p.dumpContext)-1])
	switch cxt {
	case _CONTEXT_IN_LIST, _CONTEXT_IN_OBJECT_NEXT_KEY:
		if _, e := p.writer.Write(JSON_COMMA); e != nil {
			return NewTProtocolException(e)
		}
		break
	case _CONTEXT_IN_OBJECT_NEXT_VALUE:
		if _, e := p.writer.Write(JSON_COLON); e != nil {
			return NewTProtocolException(e)
		}
		break
	}
	return nil
}

func (p *JSONProtocol) OutputPostValue() error {
	cxt := _ParseContext(p.dumpContext[len(p.dumpContext)-1])
	switch cxt {
	case _CONTEXT_IN_LIST_FIRST:
		p.dumpContext = p.dumpContext[:len(p.dumpContext)-1]
		p.dumpContext = append(p.dumpContext, int(_CONTEXT_IN_LIST))
		break
	case _CONTEXT_IN_OBJECT_FIRST:
		p.dumpContext = p.dumpContext[:len(p.dumpContext)-1]
		p.dumpContext = append(p.dumpContext, int(_CONTEXT_IN_OBJECT_NEXT_VALUE))
		break
	case _CONTEXT_IN_OBJECT_NEXT_KEY:
		p.dumpContext = p.dumpContext[:len(p.dumpContext)-1]
		p.dumpContext = append(p.dumpContext, int(_CONTEXT_IN_OBJECT_NEXT_VALUE))
		break
	case _CONTEXT_IN_OBJECT_NEXT_VALUE:
		p.dumpContext = p.dumpContext[:len(p.dumpContext)-1]
		p.dumpContext = append(p.dumpContext, int(_CONTEXT_IN_OBJECT_NEXT_KEY))
		break
	}
	return nil
}

func (p *JSONProtocol) OutputBool(value bool) error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	var v string
	if value {
		v = string(JSON_TRUE)
	} else {
		v = string(JSON_FALSE)
	}
	switch _ParseContext(p.dumpContext[len(p.dumpContext)-1]) {
	case _CONTEXT_IN_OBJECT_FIRST, _CONTEXT_IN_OBJECT_NEXT_KEY:
		v = jsonQuote(v)
	default:
	}
	if e := p.OutputStringData(v); e != nil {
		return e
	}
	return p.OutputPostValue()
}

func (p *JSONProtocol) OutputNull() error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	if _, e := p.writer.Write(JSON_NULL); e != nil {
		return NewTProtocolException(e)
	}
	return p.OutputPostValue()
}

func (p *JSONProtocol) OutputF64(value float64) error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	var v string
	if math.IsNaN(value) {
		v = string(JSON_QUOTE) + JSON_NAN + string(JSON_QUOTE)
	} else if math.IsInf(value, 1) {
		v = string(JSON_QUOTE) + JSON_INFINITY + string(JSON_QUOTE)
	} else if math.IsInf(value, -1) {
		v = string(JSON_QUOTE) + JSON_NEGATIVE_INFINITY + string(JSON_QUOTE)
	} else {
		v = strconv.FormatFloat(value, 'g', -1, 64)
		switch _ParseContext(p.dumpContext[len(p.dumpContext)-1]) {
		case _CONTEXT_IN_OBJECT_FIRST, _CONTEXT_IN_OBJECT_NEXT_KEY:
			v = string(JSON_QUOTE) + v + string(JSON_QUOTE)
		default:
		}
	}
	if e := p.OutputStringData(v); e != nil {
		return e
	}
	return p.OutputPostValue()
}

func (p *JSONProtocol) OutputI64(value int64) error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	v := strconv.FormatInt(value, 10)
	switch _ParseContext(p.dumpContext[len(p.dumpContext)-1]) {
	case _CONTEXT_IN_OBJECT_FIRST, _CONTEXT_IN_OBJECT_NEXT_KEY:
		v = jsonQuote(v)
	default:
	}
	if e := p.OutputStringData(v); e != nil {
		return e
	}
	return p.OutputPostValue()
}

func (p *JSONProtocol) OutputString(s string) error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	if e := p.OutputStringData(jsonQuote(s)); e != nil {
		return e
	}
	return p.OutputPostValue()
}

func (p *JSONProtocol) OutputStringData(s string) error {
	_, e := p.writer.Write([]byte(s))
	return NewTProtocolException(e)
}

func (p *JSONProtocol) OutputObjectBegin() error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	if _, e := p.writer.Write(JSON_LBRACE); e != nil {
		return NewTProtocolException(e)
	}
	p.dumpContext = append(p.dumpContext, int(_CONTEXT_IN_OBJECT_FIRST))
	return nil
}

func (p *JSONProtocol) OutputObjectEnd() error {
	if _, e := p.writer.Write(JSON_RBRACE); e != nil {
		return NewTProtocolException(e)
	}
	p.dumpContext = p.dumpContext[:len(p.dumpContext)-1]
	if e := p.OutputPostValue(); e != nil {
		return e
	}
	return nil
}

func (p *JSONProtocol) OutputListBegin() error {
	if e := p.OutputPreValue(); e != nil {
		return e
	}
	if _, e := p.writer.Write(JSON_LBRACKET); e != nil {
		return NewTProtocolException(e)
	}
	p.dumpContext = append(p.dumpContext, int(_CONTEXT_IN_LIST_FIRST))
	return nil
}

func (p *JSONProtocol) OutputListEnd() error {
	if _, e := p.writer.Write(JSON_RBRACKET); e != nil {
		return NewTProtocolException(e)
	}
	p.dumpContext = p.dumpContext[:len(p.dumpContext)-1]
	if e := p.OutputPostValue(); e != nil {
		return e
	}
	return nil
}

func (p *JSONProtocol) OutputElemListBegin(elemType TType, size int) error {
	if e := p.OutputListBegin(); e != nil {
		return e
	}
	return nil
}
