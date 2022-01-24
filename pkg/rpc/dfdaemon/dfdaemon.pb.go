//
//     Copyright 2020 The Dragonfly Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.17.3
// source: pkg/rpc/dfdaemon/dfdaemon.proto

package dfdaemon

import (
	base "d7y.io/dragonfly/v2/pkg/rpc/base"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type DownRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// identify one downloading, the framework will fill it automatically
	Uuid string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	// download file from the url, not only for http
	Url string `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	// pieces will be written to output path directly,
	// at the same time, dfdaemon workspace also makes soft link to the output
	Output string `protobuf:"bytes,3,opt,name=output,proto3" json:"output,omitempty"`
	// timeout duration
	Timeout uint64 `protobuf:"varint,4,opt,name=timeout,proto3" json:"timeout,omitempty"`
	// rate limit in bytes per second
	Limit             float64       `protobuf:"fixed64,5,opt,name=limit,proto3" json:"limit,omitempty"`
	DisableBackSource bool          `protobuf:"varint,6,opt,name=disable_back_source,json=disableBackSource,proto3" json:"disable_back_source,omitempty"`
	UrlMeta           *base.UrlMeta `protobuf:"bytes,7,opt,name=url_meta,json=urlMeta,proto3" json:"url_meta,omitempty"`
	// p2p/cdn/source, default is p2p
	Pattern string `protobuf:"bytes,8,opt,name=pattern,proto3" json:"pattern,omitempty"`
	// call system
	Callsystem string `protobuf:"bytes,9,opt,name=callsystem,proto3" json:"callsystem,omitempty"`
	// user id
	Uid int64 `protobuf:"varint,10,opt,name=uid,proto3" json:"uid,omitempty"`
	// group id
	Gid int64 `protobuf:"varint,11,opt,name=gid,proto3" json:"gid,omitempty"`
}

func (x *DownRequest) Reset() {
	*x = DownRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownRequest) ProtoMessage() {}

func (x *DownRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownRequest.ProtoReflect.Descriptor instead.
func (*DownRequest) Descriptor() ([]byte, []int) {
	return file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescGZIP(), []int{0}
}

func (x *DownRequest) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *DownRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *DownRequest) GetOutput() string {
	if x != nil {
		return x.Output
	}
	return ""
}

func (x *DownRequest) GetTimeout() uint64 {
	if x != nil {
		return x.Timeout
	}
	return 0
}

func (x *DownRequest) GetLimit() float64 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *DownRequest) GetDisableBackSource() bool {
	if x != nil {
		return x.DisableBackSource
	}
	return false
}

func (x *DownRequest) GetUrlMeta() *base.UrlMeta {
	if x != nil {
		return x.UrlMeta
	}
	return nil
}

func (x *DownRequest) GetPattern() string {
	if x != nil {
		return x.Pattern
	}
	return ""
}

func (x *DownRequest) GetCallsystem() string {
	if x != nil {
		return x.Callsystem
	}
	return ""
}

func (x *DownRequest) GetUid() int64 {
	if x != nil {
		return x.Uid
	}
	return 0
}

func (x *DownRequest) GetGid() int64 {
	if x != nil {
		return x.Gid
	}
	return 0
}

type DownResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TaskId          string `protobuf:"bytes,2,opt,name=task_id,json=taskId,proto3" json:"task_id,omitempty"`
	PeerId          string `protobuf:"bytes,3,opt,name=peer_id,json=peerId,proto3" json:"peer_id,omitempty"`
	CompletedLength uint64 `protobuf:"varint,4,opt,name=completed_length,json=completedLength,proto3" json:"completed_length,omitempty"`
	Done            bool   `protobuf:"varint,5,opt,name=done,proto3" json:"done,omitempty"`
}

func (x *DownResult) Reset() {
	*x = DownResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownResult) ProtoMessage() {}

func (x *DownResult) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownResult.ProtoReflect.Descriptor instead.
func (*DownResult) Descriptor() ([]byte, []int) {
	return file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescGZIP(), []int{1}
}

func (x *DownResult) GetTaskId() string {
	if x != nil {
		return x.TaskId
	}
	return ""
}

func (x *DownResult) GetPeerId() string {
	if x != nil {
		return x.PeerId
	}
	return ""
}

func (x *DownResult) GetCompletedLength() uint64 {
	if x != nil {
		return x.CompletedLength
	}
	return 0
}

func (x *DownResult) GetDone() bool {
	if x != nil {
		return x.Done
	}
	return false
}

type StatTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url       string        `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	UrlMeta   *base.UrlMeta `protobuf:"bytes,2,opt,name=url_meta,json=urlMeta,proto3" json:"url_meta,omitempty"`
	LocalOnly bool          `protobuf:"varint,3,opt,name=local_only,json=localOnly,proto3" json:"local_only,omitempty"`
}

