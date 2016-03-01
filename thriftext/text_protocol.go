/**
 * TTextProtocol 
 * archive使用该格式可以将对象输出为文本格式，便于直接放入hive
 * @WARNNING:
 *		只实现了Writer，没有实现Read，有需要的同学自己动手吧。
 */
package thriftext

import (
	"github.com/DrWrong/monica/thrift"
	"strconv"
)

type TTextProtocol struct {
	trans thrift.TTransport
}

func NewTTextProtocolTransport(trans thrift.TTransport) *TTextProtocol {
	return &TTextProtocol{trans}
}

func (p *TTextProtocol) WriteMessageBegin(name string, typeId thrift.TMessageType, seqid int32) error {
	return nil
}

func (p *TTextProtocol) WriteMessageEnd() error {
	return nil
}

func (p *TTextProtocol) WriteStructBegin(name string) error {
	return nil
}

func (p *TTextProtocol) WriteStructEnd() error {
	p.WriteString("\n")
	return nil
}

func (p *TTextProtocol) WriteFieldBegin(name string, typeId thrift.TType, id int16) error {
	return nil
}

func (p *TTextProtocol) WriteFieldEnd() error {
	p.WriteString("\t")
	return nil
}

func (p *TTextProtocol) WriteFieldStop() error {
	return nil
}

func (p *TTextProtocol) WriteMapBegin(keyType thrift.TType, valueType thrift.TType, size int) error {
	return nil
}

func (p *TTextProtocol) WriteMapEnd() error {
	return nil
}

func (p *TTextProtocol) WriteListBegin(elemType thrift.TType, size int) error {
	return nil
}

func (p *TTextProtocol) WriteListEnd() error {
	return nil
}

func (p *TTextProtocol) WriteSetBegin(elemType thrift.TType, size int) error {
	return nil
}

func (p *TTextProtocol) WriteSetEnd() error {
	return nil
}

func (p *TTextProtocol) WriteBool(value bool) error {
	if value {
		p.WriteString("1")
	} else {
		p.WriteString("0")
	}
	return nil
}

func (p *TTextProtocol) WriteByte(value byte) error {
	p.WriteI32(int32(value))
	return nil
}

func (p *TTextProtocol) WriteI16(value int16) error {
	p.WriteI64(int64(value))
	return nil
}

func (p *TTextProtocol) WriteI32(value int32) error {
	p.WriteI64(int64(value))
	return nil
}

func (p *TTextProtocol) WriteI64(value int64) error {
	v := strconv.FormatInt(value, 10)
	p.WriteString(v)
	return nil
}

func (p *TTextProtocol) WriteDouble(value float64) error {
	v := strconv.FormatFloat(value, 'g', -1, 64)
	p.WriteString(v)
	return nil
}

func (p *TTextProtocol) WriteString(value string) error {
	p.trans.Write([]byte(value))
	return nil
}

func (p *TTextProtocol) WriteBinary(value []byte) error {
	p.trans.Write(value)
	return nil
}

func (p *TTextProtocol) ReadMessageBegin() (name string, typeId thrift.TMessageType, seqid int32, err error) {
	return
}

func (p *TTextProtocol) ReadMessageEnd() error {
	return nil
}

func (p *TTextProtocol) ReadStructBegin() (name string, err error) {
	return
}

func (p *TTextProtocol) ReadStructEnd() error {
	return nil
}

func (p *TTextProtocol) ReadFieldBegin() (name string, typeId thrift.TType, id int16, err error) {
	return
}

func (p *TTextProtocol) ReadFieldEnd() error {
	return nil
}

func (p *TTextProtocol) ReadMapBegin() (keyType thrift.TType, valueType thrift.TType, size int, err error) {
	return
}

func (p *TTextProtocol) ReadMapEnd() error {
	return nil
}

func (p *TTextProtocol) ReadListBegin() (elemType thrift.TType, size int, err error) {
	return
}

func (p *TTextProtocol) ReadListEnd() error {
	return nil
}

func (p *TTextProtocol) ReadSetBegin() (elemType thrift.TType, size int, err error) {
	return
}

func (p *TTextProtocol) ReadSetEnd() error {
	return nil
}

func (p *TTextProtocol) ReadBool() (value bool, err error) {
	return
}

func (p *TTextProtocol) ReadByte() (value byte, err error) {
	return
}

func (p *TTextProtocol) ReadI16() (value int16, err error) {
	return
}

func (p *TTextProtocol) ReadI32() (value int32, err error) {
	return
}

func (p *TTextProtocol) ReadI64() (value int64, err error) {
	return
}

func (p *TTextProtocol) ReadDouble() (value float64, err error) {
	return
}

func (p *TTextProtocol) ReadString() (value string, err error) {
	return
}

func (p *TTextProtocol) ReadBinary() (value []byte, err error) {
	return
}

func (p *TTextProtocol) Skip(fieldType thrift.TType) (err error) {
	return nil
}

func (p *TTextProtocol) Flush() (err error) {
	return p.trans.Flush()
}

func (p *TTextProtocol) Transport() thrift.TTransport {
	return p.trans
}
