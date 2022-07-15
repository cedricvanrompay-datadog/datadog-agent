// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: tracer_payload.proto

package pb

import (
	fmt "fmt"

	proto "github.com/gogo/protobuf/proto"

	math "math"

	_ "github.com/gogo/protobuf/gogoproto"

	io "io"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// TraceChunk represents a list of spans with the same trace id.
type TraceChunk struct {
	// priority specifies sampling priority of the trace.
	Priority int32 `protobuf:"varint,1,opt,name=priority,proto3" json:"priority" msg:"priority"`
	// origin specifies origin product ("lambda", "rum", etc.) of the trace.
	Origin string `protobuf:"bytes,2,opt,name=origin,proto3" json:"origin" msg:"origin"`
	// spans specifies list of containing spans.
	Spans []*Span `protobuf:"bytes,3,rep,name=spans" json:"spans" msg:"spans"`
	// tags specifies tags common in all `spans`.
	Tags map[string]string `protobuf:"bytes,4,rep,name=tags" json:"tags" msg:"tags" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// droppedTrace specifies whether the trace was dropped by samplers or not.
	DroppedTrace bool `protobuf:"varint,5,opt,name=droppedTrace,proto3" json:"dropped_trace" msg:"dropped_trace"`
}

func (m *TraceChunk) Reset()                    { *m = TraceChunk{} }
func (m *TraceChunk) String() string            { return proto.CompactTextString(m) }
func (*TraceChunk) ProtoMessage()               {}
func (*TraceChunk) Descriptor() ([]byte, []int) { return fileDescriptorTracerPayload, []int{0} }

func (m *TraceChunk) GetSpans() []*Span {
	if m != nil {
		return m.Spans
	}
	return nil
}

func (m *TraceChunk) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

// TracerPayload represents a payload the trace agent receives from tracers.
type TracerPayload struct {
	// containerID specifies the ID of the container where the tracer is running on.
	ContainerID string `protobuf:"bytes,1,opt,name=containerID,proto3" json:"container_id" msg:"container_id"`
	// languageName specifies language of the tracer.
	LanguageName string `protobuf:"bytes,2,opt,name=languageName,proto3" json:"language_name" msg:"language_name"`
	// languageVersion specifies language version of the tracer.
	LanguageVersion string `protobuf:"bytes,3,opt,name=languageVersion,proto3" json:"language_version" msg:"language_version"`
	// tracerVersion specifies version of the tracer.
	TracerVersion string `protobuf:"bytes,4,opt,name=tracerVersion,proto3" json:"tracer_version" msg:"tracer_version"`
	// runtimeID specifies V4 UUID representation of a tracer session.
	RuntimeID string `protobuf:"bytes,5,opt,name=runtimeID,proto3" json:"runtime_id" msg:"runtime_id"`
	// chunks specifies list of containing trace chunks.
	Chunks []*TraceChunk `protobuf:"bytes,6,rep,name=chunks" json:"chunks" msg:"chunks"`
	// tags specifies tags common in all `chunks`.
	Tags map[string]string `protobuf:"bytes,7,rep,name=tags" json:"tags" msg:"tags" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// env specifies `env` tag that set with the tracer.
	Env string `protobuf:"bytes,8,opt,name=env,proto3" json:"env" msg:"env"`
	// hostname specifies hostname of where the tracer is running.
	Hostname string `protobuf:"bytes,9,opt,name=hostname,proto3" json:"hostname" msg:"hostname"`
	// version specifies `version` tag that set with the tracer.
	AppVersion string `protobuf:"bytes,10,opt,name=appVersion,proto3" json:"app_version" msg:"app_version"`
}

func (m *TracerPayload) Reset()                    { *m = TracerPayload{} }
func (m *TracerPayload) String() string            { return proto.CompactTextString(m) }
func (*TracerPayload) ProtoMessage()               {}
func (*TracerPayload) Descriptor() ([]byte, []int) { return fileDescriptorTracerPayload, []int{1} }

func (m *TracerPayload) GetChunks() []*TraceChunk {
	if m != nil {
		return m.Chunks
	}
	return nil
}

func (m *TracerPayload) GetTags() map[string]string {
	if m != nil {
		return m.Tags
	}
	return nil
}

func init() {
	proto.RegisterType((*TraceChunk)(nil), "pb.TraceChunk")
	proto.RegisterType((*TracerPayload)(nil), "pb.TracerPayload")
}
func (m *TraceChunk) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *TraceChunk) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Priority != 0 {
		data[i] = 0x8
		i++
		i = encodeVarintTracerPayload(data, i, uint64(m.Priority))
	}
	if len(m.Origin) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.Origin)))
		i += copy(data[i:], m.Origin)
	}
	if len(m.Spans) > 0 {
		for _, msg := range m.Spans {
			data[i] = 0x1a
			i++
			i = encodeVarintTracerPayload(data, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Tags) > 0 {
		for k := range m.Tags {
			data[i] = 0x22
			i++
			v := m.Tags[k]
			mapSize := 1 + len(k) + sovTracerPayload(uint64(len(k))) + 1 + len(v) + sovTracerPayload(uint64(len(v)))
			i = encodeVarintTracerPayload(data, i, uint64(mapSize))
			data[i] = 0xa
			i++
			i = encodeVarintTracerPayload(data, i, uint64(len(k)))
			i += copy(data[i:], k)
			data[i] = 0x12
			i++
			i = encodeVarintTracerPayload(data, i, uint64(len(v)))
			i += copy(data[i:], v)
		}
	}
	if m.DroppedTrace {
		data[i] = 0x28
		i++
		if m.DroppedTrace {
			data[i] = 1
		} else {
			data[i] = 0
		}
		i++
	}
	return i, nil
}

func (m *TracerPayload) Marshal() (data []byte, err error) {
	size := m.Size()
	data = make([]byte, size)
	n, err := m.MarshalTo(data)
	if err != nil {
		return nil, err
	}
	return data[:n], nil
}

func (m *TracerPayload) MarshalTo(data []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if len(m.ContainerID) > 0 {
		data[i] = 0xa
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.ContainerID)))
		i += copy(data[i:], m.ContainerID)
	}
	if len(m.LanguageName) > 0 {
		data[i] = 0x12
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.LanguageName)))
		i += copy(data[i:], m.LanguageName)
	}
	if len(m.LanguageVersion) > 0 {
		data[i] = 0x1a
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.LanguageVersion)))
		i += copy(data[i:], m.LanguageVersion)
	}
	if len(m.TracerVersion) > 0 {
		data[i] = 0x22
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.TracerVersion)))
		i += copy(data[i:], m.TracerVersion)
	}
	if len(m.RuntimeID) > 0 {
		data[i] = 0x2a
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.RuntimeID)))
		i += copy(data[i:], m.RuntimeID)
	}
	if len(m.Chunks) > 0 {
		for _, msg := range m.Chunks {
			data[i] = 0x32
			i++
			i = encodeVarintTracerPayload(data, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(data[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if len(m.Tags) > 0 {
		for k := range m.Tags {
			data[i] = 0x3a
			i++
			v := m.Tags[k]
			mapSize := 1 + len(k) + sovTracerPayload(uint64(len(k))) + 1 + len(v) + sovTracerPayload(uint64(len(v)))
			i = encodeVarintTracerPayload(data, i, uint64(mapSize))
			data[i] = 0xa
			i++
			i = encodeVarintTracerPayload(data, i, uint64(len(k)))
			i += copy(data[i:], k)
			data[i] = 0x12
			i++
			i = encodeVarintTracerPayload(data, i, uint64(len(v)))
			i += copy(data[i:], v)
		}
	}
	if len(m.Env) > 0 {
		data[i] = 0x42
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.Env)))
		i += copy(data[i:], m.Env)
	}
	if len(m.Hostname) > 0 {
		data[i] = 0x4a
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.Hostname)))
		i += copy(data[i:], m.Hostname)
	}
	if len(m.AppVersion) > 0 {
		data[i] = 0x52
		i++
		i = encodeVarintTracerPayload(data, i, uint64(len(m.AppVersion)))
		i += copy(data[i:], m.AppVersion)
	}
	return i, nil
}