func (x *StatTaskRequest) Reset() {
	*x = StatTaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StatTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StatTaskRequest) ProtoMessage() {}

func (x *StatTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StatTaskRequest.ProtoReflect.Descriptor instead.
func (*StatTaskRequest) Descriptor() ([]byte, []int) {
	return file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescGZIP(), []int{2}
}

func (x *StatTaskRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *StatTaskRequest) GetUrlMeta() *base.UrlMeta {
	if x != nil {
		return x.UrlMeta
	}
	return nil
}

func (x *StatTaskRequest) GetLocalOnly() bool {
	if x != nil {
		return x.LocalOnly
	}
	return false
}

type ImportTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url     string        `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	UrlMeta *base.UrlMeta `protobuf:"bytes,2,opt,name=url_meta,json=urlMeta,proto3" json:"url_meta,omitempty"`
	Path    string        `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
}

func (x *ImportTaskRequest) Reset() {
	*x = ImportTaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImportTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImportTaskRequest) ProtoMessage() {}

func (x *ImportTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImportTaskRequest.ProtoReflect.Descriptor instead.
func (*ImportTaskRequest) Descriptor() ([]byte, []int) {
	return file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescGZIP(), []int{3}
}

func (x *ImportTaskRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *ImportTaskRequest) GetUrlMeta() *base.UrlMeta {
	if x != nil {
		return x.UrlMeta
	}
	return nil
}

func (x *ImportTaskRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

type ExportTaskRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Url       string        `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	UrlMeta   *base.UrlMeta `protobuf:"bytes,2,opt,name=url_meta,json=urlMeta,proto3" json:"url_meta,omitempty"`
	Path      string        `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	LocalOnly bool          `protobuf:"varint,4,opt,name=local_only,json=localOnly,proto3" json:"local_only,omitempty"` // TODO: other DownRequest common fields
}

func (x *ExportTaskRequest) Reset() {
	*x = ExportTaskRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExportTaskRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExportTaskRequest) ProtoMessage() {}

func (x *ExportTaskRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExportTaskRequest.ProtoReflect.Descriptor instead.
func (*ExportTaskRequest) Descriptor() ([]byte, []int) {
	return file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescGZIP(), []int{4}
}

func (x *ExportTaskRequest) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *ExportTaskRequest) GetUrlMeta() *base.UrlMeta {
	if x != nil {
		return x.UrlMeta
	}
	return nil
}

func (x *ExportTaskRequest) GetPath() string {
	if x != nil {
		return x.Path
	}
	return ""
}

func (x *ExportTaskRequest) GetLocalOnly() bool {
	if x != nil {
		return x.LocalOnly
	}
	return false
}

var File_pkg_rpc_dfdaemon_dfdaemon_proto protoreflect.FileDescriptor

var file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x70, 0x6b, 0x67, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d,
	0x6f, 0x6e, 0x2f, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x08, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x1a, 0x17, 0x70, 0x6b, 0x67,
	0x2f, 0x72, 0x70, 0x63, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2f, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x17, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69,
	0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x85, 0x03, 0x0a, 0x0b, 0x44,
	0x6f, 0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1c, 0x0a, 0x04, 0x75, 0x75,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x72, 0x03, 0xb0,
	0x01, 0x01, 0x52, 0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x1a, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x72, 0x03, 0x88, 0x01, 0x01, 0x52,
	0x03, 0x75, 0x72, 0x6c, 0x12, 0x1f, 0x0a, 0x06, 0x6f, 0x75, 0x74, 0x70, 0x75, 0x74, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x06, 0x6f,
	0x75, 0x74, 0x70, 0x75, 0x74, 0x12, 0x21, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x32, 0x02, 0x28, 0x00, 0x52,
	0x07, 0x74, 0x69, 0x6d, 0x65, 0x6f, 0x75, 0x74, 0x12, 0x24, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69,
	0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x42, 0x0e, 0xfa, 0x42, 0x0b, 0x12, 0x09, 0x29, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x2e,
	0x0a, 0x13, 0x64, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x62, 0x61, 0x63, 0x6b, 0x5f, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x11, 0x64, 0x69, 0x73,
	0x61, 0x62, 0x6c, 0x65, 0x42, 0x61, 0x63, 0x6b, 0x53, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x28,
	0x0a, 0x08, 0x75, 0x72, 0x6c, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x0d, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x55, 0x72, 0x6c, 0x4d, 0x65, 0x74, 0x61, 0x52,
	0x07, 0x75, 0x72, 0x6c, 0x4d, 0x65, 0x74, 0x61, 0x12, 0x34, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74,
	0x65, 0x72, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x42, 0x1a, 0xfa, 0x42, 0x17, 0x72, 0x15,
	0x52, 0x03, 0x70, 0x32, 0x70, 0x52, 0x03, 0x63, 0x64, 0x6e, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0xd0, 0x01, 0x01, 0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x1e,
	0x0a, 0x0a, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x63, 0x61, 0x6c, 0x6c, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x12, 0x10,
	0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x75, 0x69, 0x64,
	0x12, 0x10, 0x0a, 0x03, 0x67, 0x69, 0x64, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x67,
	0x69, 0x64, 0x22, 0x98, 0x01, 0x0a, 0x0a, 0x44, 0x6f, 0x77, 0x6e, 0x52, 0x65, 0x73, 0x75, 0x6c,
	0x74, 0x12, 0x20, 0x0a, 0x07, 0x74, 0x61, 0x73, 0x6b, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x06, 0x74, 0x61, 0x73,
	0x6b, 0x49, 0x64, 0x12, 0x20, 0x0a, 0x07, 0x70, 0x65, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x06, 0x70,
	0x65, 0x65, 0x72, 0x49, 0x64, 0x12, 0x32, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74,
	0x65, 0x64, 0x5f, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x42,
	0x07, 0xfa, 0x42, 0x04, 0x32, 0x02, 0x28, 0x00, 0x52, 0x0f, 0x63, 0x6f, 0x6d, 0x70, 0x6c, 0x65,
	0x74, 0x65, 0x64, 0x4c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x6f, 0x6e,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x64, 0x6f, 0x6e, 0x65, 0x22, 0x75, 0x0a,
	0x0f, 0x53, 0x74, 0x61, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x19, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa,
	0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x28, 0x0a, 0x08, 0x75,
	0x72, 0x6c, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e,
	0x62, 0x61, 0x73, 0x65, 0x2e, 0x55, 0x72, 0x6c, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x07, 0x75, 0x72,
	0x6c, 0x4d, 0x65, 0x74, 0x61, 0x12, 0x1d, 0x0a, 0x0a, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x6f,
	0x6e, 0x6c, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x6c, 0x6f, 0x63, 0x61, 0x6c,
	0x4f, 0x6e, 0x6c, 0x79, 0x22, 0x75, 0x0a, 0x11, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61,
	0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x19, 0x0a, 0x03, 0x75, 0x72, 0x6c,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52,
	0x03, 0x75, 0x72, 0x6c, 0x12, 0x28, 0x0a, 0x08, 0x75, 0x72, 0x6c, 0x5f, 0x6d, 0x65, 0x74, 0x61,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x55, 0x72,
	0x6c, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x07, 0x75, 0x72, 0x6c, 0x4d, 0x65, 0x74, 0x61, 0x12, 0x1b,
	0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42,
	0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x04, 0x70, 0x61, 0x74, 0x68, 0x22, 0x94, 0x01, 0x0a, 0x11,
	0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x19, 0x0a, 0x03, 0x75, 0x72, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x42, 0x07,
	0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x03, 0x75, 0x72, 0x6c, 0x12, 0x28, 0x0a, 0x08,
	0x75, 0x72, 0x6c, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0d,
	0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x55, 0x72, 0x6c, 0x4d, 0x65, 0x74, 0x61, 0x52, 0x07, 0x75,
	0x72, 0x6c, 0x4d, 0x65, 0x74, 0x61, 0x12, 0x1b, 0x0a, 0x04, 0x70, 0x61, 0x74, 0x68, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x72, 0x02, 0x10, 0x01, 0x52, 0x04, 0x70,
	0x61, 0x74, 0x68, 0x12, 0x1d, 0x0a, 0x0a, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x5f, 0x6f, 0x6e, 0x6c,
	0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x4f, 0x6e,
	0x6c, 0x79, 0x32, 0x83, 0x03, 0x0a, 0x06, 0x44, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x12, 0x39, 0x0a,
	0x08, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x12, 0x15, 0x2e, 0x64, 0x66, 0x64, 0x61,
	0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x14, 0x2e, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x44, 0x6f, 0x77, 0x6e,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x30, 0x01, 0x12, 0x3a, 0x0a, 0x0d, 0x47, 0x65, 0x74, 0x50,
	0x69, 0x65, 0x63, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x73, 0x12, 0x16, 0x2e, 0x62, 0x61, 0x73, 0x65,
	0x2e, 0x50, 0x69, 0x65, 0x63, 0x65, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x11, 0x2e, 0x62, 0x61, 0x73, 0x65, 0x2e, 0x50, 0x69, 0x65, 0x63, 0x65, 0x50, 0x61,
	0x63, 0x6b, 0x65, 0x74, 0x12, 0x3d, 0x0a, 0x0b, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x48, 0x65, 0x61,
	0x6c, 0x74, 0x68, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d,
	0x70, 0x74, 0x79, 0x12, 0x3d, 0x0a, 0x08, 0x53, 0x74, 0x61, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x12,
	0x19, 0x2e, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x54,
	0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x12, 0x41, 0x0a, 0x0a, 0x49, 0x6d, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x73, 0x6b,
	0x12, 0x1b, 0x2e, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x49, 0x6d, 0x70, 0x6f,
	0x72, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e,
	0x45, 0x6d, 0x70, 0x74, 0x79, 0x12, 0x41, 0x0a, 0x0a, 0x45, 0x78, 0x70, 0x6f, 0x72, 0x74, 0x54,
	0x61, 0x73, 0x6b, 0x12, 0x1b, 0x2e, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x45,
	0x78, 0x70, 0x6f, 0x72, 0x74, 0x54, 0x61, 0x73, 0x6b, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x42, 0x26, 0x5a, 0x24, 0x64, 0x37, 0x79, 0x2e,
	0x69, 0x6f, 0x2f, 0x64, 0x72, 0x61, 0x67, 0x6f, 0x6e, 0x66, 0x6c, 0x79, 0x2f, 0x76, 0x32, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x72, 0x70, 0x63, 0x2f, 0x64, 0x66, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescOnce sync.Once
	file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescData = file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDesc
)

func file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescGZIP() []byte {
	file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescOnce.Do(func() {
		file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescData)
	})
	return file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDescData
}

var file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_pkg_rpc_dfdaemon_dfdaemon_proto_goTypes = []interface{}{
	(*DownRequest)(nil),           // 0: dfdaemon.DownRequest
	(*DownResult)(nil),            // 1: dfdaemon.DownResult
	(*StatTaskRequest)(nil),       // 2: dfdaemon.StatTaskRequest
	(*ImportTaskRequest)(nil),     // 3: dfdaemon.ImportTaskRequest
	(*ExportTaskRequest)(nil),     // 4: dfdaemon.ExportTaskRequest
	(*base.UrlMeta)(nil),          // 5: base.UrlMeta
	(*base.PieceTaskRequest)(nil), // 6: base.PieceTaskRequest
	(*emptypb.Empty)(nil),         // 7: google.protobuf.Empty
	(*base.PiecePacket)(nil),      // 8: base.PiecePacket
}
var file_pkg_rpc_dfdaemon_dfdaemon_proto_depIdxs = []int32{
	5,  // 0: dfdaemon.DownRequest.url_meta:type_name -> base.UrlMeta
	5,  // 1: dfdaemon.StatTaskRequest.url_meta:type_name -> base.UrlMeta
	5,  // 2: dfdaemon.ImportTaskRequest.url_meta:type_name -> base.UrlMeta
	5,  // 3: dfdaemon.ExportTaskRequest.url_meta:type_name -> base.UrlMeta
	0,  // 4: dfdaemon.Daemon.Download:input_type -> dfdaemon.DownRequest
	6,  // 5: dfdaemon.Daemon.GetPieceTasks:input_type -> base.PieceTaskRequest
	7,  // 6: dfdaemon.Daemon.CheckHealth:input_type -> google.protobuf.Empty
	2,  // 7: dfdaemon.Daemon.StatTask:input_type -> dfdaemon.StatTaskRequest
	3,  // 8: dfdaemon.Daemon.ImportTask:input_type -> dfdaemon.ImportTaskRequest
	4,  // 9: dfdaemon.Daemon.ExportTask:input_type -> dfdaemon.ExportTaskRequest
	1,  // 10: dfdaemon.Daemon.Download:output_type -> dfdaemon.DownResult
	8,  // 11: dfdaemon.Daemon.GetPieceTasks:output_type -> base.PiecePacket
	7,  // 12: dfdaemon.Daemon.CheckHealth:output_type -> google.protobuf.Empty
	7,  // 13: dfdaemon.Daemon.StatTask:output_type -> google.protobuf.Empty
	7,  // 14: dfdaemon.Daemon.ImportTask:output_type -> google.protobuf.Empty
	7,  // 15: dfdaemon.Daemon.ExportTask:output_type -> google.protobuf.Empty
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_pkg_rpc_dfdaemon_dfdaemon_proto_init() }
func file_pkg_rpc_dfdaemon_dfdaemon_proto_init() {
	if File_pkg_rpc_dfdaemon_dfdaemon_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StatTaskRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImportTaskRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExportTaskRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_rpc_dfdaemon_dfdaemon_proto_goTypes,
		DependencyIndexes: file_pkg_rpc_dfdaemon_dfdaemon_proto_depIdxs,
		MessageInfos:      file_pkg_rpc_dfdaemon_dfdaemon_proto_msgTypes,
	}.Build()
	File_pkg_rpc_dfdaemon_dfdaemon_proto = out.File
	file_pkg_rpc_dfdaemon_dfdaemon_proto_rawDesc = nil
	file_pkg_rpc_dfdaemon_dfdaemon_proto_goTypes = nil
	file_pkg_rpc_dfdaemon_dfdaemon_proto_depIdxs = nil
}