func encodeFixed64TracerPayload(data []byte, offset int, v uint64) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	data[offset+4] = uint8(v >> 32)
	data[offset+5] = uint8(v >> 40)
	data[offset+6] = uint8(v >> 48)
	data[offset+7] = uint8(v >> 56)
	return offset + 8
}
func encodeFixed32TracerPayload(data []byte, offset int, v uint32) int {
	data[offset] = uint8(v)
	data[offset+1] = uint8(v >> 8)
	data[offset+2] = uint8(v >> 16)
	data[offset+3] = uint8(v >> 24)
	return offset + 4
}
func encodeVarintTracerPayload(data []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		data[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	data[offset] = uint8(v)
	return offset + 1
}
func (m *TraceChunk) Size() (n int) {
	var l int
	_ = l
	if m.Priority != 0 {
		n += 1 + sovTracerPayload(uint64(m.Priority))
	}
	l = len(m.Origin)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	if len(m.Spans) > 0 {
		for _, e := range m.Spans {
			l = e.Size()
			n += 1 + l + sovTracerPayload(uint64(l))
		}
	}
	if len(m.Tags) > 0 {
		for k, v := range m.Tags {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovTracerPayload(uint64(len(k))) + 1 + len(v) + sovTracerPayload(uint64(len(v)))
			n += mapEntrySize + 1 + sovTracerPayload(uint64(mapEntrySize))
		}
	}
	if m.DroppedTrace {
		n += 2
	}
	return n
}

func (m *TracerPayload) Size() (n int) {
	var l int
	_ = l
	l = len(m.ContainerID)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	l = len(m.LanguageName)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	l = len(m.LanguageVersion)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	l = len(m.TracerVersion)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	l = len(m.RuntimeID)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	if len(m.Chunks) > 0 {
		for _, e := range m.Chunks {
			l = e.Size()
			n += 1 + l + sovTracerPayload(uint64(l))
		}
	}
	if len(m.Tags) > 0 {
		for k, v := range m.Tags {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovTracerPayload(uint64(len(k))) + 1 + len(v) + sovTracerPayload(uint64(len(v)))
			n += mapEntrySize + 1 + sovTracerPayload(uint64(mapEntrySize))
		}
	}
	l = len(m.Env)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	l = len(m.Hostname)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	l = len(m.AppVersion)
	if l > 0 {
		n += 1 + l + sovTracerPayload(uint64(l))
	}
	return n
}

func sovTracerPayload(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozTracerPayload(x uint64) (n int) {
	return sovTracerPayload(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TraceChunk) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTracerPayload
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TraceChunk: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TraceChunk: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Priority", wireType)
			}
			m.Priority = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				m.Priority |= (int32(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Origin", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Origin = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Spans", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Spans = append(m.Spans, &Span{})
			if err := m.Spans[len(m.Spans)-1].Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tags", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var keykey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				keykey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			var stringLenmapkey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLenmapkey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLenmapkey := int(stringLenmapkey)
			if intStringLenmapkey < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postStringIndexmapkey := iNdEx + intStringLenmapkey
			if postStringIndexmapkey > l {
				return io.ErrUnexpectedEOF
			}
			mapkey := string(data[iNdEx:postStringIndexmapkey])
			iNdEx = postStringIndexmapkey
			if m.Tags == nil {
				m.Tags = make(map[string]string)
			}
			if iNdEx < postIndex {
				var valuekey uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTracerPayload
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					valuekey |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				var stringLenmapvalue uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTracerPayload
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					stringLenmapvalue |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLenmapvalue := int(stringLenmapvalue)
				if intStringLenmapvalue < 0 {
					return ErrInvalidLengthTracerPayload
				}
				postStringIndexmapvalue := iNdEx + intStringLenmapvalue
				if postStringIndexmapvalue > l {
					return io.ErrUnexpectedEOF
				}
				mapvalue := string(data[iNdEx:postStringIndexmapvalue])
				iNdEx = postStringIndexmapvalue
				m.Tags[mapkey] = mapvalue
			} else {
				var mapvalue string
				m.Tags[mapkey] = mapvalue
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field DroppedTrace", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				v |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.DroppedTrace = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipTracerPayload(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTracerPayload
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *TracerPayload) Unmarshal(data []byte) error {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTracerPayload
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TracerPayload: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TracerPayload: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContainerID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContainerID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LanguageName", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LanguageName = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LanguageVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.LanguageVersion = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TracerVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.TracerVersion = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RuntimeID", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RuntimeID = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Chunks", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Chunks = append(m.Chunks, &TraceChunk{})
			if err := m.Chunks[len(m.Chunks)-1].Unmarshal(data[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Tags", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			var keykey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				keykey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			var stringLenmapkey uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLenmapkey |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLenmapkey := int(stringLenmapkey)
			if intStringLenmapkey < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postStringIndexmapkey := iNdEx + intStringLenmapkey
			if postStringIndexmapkey > l {
				return io.ErrUnexpectedEOF
			}
			mapkey := string(data[iNdEx:postStringIndexmapkey])
			iNdEx = postStringIndexmapkey
			if m.Tags == nil {
				m.Tags = make(map[string]string)
			}
			if iNdEx < postIndex {
				var valuekey uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTracerPayload
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					valuekey |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				var stringLenmapvalue uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowTracerPayload
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					stringLenmapvalue |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				intStringLenmapvalue := int(stringLenmapvalue)
				if intStringLenmapvalue < 0 {
					return ErrInvalidLengthTracerPayload
				}
				postStringIndexmapvalue := iNdEx + intStringLenmapvalue
				if postStringIndexmapvalue > l {
					return io.ErrUnexpectedEOF
				}
				mapvalue := string(data[iNdEx:postStringIndexmapvalue])
				iNdEx = postStringIndexmapvalue
				m.Tags[mapkey] = mapvalue
			} else {
				var mapvalue string
				m.Tags[mapkey] = mapvalue
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Env", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Env = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 9:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Hostname", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Hostname = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field AppVersion", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthTracerPayload
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.AppVersion = string(data[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTracerPayload(data[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthTracerPayload
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipTracerPayload(data []byte) (n int, err error) {
	l := len(data)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTracerPayload
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := data[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if data[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowTracerPayload
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := data[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthTracerPayload
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowTracerPayload
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := data[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipTracerPayload(data[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthTracerPayload = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTracerPayload   = fmt.Errorf("proto: integer overflow")
)

func init() { proto.RegisterFile("tracer_payload.proto", fileDescriptorTracerPayload) }

var fileDescriptorTracerPayload = []byte{
	// 630 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x94, 0x4f, 0x6b, 0xdb, 0x4c,
	0x10, 0xc6, 0x5f, 0xc5, 0xb1, 0x5f, 0x6b, 0x9c, 0xa4, 0xee, 0xd6, 0x04, 0xe1, 0x82, 0xd7, 0x2c,
	0x25, 0x98, 0x42, 0x1d, 0x48, 0x0f, 0x0d, 0x39, 0x15, 0xd5, 0xa5, 0x0d, 0x94, 0x12, 0xb6, 0xa1,
	0xf4, 0x66, 0xd6, 0xb6, 0xaa, 0x88, 0xc4, 0xab, 0x45, 0x92, 0x0d, 0x39, 0x97, 0xde, 0xfb, 0xb1,
	0x7a, 0xec, 0x27, 0x58, 0x8a, 0x7b, 0xd3, 0x51, 0x9f, 0xa0, 0x68, 0x56, 0x92, 0xad, 0x9e, 0x4a,
	0x6f, 0x3b, 0xbf, 0xd9, 0xe7, 0xb1, 0xe6, 0xcf, 0x1a, 0x7a, 0x49, 0x24, 0xe6, 0x5e, 0x34, 0x55,
	0xe2, 0xfe, 0x2e, 0x14, 0x8b, 0xb1, 0x8a, 0xc2, 0x24, 0x24, 0x7b, 0x6a, 0xd6, 0x7f, 0xe6, 0x07,
	0xc9, 0xcd, 0x6a, 0x36, 0x9e, 0x87, 0xcb, 0x53, 0x3f, 0xf4, 0xc3, 0x53, 0x4c, 0xcd, 0x56, 0x9f,
	0x31, 0xc2, 0x00, 0x4f, 0x46, 0xd2, 0x87, 0x58, 0x09, 0x69, 0xce, 0xec, 0x4b, 0x03, 0xe0, 0x3a,
	0xf7, 0x7d, 0x75, 0xb3, 0x92, 0xb7, 0xe4, 0x02, 0xda, 0x2a, 0x0a, 0xc2, 0x28, 0x48, 0xee, 0x1d,
	0x6b, 0x68, 0x8d, 0x9a, 0xee, 0x20, 0xd5, 0xb4, 0x62, 0x99, 0xa6, 0x47, 0xcb, 0xd8, 0xbf, 0x60,
	0x25, 0x60, 0xbc, 0xca, 0x91, 0x33, 0x68, 0x85, 0x51, 0xe0, 0x07, 0xd2, 0xd9, 0x1b, 0x5a, 0x23,
	0xdb, 0xed, 0xa7, 0x9a, 0x16, 0x24, 0xd3, 0xf4, 0x00, 0x75, 0x26, 0x64, 0xbc, 0xe0, 0xe4, 0x1c,
	0x9a, 0xf9, 0xc7, 0xc4, 0x4e, 0x63, 0xd8, 0x18, 0x75, 0xce, 0xda, 0x63, 0x35, 0x1b, 0x7f, 0x50,
	0x42, 0xba, 0x4e, 0xaa, 0xa9, 0x49, 0x65, 0x9a, 0x76, 0x50, 0x8b, 0x11, 0xe3, 0x86, 0x92, 0x09,
	0xec, 0x27, 0xc2, 0x8f, 0x9d, 0x7d, 0x14, 0x3a, 0xb9, 0x70, 0x5b, 0xc7, 0xf8, 0x5a, 0xf8, 0xf1,
	0x6b, 0x99, 0x44, 0xf7, 0xee, 0x71, 0xaa, 0x29, 0xde, 0xcc, 0x34, 0x05, 0xf4, 0xc9, 0x03, 0xc6,
	0x91, 0x91, 0x77, 0x70, 0xb0, 0x88, 0x42, 0xa5, 0xbc, 0x05, 0x8a, 0x9d, 0xe6, 0xd0, 0x1a, 0xb5,
	0xdd, 0x51, 0xaa, 0xe9, 0x61, 0xc1, 0xa7, 0xd8, 0xf5, 0x4c, 0xd3, 0x47, 0x28, 0xae, 0x51, 0xc6,
	0x6b, 0xea, 0xfe, 0x0b, 0xb0, 0xab, 0x1f, 0x26, 0x5d, 0x68, 0xdc, 0x7a, 0xa6, 0x8b, 0x36, 0xcf,
	0x8f, 0xa4, 0x07, 0xcd, 0xb5, 0xb8, 0x5b, 0x79, 0xa6, 0x3f, 0xdc, 0x04, 0x17, 0x7b, 0xe7, 0x16,
	0xfb, 0xda, 0x82, 0x43, 0xb4, 0x88, 0xae, 0xcc, 0x70, 0xc9, 0x5b, 0xe8, 0xcc, 0x43, 0x99, 0x88,
	0x40, 0x7a, 0xd1, 0xe5, 0xc4, 0xb8, 0xb8, 0x27, 0xa9, 0xa6, 0x07, 0x15, 0x9e, 0x06, 0x8b, 0x4c,
	0x53, 0x82, 0x9f, 0xb5, 0x0b, 0x19, 0xdf, 0x95, 0xe6, 0x25, 0xde, 0x09, 0xe9, 0xaf, 0x84, 0xef,
	0xbd, 0x17, 0xcb, 0xe2, 0xc7, 0x4d, 0x89, 0x25, 0x9f, 0x4a, 0xb1, 0xdc, 0x96, 0x58, 0xa3, 0x8c,
	0xd7, 0xd4, 0xe4, 0x13, 0x3c, 0x28, 0xe3, 0x8f, 0x5e, 0x14, 0x07, 0xa1, 0x74, 0x1a, 0x68, 0x38,
	0x4e, 0x35, 0xed, 0x56, 0xd2, 0xb5, 0xc9, 0x65, 0x9a, 0x1e, 0xd7, 0x3d, 0x8b, 0x04, 0xe3, 0x7f,
	0xda, 0x90, 0x2b, 0x38, 0x34, 0x0b, 0x5e, 0xfa, 0xee, 0xa3, 0xef, 0xd3, 0x54, 0xd3, 0xa3, 0x62,
	0xf3, 0xb7, 0xae, 0x3d, 0x33, 0xc9, 0x1a, 0x66, 0xbc, 0x6e, 0x40, 0x5e, 0x82, 0x1d, 0xad, 0x64,
	0x12, 0x2c, 0xbd, 0xcb, 0x09, 0x4e, 0xd6, 0x76, 0x59, 0xaa, 0x29, 0x14, 0xd0, 0xf4, 0xaf, 0x8b,
	0x4e, 0x5b, 0xc4, 0xf8, 0x56, 0x44, 0x5c, 0x68, 0xcd, 0xf3, 0x7d, 0x8a, 0x9d, 0x16, 0xae, 0xd9,
	0x51, 0x7d, 0xcd, 0xcc, 0x8a, 0x9b, 0x1b, 0xd5, 0x8a, 0x9b, 0x90, 0xf1, 0x82, 0x93, 0x37, 0xc5,
	0xa2, 0xfe, 0x8f, 0x0e, 0x8f, 0x2b, 0x87, 0x72, 0xd4, 0x7f, 0xbd, 0xab, 0x27, 0xd0, 0xf0, 0xe4,
	0xda, 0x69, 0x63, 0x21, 0xbd, 0x54, 0xd3, 0x3c, 0xcc, 0x34, 0xb5, 0xf1, 0xa6, 0x27, 0xd7, 0x8c,
	0xe7, 0x24, 0x7f, 0xc3, 0x37, 0x61, 0x9c, 0xe4, 0xd3, 0x73, 0x6c, 0xbc, 0x8c, 0x6f, 0xb8, 0x64,
	0xd5, 0x1b, 0x2e, 0x01, 0xe3, 0x55, 0x8e, 0x4c, 0x00, 0x84, 0x52, 0xe5, 0x04, 0x00, 0xd5, 0x4f,
	0x52, 0x4d, 0x3b, 0x42, 0xa9, 0x9d, 0xf6, 0x3f, 0x44, 0x83, 0x1d, 0xc6, 0xf8, 0x8e, 0xee, 0x9f,
	0xdf, 0x81, 0xdb, 0xfd, 0xbe, 0x19, 0x58, 0x3f, 0x36, 0x03, 0xeb, 0xe7, 0x66, 0x60, 0x7d, 0xfb,
	0x35, 0xf8, 0x6f, 0xd6, 0xc2, 0xbf, 0xa9, 0xe7, 0xbf, 0x03, 0x00, 0x00, 0xff, 0xff, 0xfa, 0x46,
	0xa3, 0x77, 0xfd, 0x04, 0x00, 0x00,
}
