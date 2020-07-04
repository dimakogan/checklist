package main

import (
	fmt "fmt"

	proto "github.com/golang/protobuf/proto"

	math "math"

	duration "github.com/golang/protobuf/ptypes/duration"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Types of threats.
type ThreatType int32

const (
	// Unknown.
	ThreatType_THREAT_TYPE_UNSPECIFIED ThreatType = 0
	// Malware threat type.
	ThreatType_MALWARE ThreatType = 1
	// Social engineering threat type.
	ThreatType_SOCIAL_ENGINEERING ThreatType = 2
	// Unwanted software threat type.
	ThreatType_UNWANTED_SOFTWARE ThreatType = 3
	// Potentially harmful application threat type.
	ThreatType_POTENTIALLY_HARMFUL_APPLICATION ThreatType = 4
)

var ThreatType_name = map[int32]string{
	0: "THREAT_TYPE_UNSPECIFIED",
	1: "MALWARE",
	2: "SOCIAL_ENGINEERING",
	3: "UNWANTED_SOFTWARE",
	4: "POTENTIALLY_HARMFUL_APPLICATION",
}
var ThreatType_value = map[string]int32{
	"THREAT_TYPE_UNSPECIFIED":         0,
	"MALWARE":                         1,
	"SOCIAL_ENGINEERING":              2,
	"UNWANTED_SOFTWARE":               3,
	"POTENTIALLY_HARMFUL_APPLICATION": 4,
}

func (x ThreatType) String() string {
	return proto.EnumName(ThreatType_name, int32(x))
}
func (ThreatType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{0}
}

// Types of platforms.
type PlatformType int32

const (
	// Unknown platform.
	PlatformType_PLATFORM_TYPE_UNSPECIFIED PlatformType = 0
	// Threat posed to Windows.
	PlatformType_WINDOWS PlatformType = 1
	// Threat posed to Linux.
	PlatformType_LINUX PlatformType = 2
	// Threat posed to Android.
	PlatformType_ANDROID PlatformType = 3
	// Threat posed to OSX.
	PlatformType_OSX PlatformType = 4
	// Threat posed to iOS.
	PlatformType_IOS PlatformType = 5
	// Threat posed to at least one of the defined platforms.
	PlatformType_ANY_PLATFORM PlatformType = 6
	// Threat posed to all defined platforms.
	PlatformType_ALL_PLATFORMS PlatformType = 7
	// Threat posed to Chrome.
	PlatformType_CHROME PlatformType = 8
)

var PlatformType_name = map[int32]string{
	0: "PLATFORM_TYPE_UNSPECIFIED",
	1: "WINDOWS",
	2: "LINUX",
	3: "ANDROID",
	4: "OSX",
	5: "IOS",
	6: "ANY_PLATFORM",
	7: "ALL_PLATFORMS",
	8: "CHROME",
}
var PlatformType_value = map[string]int32{
	"PLATFORM_TYPE_UNSPECIFIED": 0,
	"WINDOWS":                   1,
	"LINUX":                     2,
	"ANDROID":                   3,
	"OSX":                       4,
	"IOS":                       5,
	"ANY_PLATFORM":              6,
	"ALL_PLATFORMS":             7,
	"CHROME":                    8,
}

func (x PlatformType) String() string {
	return proto.EnumName(PlatformType_name, int32(x))
}
func (PlatformType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{1}
}

// The ways in which threat entry sets can be compressed.
type CompressionType int32

const (
	// Unknown.
	CompressionType_COMPRESSION_TYPE_UNSPECIFIED CompressionType = 0
	// Raw, uncompressed data.
	CompressionType_RAW CompressionType = 1
	// Rice-Golomb encoded data.
	CompressionType_RICE CompressionType = 2
)

var CompressionType_name = map[int32]string{
	0: "COMPRESSION_TYPE_UNSPECIFIED",
	1: "RAW",
	2: "RICE",
}
var CompressionType_value = map[string]int32{
	"COMPRESSION_TYPE_UNSPECIFIED": 0,
	"RAW":                          1,
	"RICE":                         2,
}

func (x CompressionType) String() string {
	return proto.EnumName(CompressionType_name, int32(x))
}
func (CompressionType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{2}
}

// Types of entries that pose threats. Threat lists are collections of entries
// of a single type.
type ThreatEntryType int32

const (
	// Unspecified.
	ThreatEntryType_THREAT_ENTRY_TYPE_UNSPECIFIED ThreatEntryType = 0
	// A URL.
	ThreatEntryType_URL ThreatEntryType = 1
	// An executable program.
	ThreatEntryType_EXECUTABLE ThreatEntryType = 2
	// An IP range.
	ThreatEntryType_IP_RANGE ThreatEntryType = 3
)

var ThreatEntryType_name = map[int32]string{
	0: "THREAT_ENTRY_TYPE_UNSPECIFIED",
	1: "URL",
	2: "EXECUTABLE",
	3: "IP_RANGE",
}
var ThreatEntryType_value = map[string]int32{
	"THREAT_ENTRY_TYPE_UNSPECIFIED": 0,
	"URL":                           1,
	"EXECUTABLE":                    2,
	"IP_RANGE":                      3,
}

func (x ThreatEntryType) String() string {
	return proto.EnumName(ThreatEntryType_name, int32(x))
}
func (ThreatEntryType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{3}
}

// The type of response sent to the client.
type FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType int32

const (
	// Unknown.
	FetchThreatListUpdatesResponse_ListUpdateResponse_RESPONSE_TYPE_UNSPECIFIED FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType = 0
	// Partial updates are applied to the client's existing local database.
	FetchThreatListUpdatesResponse_ListUpdateResponse_PARTIAL_UPDATE FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType = 1
	// Full updates replace the client's entire local database. This means
	// that either the client was seriously out-of-date or the client is
	// believed to be corrupt.
	FetchThreatListUpdatesResponse_ListUpdateResponse_FULL_UPDATE FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType = 2
)

var FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType_name = map[int32]string{
	0: "RESPONSE_TYPE_UNSPECIFIED",
	1: "PARTIAL_UPDATE",
	2: "FULL_UPDATE",
}
var FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType_value = map[string]int32{
	"RESPONSE_TYPE_UNSPECIFIED": 0,
	"PARTIAL_UPDATE":            1,
	"FULL_UPDATE":               2,
}

func (x FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType) String() string {
	return proto.EnumName(FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType_name, int32(x))
}
func (FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{5, 0, 0}
}

// The information regarding one or more threats that a client submits when
// checking for matches in threat lists.
type ThreatInfo struct {
	// The threat types to be checked.
	ThreatTypes []ThreatType `protobuf:"varint,1,rep,packed,name=threat_types,json=threatTypes,proto3,enum=safebrowsing_proto.ThreatType" json:"threat_types,omitempty"`
	// The platform types to be checked.
	PlatformTypes []PlatformType `protobuf:"varint,2,rep,packed,name=platform_types,json=platformTypes,proto3,enum=safebrowsing_proto.PlatformType" json:"platform_types,omitempty"`
	// The entry types to be checked.
	ThreatEntryTypes []ThreatEntryType `protobuf:"varint,4,rep,packed,name=threat_entry_types,json=threatEntryTypes,proto3,enum=safebrowsing_proto.ThreatEntryType" json:"threat_entry_types,omitempty"`
	// The threat entries to be checked.
	ThreatEntries        []*ThreatEntry `protobuf:"bytes,3,rep,name=threat_entries,json=threatEntries,proto3" json:"threat_entries,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *ThreatInfo) Reset()         { *m = ThreatInfo{} }
func (m *ThreatInfo) String() string { return proto.CompactTextString(m) }
func (*ThreatInfo) ProtoMessage()    {}
func (*ThreatInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{0}
}
func (m *ThreatInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatInfo.Unmarshal(m, b)
}
func (m *ThreatInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatInfo.Marshal(b, m, deterministic)
}
func (dst *ThreatInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatInfo.Merge(dst, src)
}
func (m *ThreatInfo) XXX_Size() int {
	return xxx_messageInfo_ThreatInfo.Size(m)
}
func (m *ThreatInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatInfo proto.InternalMessageInfo

func (m *ThreatInfo) GetThreatTypes() []ThreatType {
	if m != nil {
		return m.ThreatTypes
	}
	return nil
}

func (m *ThreatInfo) GetPlatformTypes() []PlatformType {
	if m != nil {
		return m.PlatformTypes
	}
	return nil
}

func (m *ThreatInfo) GetThreatEntryTypes() []ThreatEntryType {
	if m != nil {
		return m.ThreatEntryTypes
	}
	return nil
}

func (m *ThreatInfo) GetThreatEntries() []*ThreatEntry {
	if m != nil {
		return m.ThreatEntries
	}
	return nil
}

// A match when checking a threat entry in the Safe Browsing threat lists.
type ThreatMatch struct {
	// The threat type matching this threat.
	ThreatType ThreatType `protobuf:"varint,1,opt,name=threat_type,json=threatType,proto3,enum=safebrowsing_proto.ThreatType" json:"threat_type,omitempty"`
	// The platform type matching this threat.
	PlatformType PlatformType `protobuf:"varint,2,opt,name=platform_type,json=platformType,proto3,enum=safebrowsing_proto.PlatformType" json:"platform_type,omitempty"`
	// The threat entry type matching this threat.
	ThreatEntryType ThreatEntryType `protobuf:"varint,6,opt,name=threat_entry_type,json=threatEntryType,proto3,enum=safebrowsing_proto.ThreatEntryType" json:"threat_entry_type,omitempty"`
	// The threat matching this threat.
	Threat *ThreatEntry `protobuf:"bytes,3,opt,name=threat,proto3" json:"threat,omitempty"`
	// Optional metadata associated with this threat.
	ThreatEntryMetadata *ThreatEntryMetadata `protobuf:"bytes,4,opt,name=threat_entry_metadata,json=threatEntryMetadata,proto3" json:"threat_entry_metadata,omitempty"`
	// The cache lifetime for the returned match. Clients must not cache this
	// response for more than this duration to avoid false positives.
	CacheDuration        *duration.Duration `protobuf:"bytes,5,opt,name=cache_duration,json=cacheDuration,proto3" json:"cache_duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *ThreatMatch) Reset()         { *m = ThreatMatch{} }
func (m *ThreatMatch) String() string { return proto.CompactTextString(m) }
func (*ThreatMatch) ProtoMessage()    {}
func (*ThreatMatch) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{1}
}
func (m *ThreatMatch) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatMatch.Unmarshal(m, b)
}
func (m *ThreatMatch) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatMatch.Marshal(b, m, deterministic)
}
func (dst *ThreatMatch) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatMatch.Merge(dst, src)
}
func (m *ThreatMatch) XXX_Size() int {
	return xxx_messageInfo_ThreatMatch.Size(m)
}
func (m *ThreatMatch) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatMatch.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatMatch proto.InternalMessageInfo

func (m *ThreatMatch) GetThreatType() ThreatType {
	if m != nil {
		return m.ThreatType
	}
	return ThreatType_THREAT_TYPE_UNSPECIFIED
}

func (m *ThreatMatch) GetPlatformType() PlatformType {
	if m != nil {
		return m.PlatformType
	}
	return PlatformType_PLATFORM_TYPE_UNSPECIFIED
}

func (m *ThreatMatch) GetThreatEntryType() ThreatEntryType {
	if m != nil {
		return m.ThreatEntryType
	}
	return ThreatEntryType_THREAT_ENTRY_TYPE_UNSPECIFIED
}

func (m *ThreatMatch) GetThreat() *ThreatEntry {
	if m != nil {
		return m.Threat
	}
	return nil
}

func (m *ThreatMatch) GetThreatEntryMetadata() *ThreatEntryMetadata {
	if m != nil {
		return m.ThreatEntryMetadata
	}
	return nil
}

func (m *ThreatMatch) GetCacheDuration() *duration.Duration {
	if m != nil {
		return m.CacheDuration
	}
	return nil
}

// Request to check entries against lists.
type FindThreatMatchesRequest struct {
	// The client metadata.
	Client *ClientInfo `protobuf:"bytes,1,opt,name=client,proto3" json:"client,omitempty"`
	// The lists and entries to be checked for matches.
	ThreatInfo           *ThreatInfo `protobuf:"bytes,2,opt,name=threat_info,json=threatInfo,proto3" json:"threat_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *FindThreatMatchesRequest) Reset()         { *m = FindThreatMatchesRequest{} }
func (m *FindThreatMatchesRequest) String() string { return proto.CompactTextString(m) }
func (*FindThreatMatchesRequest) ProtoMessage()    {}
func (*FindThreatMatchesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{2}
}
func (m *FindThreatMatchesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindThreatMatchesRequest.Unmarshal(m, b)
}
func (m *FindThreatMatchesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindThreatMatchesRequest.Marshal(b, m, deterministic)
}
func (dst *FindThreatMatchesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindThreatMatchesRequest.Merge(dst, src)
}
func (m *FindThreatMatchesRequest) XXX_Size() int {
	return xxx_messageInfo_FindThreatMatchesRequest.Size(m)
}
func (m *FindThreatMatchesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FindThreatMatchesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FindThreatMatchesRequest proto.InternalMessageInfo

func (m *FindThreatMatchesRequest) GetClient() *ClientInfo {
	if m != nil {
		return m.Client
	}
	return nil
}

func (m *FindThreatMatchesRequest) GetThreatInfo() *ThreatInfo {
	if m != nil {
		return m.ThreatInfo
	}
	return nil
}

// Response type for requests to find threat matches.
type FindThreatMatchesResponse struct {
	// The threat list matches.
	Matches              []*ThreatMatch `protobuf:"bytes,1,rep,name=matches,proto3" json:"matches,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *FindThreatMatchesResponse) Reset()         { *m = FindThreatMatchesResponse{} }
func (m *FindThreatMatchesResponse) String() string { return proto.CompactTextString(m) }
func (*FindThreatMatchesResponse) ProtoMessage()    {}
func (*FindThreatMatchesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{3}
}
func (m *FindThreatMatchesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindThreatMatchesResponse.Unmarshal(m, b)
}
func (m *FindThreatMatchesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindThreatMatchesResponse.Marshal(b, m, deterministic)
}
func (dst *FindThreatMatchesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindThreatMatchesResponse.Merge(dst, src)
}
func (m *FindThreatMatchesResponse) XXX_Size() int {
	return xxx_messageInfo_FindThreatMatchesResponse.Size(m)
}
func (m *FindThreatMatchesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FindThreatMatchesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FindThreatMatchesResponse proto.InternalMessageInfo

func (m *FindThreatMatchesResponse) GetMatches() []*ThreatMatch {
	if m != nil {
		return m.Matches
	}
	return nil
}

// Describes a Safe Browsing API update request. Clients can request updates for
// multiple lists in a single request.
// NOTE: Field index 2 is unused.
type FetchThreatListUpdatesRequest struct {
	// The client metadata.
	Client *ClientInfo `protobuf:"bytes,1,opt,name=client,proto3" json:"client,omitempty"`
	// The requested threat list updates.
	ListUpdateRequests   []*FetchThreatListUpdatesRequest_ListUpdateRequest `protobuf:"bytes,3,rep,name=list_update_requests,json=listUpdateRequests,proto3" json:"list_update_requests,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                           `json:"-"`
	XXX_unrecognized     []byte                                             `json:"-"`
	XXX_sizecache        int32                                              `json:"-"`
}

func (m *FetchThreatListUpdatesRequest) Reset()         { *m = FetchThreatListUpdatesRequest{} }
func (m *FetchThreatListUpdatesRequest) String() string { return proto.CompactTextString(m) }
func (*FetchThreatListUpdatesRequest) ProtoMessage()    {}
func (*FetchThreatListUpdatesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{4}
}
func (m *FetchThreatListUpdatesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FetchThreatListUpdatesRequest.Unmarshal(m, b)
}
func (m *FetchThreatListUpdatesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FetchThreatListUpdatesRequest.Marshal(b, m, deterministic)
}
func (dst *FetchThreatListUpdatesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FetchThreatListUpdatesRequest.Merge(dst, src)
}
func (m *FetchThreatListUpdatesRequest) XXX_Size() int {
	return xxx_messageInfo_FetchThreatListUpdatesRequest.Size(m)
}
func (m *FetchThreatListUpdatesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FetchThreatListUpdatesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FetchThreatListUpdatesRequest proto.InternalMessageInfo

func (m *FetchThreatListUpdatesRequest) GetClient() *ClientInfo {
	if m != nil {
		return m.Client
	}
	return nil
}

func (m *FetchThreatListUpdatesRequest) GetListUpdateRequests() []*FetchThreatListUpdatesRequest_ListUpdateRequest {
	if m != nil {
		return m.ListUpdateRequests
	}
	return nil
}

// A single list update request.
type FetchThreatListUpdatesRequest_ListUpdateRequest struct {
	// The type of threat posed by entries present in the list.
	ThreatType ThreatType `protobuf:"varint,1,opt,name=threat_type,json=threatType,proto3,enum=safebrowsing_proto.ThreatType" json:"threat_type,omitempty"`
	// The type of platform at risk by entries present in the list.
	PlatformType PlatformType `protobuf:"varint,2,opt,name=platform_type,json=platformType,proto3,enum=safebrowsing_proto.PlatformType" json:"platform_type,omitempty"`
	// The types of entries present in the list.
	ThreatEntryType ThreatEntryType `protobuf:"varint,5,opt,name=threat_entry_type,json=threatEntryType,proto3,enum=safebrowsing_proto.ThreatEntryType" json:"threat_entry_type,omitempty"`
	// The current state of the client for the requested list (the encrypted
	// ClientState that was sent to the client from the previous update
	// request).
	State []byte `protobuf:"bytes,3,opt,name=state,proto3" json:"state,omitempty"`
	// The constraints associated with this request.
	Constraints          *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints `protobuf:"bytes,4,opt,name=constraints,proto3" json:"constraints,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                                     `json:"-"`
	XXX_unrecognized     []byte                                                       `json:"-"`
	XXX_sizecache        int32                                                        `json:"-"`
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) Reset() {
	*m = FetchThreatListUpdatesRequest_ListUpdateRequest{}
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) String() string {
	return proto.CompactTextString(m)
}
func (*FetchThreatListUpdatesRequest_ListUpdateRequest) ProtoMessage() {}
func (*FetchThreatListUpdatesRequest_ListUpdateRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{4, 0}
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest.Unmarshal(m, b)
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest.Marshal(b, m, deterministic)
}
func (dst *FetchThreatListUpdatesRequest_ListUpdateRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest.Merge(dst, src)
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) XXX_Size() int {
	return xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest.Size(m)
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest proto.InternalMessageInfo

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) GetThreatType() ThreatType {
	if m != nil {
		return m.ThreatType
	}
	return ThreatType_THREAT_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) GetPlatformType() PlatformType {
	if m != nil {
		return m.PlatformType
	}
	return PlatformType_PLATFORM_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) GetThreatEntryType() ThreatEntryType {
	if m != nil {
		return m.ThreatEntryType
	}
	return ThreatEntryType_THREAT_ENTRY_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) GetState() []byte {
	if m != nil {
		return m.State
	}
	return nil
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest) GetConstraints() *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints {
	if m != nil {
		return m.Constraints
	}
	return nil
}

// The constraints for this update.
type FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints struct {
	// The maximum size in number of entries. The update will not contain more
	// entries than this value.  This should be a power of 2 between 2**10 and
	// 2**20.  If zero, no update size limit is set.
	MaxUpdateEntries int32 `protobuf:"varint,1,opt,name=max_update_entries,json=maxUpdateEntries,proto3" json:"max_update_entries,omitempty"`
	// Sets the maximum number of entries that the client is willing to have
	// in the local database. This should be a power of 2 between 2**10 and
	// 2**20. If zero, no database size limit is set.
	MaxDatabaseEntries int32 `protobuf:"varint,2,opt,name=max_database_entries,json=maxDatabaseEntries,proto3" json:"max_database_entries,omitempty"`
	// Requests the list for a specific geographic location. If not set the
	// server may pick that value based on the user's IP address. Expects ISO
	// 3166-1 alpha-2 format.
	Region string `protobuf:"bytes,3,opt,name=region,proto3" json:"region,omitempty"`
	// The compression types supported by the client.
	SupportedCompressions []CompressionType `protobuf:"varint,4,rep,packed,name=supported_compressions,json=supportedCompressions,proto3,enum=safebrowsing_proto.CompressionType" json:"supported_compressions,omitempty"`
	XXX_NoUnkeyedLiteral  struct{}          `json:"-"`
	XXX_unrecognized      []byte            `json:"-"`
	XXX_sizecache         int32             `json:"-"`
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) Reset() {
	*m = FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints{}
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) String() string {
	return proto.CompactTextString(m)
}
func (*FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) ProtoMessage() {}
func (*FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{4, 0, 0}
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints.Unmarshal(m, b)
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints.Marshal(b, m, deterministic)
}
func (dst *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints.Merge(dst, src)
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) XXX_Size() int {
	return xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints.Size(m)
}
func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) XXX_DiscardUnknown() {
	xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints.DiscardUnknown(m)
}

var xxx_messageInfo_FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints proto.InternalMessageInfo

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) GetMaxUpdateEntries() int32 {
	if m != nil {
		return m.MaxUpdateEntries
	}
	return 0
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) GetMaxDatabaseEntries() int32 {
	if m != nil {
		return m.MaxDatabaseEntries
	}
	return 0
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) GetRegion() string {
	if m != nil {
		return m.Region
	}
	return ""
}

func (m *FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints) GetSupportedCompressions() []CompressionType {
	if m != nil {
		return m.SupportedCompressions
	}
	return nil
}

// Response type for threat list update requests.
type FetchThreatListUpdatesResponse struct {
	// The list updates requested by the clients.
	ListUpdateResponses []*FetchThreatListUpdatesResponse_ListUpdateResponse `protobuf:"bytes,1,rep,name=list_update_responses,json=listUpdateResponses,proto3" json:"list_update_responses,omitempty"`
	// The minimum duration the client must wait before issuing any update
	// request. If this field is not set clients may update as soon as they want.
	MinimumWaitDuration  *duration.Duration `protobuf:"bytes,2,opt,name=minimum_wait_duration,json=minimumWaitDuration,proto3" json:"minimum_wait_duration,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *FetchThreatListUpdatesResponse) Reset()         { *m = FetchThreatListUpdatesResponse{} }
func (m *FetchThreatListUpdatesResponse) String() string { return proto.CompactTextString(m) }
func (*FetchThreatListUpdatesResponse) ProtoMessage()    {}
func (*FetchThreatListUpdatesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{5}
}
func (m *FetchThreatListUpdatesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FetchThreatListUpdatesResponse.Unmarshal(m, b)
}
func (m *FetchThreatListUpdatesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FetchThreatListUpdatesResponse.Marshal(b, m, deterministic)
}
func (dst *FetchThreatListUpdatesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FetchThreatListUpdatesResponse.Merge(dst, src)
}
func (m *FetchThreatListUpdatesResponse) XXX_Size() int {
	return xxx_messageInfo_FetchThreatListUpdatesResponse.Size(m)
}
func (m *FetchThreatListUpdatesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FetchThreatListUpdatesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FetchThreatListUpdatesResponse proto.InternalMessageInfo

func (m *FetchThreatListUpdatesResponse) GetListUpdateResponses() []*FetchThreatListUpdatesResponse_ListUpdateResponse {
	if m != nil {
		return m.ListUpdateResponses
	}
	return nil
}

func (m *FetchThreatListUpdatesResponse) GetMinimumWaitDuration() *duration.Duration {
	if m != nil {
		return m.MinimumWaitDuration
	}
	return nil
}

// An update to an individual list.
type FetchThreatListUpdatesResponse_ListUpdateResponse struct {
	// The threat type for which data is returned.
	ThreatType ThreatType `protobuf:"varint,1,opt,name=threat_type,json=threatType,proto3,enum=safebrowsing_proto.ThreatType" json:"threat_type,omitempty"`
	// The format of the threats.
	ThreatEntryType ThreatEntryType `protobuf:"varint,2,opt,name=threat_entry_type,json=threatEntryType,proto3,enum=safebrowsing_proto.ThreatEntryType" json:"threat_entry_type,omitempty"`
	// The platform type for which data is returned.
	PlatformType PlatformType `protobuf:"varint,3,opt,name=platform_type,json=platformType,proto3,enum=safebrowsing_proto.PlatformType" json:"platform_type,omitempty"`
	// The type of response. This may indicate that an action is required by the
	// client when the response is received.
	ResponseType FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType `protobuf:"varint,4,opt,name=response_type,json=responseType,proto3,enum=safebrowsing_proto.FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType" json:"response_type,omitempty"`
	// A set of entries to add to a local threat type's list. Repeated to allow
	// for a combination of compressed and raw data to be sent in a single
	// response.
	Additions []*ThreatEntrySet `protobuf:"bytes,5,rep,name=additions,proto3" json:"additions,omitempty"`
	// A set of entries to remove from a local threat type's list. Repeated for
	// the same reason as above.
	Removals []*ThreatEntrySet `protobuf:"bytes,6,rep,name=removals,proto3" json:"removals,omitempty"`
	// The new client state, in encrypted format. Opaque to clients.
	NewClientState []byte `protobuf:"bytes,7,opt,name=new_client_state,json=newClientState,proto3" json:"new_client_state,omitempty"`
	// The expected SHA256 hash of the client state; that is, of the sorted list
	// of all hashes present in the database after applying the provided update.
	// If the client state doesn't match the expected state, the client must
	// disregard this update and retry later.
	Checksum             *Checksum `protobuf:"bytes,8,opt,name=checksum,proto3" json:"checksum,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) Reset() {
	*m = FetchThreatListUpdatesResponse_ListUpdateResponse{}
}
func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) String() string {
	return proto.CompactTextString(m)
}
func (*FetchThreatListUpdatesResponse_ListUpdateResponse) ProtoMessage() {}
func (*FetchThreatListUpdatesResponse_ListUpdateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{5, 0}
}
func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FetchThreatListUpdatesResponse_ListUpdateResponse.Unmarshal(m, b)
}
func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FetchThreatListUpdatesResponse_ListUpdateResponse.Marshal(b, m, deterministic)
}
func (dst *FetchThreatListUpdatesResponse_ListUpdateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FetchThreatListUpdatesResponse_ListUpdateResponse.Merge(dst, src)
}
func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) XXX_Size() int {
	return xxx_messageInfo_FetchThreatListUpdatesResponse_ListUpdateResponse.Size(m)
}
func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FetchThreatListUpdatesResponse_ListUpdateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FetchThreatListUpdatesResponse_ListUpdateResponse proto.InternalMessageInfo

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetThreatType() ThreatType {
	if m != nil {
		return m.ThreatType
	}
	return ThreatType_THREAT_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetThreatEntryType() ThreatEntryType {
	if m != nil {
		return m.ThreatEntryType
	}
	return ThreatEntryType_THREAT_ENTRY_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetPlatformType() PlatformType {
	if m != nil {
		return m.PlatformType
	}
	return PlatformType_PLATFORM_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetResponseType() FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType {
	if m != nil {
		return m.ResponseType
	}
	return FetchThreatListUpdatesResponse_ListUpdateResponse_RESPONSE_TYPE_UNSPECIFIED
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetAdditions() []*ThreatEntrySet {
	if m != nil {
		return m.Additions
	}
	return nil
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetRemovals() []*ThreatEntrySet {
	if m != nil {
		return m.Removals
	}
	return nil
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetNewClientState() []byte {
	if m != nil {
		return m.NewClientState
	}
	return nil
}

func (m *FetchThreatListUpdatesResponse_ListUpdateResponse) GetChecksum() *Checksum {
	if m != nil {
		return m.Checksum
	}
	return nil
}

// Request to return full hashes matched by the provided hash prefixes.
type FindFullHashesRequest struct {
	// The client metadata.
	Client *ClientInfo `protobuf:"bytes,1,opt,name=client,proto3" json:"client,omitempty"`
	// The current client states for each of the client's local threat lists.
	ClientStates [][]byte `protobuf:"bytes,2,rep,name=client_states,json=clientStates,proto3" json:"client_states,omitempty"`
	// The lists and hashes to be checked.
	ThreatInfo           *ThreatInfo `protobuf:"bytes,3,opt,name=threat_info,json=threatInfo,proto3" json:"threat_info,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *FindFullHashesRequest) Reset()         { *m = FindFullHashesRequest{} }
func (m *FindFullHashesRequest) String() string { return proto.CompactTextString(m) }
func (*FindFullHashesRequest) ProtoMessage()    {}
func (*FindFullHashesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{6}
}
func (m *FindFullHashesRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindFullHashesRequest.Unmarshal(m, b)
}
func (m *FindFullHashesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindFullHashesRequest.Marshal(b, m, deterministic)
}
func (dst *FindFullHashesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindFullHashesRequest.Merge(dst, src)
}
func (m *FindFullHashesRequest) XXX_Size() int {
	return xxx_messageInfo_FindFullHashesRequest.Size(m)
}
func (m *FindFullHashesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_FindFullHashesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_FindFullHashesRequest proto.InternalMessageInfo

func (m *FindFullHashesRequest) GetClient() *ClientInfo {
	if m != nil {
		return m.Client
	}
	return nil
}

func (m *FindFullHashesRequest) GetClientStates() [][]byte {
	if m != nil {
		return m.ClientStates
	}
	return nil
}

func (m *FindFullHashesRequest) GetThreactInfo() *ThreatInfo {
	if m != nil {
		return m.ThreatInfo
	}
	return nil
}

// Response type for requests to find full hashes.
type FindFullHashesResponse struct {
	// The full hashes that matched the requested prefixes.
	Matches []*ThreatMatch `protobuf:"bytes,1,rep,name=matches,proto3" json:"matches,omitempty"`
	// The minimum duration the client must wait before issuing any find hashes
	// request. If this field is not set, clients can issue a request as soon as
	// they want.
	MinimumWaitDuration *duration.Duration `protobuf:"bytes,2,opt,name=minimum_wait_duration,json=minimumWaitDuration,proto3" json:"minimum_wait_duration,omitempty"`
	// For requested entities that did not match the threat list, how long to
	// cache the response.
	NegativeCacheDuration *duration.Duration `protobuf:"bytes,3,opt,name=negative_cache_duration,json=negativeCacheDuration,proto3" json:"negative_cache_duration,omitempty"`
	XXX_NoUnkeyedLiteral  struct{}           `json:"-"`
	XXX_unrecognized      []byte             `json:"-"`
	XXX_sizecache         int32              `json:"-"`
}

func (m *FindFullHashesResponse) Reset()         { *m = FindFullHashesResponse{} }
func (m *FindFullHashesResponse) String() string { return proto.CompactTextString(m) }
func (*FindFullHashesResponse) ProtoMessage()    {}
func (*FindFullHashesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{7}
}
func (m *FindFullHashesResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FindFullHashesResponse.Unmarshal(m, b)
}
func (m *FindFullHashesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FindFullHashesResponse.Marshal(b, m, deterministic)
}
func (dst *FindFullHashesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FindFullHashesResponse.Merge(dst, src)
}
func (m *FindFullHashesResponse) XXX_Size() int {
	return xxx_messageInfo_FindFullHashesResponse.Size(m)
}
func (m *FindFullHashesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_FindFullHashesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_FindFullHashesResponse proto.InternalMessageInfo

func (m *FindFullHashesResponse) GetMatches() []*ThreatMatch {
	if m != nil {
		return m.Matches
	}
	return nil
}

func (m *FindFullHashesResponse) GetMinimumWaitDuration() *duration.Duration {
	if m != nil {
		return m.MinimumWaitDuration
	}
	return nil
}

func (m *FindFullHashesResponse) GetNegativeCacheDuration() *duration.Duration {
	if m != nil {
		return m.NegativeCacheDuration
	}
	return nil
}

// The client metadata associated with Safe Browsing API requests.
type ClientInfo struct {
	// A client ID that (hopefully) uniquely identifies the client implementation
	// of the Safe Browsing API.
	ClientId string `protobuf:"bytes,1,opt,name=client_id,json=clientId,proto3" json:"client_id,omitempty"`
	// The version of the client implementation.
	ClientVersion        string   `protobuf:"bytes,2,opt,name=client_version,json=clientVersion,proto3" json:"client_version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClientInfo) Reset()         { *m = ClientInfo{} }
func (m *ClientInfo) String() string { return proto.CompactTextString(m) }
func (*ClientInfo) ProtoMessage()    {}
func (*ClientInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{8}
}
func (m *ClientInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClientInfo.Unmarshal(m, b)
}
func (m *ClientInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClientInfo.Marshal(b, m, deterministic)
}
func (dst *ClientInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClientInfo.Merge(dst, src)
}
func (m *ClientInfo) XXX_Size() int {
	return xxx_messageInfo_ClientInfo.Size(m)
}
func (m *ClientInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ClientInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ClientInfo proto.InternalMessageInfo

func (m *ClientInfo) GetClientId() string {
	if m != nil {
		return m.ClientId
	}
	return ""
}

func (m *ClientInfo) GetClientVersion() string {
	if m != nil {
		return m.ClientVersion
	}
	return ""
}

// The expected state of a client's local database.
type Checksum struct {
	// The SHA256 hash of the client state; that is, of the sorted list of all
	// hashes present in the database.
	Sha256               []byte   `protobuf:"bytes,1,opt,name=sha256,proto3" json:"sha256,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Checksum) Reset()         { *m = Checksum{} }
func (m *Checksum) String() string { return proto.CompactTextString(m) }
func (*Checksum) ProtoMessage()    {}
func (*Checksum) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{9}
}
func (m *Checksum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Checksum.Unmarshal(m, b)
}
func (m *Checksum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Checksum.Marshal(b, m, deterministic)
}
func (dst *Checksum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Checksum.Merge(dst, src)
}
func (m *Checksum) XXX_Size() int {
	return xxx_messageInfo_Checksum.Size(m)
}
func (m *Checksum) XXX_DiscardUnknown() {
	xxx_messageInfo_Checksum.DiscardUnknown(m)
}

var xxx_messageInfo_Checksum proto.InternalMessageInfo

func (m *Checksum) GetSha256() []byte {
	if m != nil {
		return m.Sha256
	}
	return nil
}

// An individual threat; for example, a malicious URL or its hash
// representation. Only one of these fields should be set.
type ThreatEntry struct {
	// A hash prefix, consisting of the most significant 4-32 bytes of a SHA256
	// hash.
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
	// A URL.
	Url                  string   `protobuf:"bytes,2,opt,name=url,proto3" json:"url,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ThreatEntry) Reset()         { *m = ThreatEntry{} }
func (m *ThreatEntry) String() string { return proto.CompactTextString(m) }
func (*ThreatEntry) ProtoMessage()    {}
func (*ThreatEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{10}
}
func (m *ThreatEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatEntry.Unmarshal(m, b)
}
func (m *ThreatEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatEntry.Marshal(b, m, deterministic)
}
func (dst *ThreatEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatEntry.Merge(dst, src)
}
func (m *ThreatEntry) XXX_Size() int {
	return xxx_messageInfo_ThreatEntry.Size(m)
}
func (m *ThreatEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatEntry.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatEntry proto.InternalMessageInfo

func (m *ThreatEntry) GetHash() []byte {
	if m != nil {
		return m.Hash
	}
	return nil
}

func (m *ThreatEntry) GetUrl() string {
	if m != nil {
		return m.Url
	}
	return ""
}

// A set of threats that should be added or removed from a client's local
// database.
type ThreatEntrySet struct {
	// The compression type for the entries in this set.
	CompressionType CompressionType `protobuf:"varint,1,opt,name=compression_type,json=compressionType,proto3,enum=safebrowsing_proto.CompressionType" json:"compression_type,omitempty"`
	// The raw SHA256-formatted entries.
	RawHashes *RawHashes `protobuf:"bytes,2,opt,name=raw_hashes,json=rawHashes,proto3" json:"raw_hashes,omitempty"`
	// The raw removal indices for a local list.
	RawIndices *RawIndices `protobuf:"bytes,3,opt,name=raw_indices,json=rawIndices,proto3" json:"raw_indices,omitempty"`
	// The encoded 4-byte prefixes of SHA256-formatted entries, using a
	// Golomb-Rice encoding.
	RiceHashes *RiceDeltaEncoding `protobuf:"bytes,4,opt,name=rice_hashes,json=riceHashes,proto3" json:"rice_hashes,omitempty"`
	// The encoded local, lexicographically-sorted list indices, using a
	// Golomb-Rice encoding. Used for sending compressed removal indices.
	RiceIndices          *RiceDeltaEncoding `protobuf:"bytes,5,opt,name=rice_indices,json=riceIndices,proto3" json:"rice_indices,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *ThreatEntrySet) Reset()         { *m = ThreatEntrySet{} }
func (m *ThreatEntrySet) String() string { return proto.CompactTextString(m) }
func (*ThreatEntrySet) ProtoMessage()    {}
func (*ThreatEntrySet) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{11}
}
func (m *ThreatEntrySet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatEntrySet.Unmarshal(m, b)
}
func (m *ThreatEntrySet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatEntrySet.Marshal(b, m, deterministic)
}
func (dst *ThreatEntrySet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatEntrySet.Merge(dst, src)
}
func (m *ThreatEntrySet) XXX_Size() int {
	return xxx_messageInfo_ThreatEntrySet.Size(m)
}
func (m *ThreatEntrySet) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatEntrySet.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatEntrySet proto.InternalMessageInfo

func (m *ThreatEntrySet) GetCompressionType() CompressionType {
	if m != nil {
		return m.CompressionType
	}
	return CompressionType_COMPRESSION_TYPE_UNSPECIFIED
}

func (m *ThreatEntrySet) GetRawHashes() *RawHashes {
	if m != nil {
		return m.RawHashes
	}
	return nil
}

func (m *ThreatEntrySet) GetRawIndices() *RawIndices {
	if m != nil {
		return m.RawIndices
	}
	return nil
}

func (m *ThreatEntrySet) GetRiceHashes() *RiceDeltaEncoding {
	if m != nil {
		return m.RiceHashes
	}
	return nil
}

func (m *ThreatEntrySet) GetRiceIndices() *RiceDeltaEncoding {
	if m != nil {
		return m.RiceIndices
	}
	return nil
}

// A set of raw indices to remove from a local list.
type RawIndices struct {
	// The indices to remove from a lexicographically-sorted local list.
	Indices              []int32  `protobuf:"varint,1,rep,packed,name=indices,proto3" json:"indices,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RawIndices) Reset()         { *m = RawIndices{} }
func (m *RawIndices) String() string { return proto.CompactTextString(m) }
func (*RawIndices) ProtoMessage()    {}
func (*RawIndices) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{12}
}
func (m *RawIndices) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RawIndices.Unmarshal(m, b)
}
func (m *RawIndices) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RawIndices.Marshal(b, m, deterministic)
}
func (dst *RawIndices) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RawIndices.Merge(dst, src)
}
func (m *RawIndices) XXX_Size() int {
	return xxx_messageInfo_RawIndices.Size(m)
}
func (m *RawIndices) XXX_DiscardUnknown() {
	xxx_messageInfo_RawIndices.DiscardUnknown(m)
}

var xxx_messageInfo_RawIndices proto.InternalMessageInfo

func (m *RawIndices) GetIndices() []int32 {
	if m != nil {
		return m.Indices
	}
	return nil
}

// The uncompressed threat entries in hash format of a particular prefix length.
// Hashes can be anywhere from 4 to 32 bytes in size. A large majority are 4
// bytes, but some hashes are lengthened if they collide with the hash of a
// popular URL.
//
// Used for sending ThreatEntrySet to clients that do not support compression,
// or when sending non-4-byte hashes to clients that do support compression.
type RawHashes struct {
	// The number of bytes for each prefix encoded below.  This field can be
	// anywhere from 4 (shortest prefix) to 32 (full SHA256 hash).
	PrefixSize int32 `protobuf:"varint,1,opt,name=prefix_size,json=prefixSize,proto3" json:"prefix_size,omitempty"`
	// The hashes, all concatenated into one long string.  Each hash has a prefix
	// size of |prefix_size| above. Hashes are sorted in lexicographic order.
	RawHashes            []byte   `protobuf:"bytes,2,opt,name=raw_hashes,json=rawHashes,proto3" json:"raw_hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RawHashes) Reset()         { *m = RawHashes{} }
func (m *RawHashes) String() string { return proto.CompactTextString(m) }
func (*RawHashes) ProtoMessage()    {}
func (*RawHashes) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{13}
}
func (m *RawHashes) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RawHashes.Unmarshal(m, b)
}
func (m *RawHashes) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RawHashes.Marshal(b, m, deterministic)
}
func (dst *RawHashes) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RawHashes.Merge(dst, src)
}
func (m *RawHashes) XXX_Size() int {
	return xxx_messageInfo_RawHashes.Size(m)
}
func (m *RawHashes) XXX_DiscardUnknown() {
	xxx_messageInfo_RawHashes.DiscardUnknown(m)
}

var xxx_messageInfo_RawHashes proto.InternalMessageInfo

func (m *RawHashes) GetPrefixSize() int32 {
	if m != nil {
		return m.PrefixSize
	}
	return 0
}

func (m *RawHashes) GetRawHashes() []byte {
	if m != nil {
		return m.RawHashes
	}
	return nil
}

// The Rice-Golomb encoded data. Used for sending compressed 4-byte hashes or
// compressed removal indices.
type RiceDeltaEncoding struct {
	// The offset of the first entry in the encoded data, or, if only a single
	// integer was encoded, that single integer's value.
	FirstValue int64 `protobuf:"varint,1,opt,name=first_value,json=firstValue,proto3" json:"first_value,omitempty"`
	// The Golomb-Rice parameter which is a number between 2 and 28. This field
	// is missing (that is, zero) if num_entries is zero.
	RiceParameter int32 `protobuf:"varint,2,opt,name=rice_parameter,json=riceParameter,proto3" json:"rice_parameter,omitempty"`
	// The number of entries that are delta encoded in the encoded data. If only a
	// single integer was encoded, this will be zero and the single value will be
	// stored in first_value.
	NumEntries int32 `protobuf:"varint,3,opt,name=num_entries,json=numEntries,proto3" json:"num_entries,omitempty"`
	// The encoded deltas that are encoded using the Golomb-Rice coder.
	EncodedData          []byte   `protobuf:"bytes,4,opt,name=encoded_data,json=encodedData,proto3" json:"encoded_data,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RiceDeltaEncoding) Reset()         { *m = RiceDeltaEncoding{} }
func (m *RiceDeltaEncoding) String() string { return proto.CompactTextString(m) }
func (*RiceDeltaEncoding) ProtoMessage()    {}
func (*RiceDeltaEncoding) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{14}
}
func (m *RiceDeltaEncoding) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RiceDeltaEncoding.Unmarshal(m, b)
}
func (m *RiceDeltaEncoding) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RiceDeltaEncoding.Marshal(b, m, deterministic)
}
func (dst *RiceDeltaEncoding) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RiceDeltaEncoding.Merge(dst, src)
}
func (m *RiceDeltaEncoding) XXX_Size() int {
	return xxx_messageInfo_RiceDeltaEncoding.Size(m)
}
func (m *RiceDeltaEncoding) XXX_DiscardUnknown() {
	xxx_messageInfo_RiceDeltaEncoding.DiscardUnknown(m)
}

var xxx_messageInfo_RiceDeltaEncoding proto.InternalMessageInfo

func (m *RiceDeltaEncoding) GetFirstValue() int64 {
	if m != nil {
		return m.FirstValue
	}
	return 0
}

func (m *RiceDeltaEncoding) GetRiceParameter() int32 {
	if m != nil {
		return m.RiceParameter
	}
	return 0
}

func (m *RiceDeltaEncoding) GetNumEntries() int32 {
	if m != nil {
		return m.NumEntries
	}
	return 0
}

func (m *RiceDeltaEncoding) GetEncodedData() []byte {
	if m != nil {
		return m.EncodedData
	}
	return nil
}

// The metadata associated with a specific threat entry. The client is expected
// to know the metadata key/value pairs associated with each threat type.
type ThreatEntryMetadata struct {
	// The metadata entries.
	Entries              []*ThreatEntryMetadata_MetadataEntry `protobuf:"bytes,1,rep,name=entries,proto3" json:"entries,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
}

func (m *ThreatEntryMetadata) Reset()         { *m = ThreatEntryMetadata{} }
func (m *ThreatEntryMetadata) String() string { return proto.CompactTextString(m) }
func (*ThreatEntryMetadata) ProtoMessage()    {}
func (*ThreatEntryMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{15}
}
func (m *ThreatEntryMetadata) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatEntryMetadata.Unmarshal(m, b)
}
func (m *ThreatEntryMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatEntryMetadata.Marshal(b, m, deterministic)
}
func (dst *ThreatEntryMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatEntryMetadata.Merge(dst, src)
}
func (m *ThreatEntryMetadata) XXX_Size() int {
	return xxx_messageInfo_ThreatEntryMetadata.Size(m)
}
func (m *ThreatEntryMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatEntryMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatEntryMetadata proto.InternalMessageInfo

func (m *ThreatEntryMetadata) GetEntries() []*ThreatEntryMetadata_MetadataEntry {
	if m != nil {
		return m.Entries
	}
	return nil
}

// A single metadata entry.
type ThreatEntryMetadata_MetadataEntry struct {
	// The metadata entry key.
	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	// The metadata entry value.
	Value                []byte   `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ThreatEntryMetadata_MetadataEntry) Reset()         { *m = ThreatEntryMetadata_MetadataEntry{} }
func (m *ThreatEntryMetadata_MetadataEntry) String() string { return proto.CompactTextString(m) }
func (*ThreatEntryMetadata_MetadataEntry) ProtoMessage()    {}
func (*ThreatEntryMetadata_MetadataEntry) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{15, 0}
}
func (m *ThreatEntryMetadata_MetadataEntry) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatEntryMetadata_MetadataEntry.Unmarshal(m, b)
}
func (m *ThreatEntryMetadata_MetadataEntry) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatEntryMetadata_MetadataEntry.Marshal(b, m, deterministic)
}
func (dst *ThreatEntryMetadata_MetadataEntry) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatEntryMetadata_MetadataEntry.Merge(dst, src)
}
func (m *ThreatEntryMetadata_MetadataEntry) XXX_Size() int {
	return xxx_messageInfo_ThreatEntryMetadata_MetadataEntry.Size(m)
}
func (m *ThreatEntryMetadata_MetadataEntry) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatEntryMetadata_MetadataEntry.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatEntryMetadata_MetadataEntry proto.InternalMessageInfo

func (m *ThreatEntryMetadata_MetadataEntry) GetKey() []byte {
	if m != nil {
		return m.Key
	}
	return nil
}

func (m *ThreatEntryMetadata_MetadataEntry) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

// Describes an individual threat list. A list is defined by three parameters:
// the type of threat posed, the type of platform targeted by the threat, and
// the type of entries in the list.
type ThreatListDescriptor struct {
	// The threat type posed by the list's entries.
	ThreatType ThreatType `protobuf:"varint,1,opt,name=threat_type,json=threatType,proto3,enum=safebrowsing_proto.ThreatType" json:"threat_type,omitempty"`
	// The platform type targeted by the list's entries.
	PlatformType PlatformType `protobuf:"varint,2,opt,name=platform_type,json=platformType,proto3,enum=safebrowsing_proto.PlatformType" json:"platform_type,omitempty"`
	// The entry types contained in the list.
	ThreatEntryType      ThreatEntryType `protobuf:"varint,3,opt,name=threat_entry_type,json=threatEntryType,proto3,enum=safebrowsing_proto.ThreatEntryType" json:"threat_entry_type,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ThreatListDescriptor) Reset()         { *m = ThreatListDescriptor{} }
func (m *ThreatListDescriptor) String() string { return proto.CompactTextString(m) }
func (*ThreatListDescriptor) ProtoMessage()    {}
func (*ThreatListDescriptor) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{16}
}
func (m *ThreatListDescriptor) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ThreatListDescriptor.Unmarshal(m, b)
}
func (m *ThreatListDescriptor) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ThreatListDescriptor.Marshal(b, m, deterministic)
}
func (dst *ThreatListDescriptor) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ThreatListDescriptor.Merge(dst, src)
}
func (m *ThreatListDescriptor) XXX_Size() int {
	return xxx_messageInfo_ThreatListDescriptor.Size(m)
}
func (m *ThreatListDescriptor) XXX_DiscardUnknown() {
	xxx_messageInfo_ThreatListDescriptor.DiscardUnknown(m)
}

var xxx_messageInfo_ThreatListDescriptor proto.InternalMessageInfo

func (m *ThreatListDescriptor) GetThreatType() ThreatType {
	if m != nil {
		return m.ThreatType
	}
	return ThreatType_THREAT_TYPE_UNSPECIFIED
}

func (m *ThreatListDescriptor) GetPlatformType() PlatformType {
	if m != nil {
		return m.PlatformType
	}
	return PlatformType_PLATFORM_TYPE_UNSPECIFIED
}

func (m *ThreatListDescriptor) GetThreatEntryType() ThreatEntryType {
	if m != nil {
		return m.ThreatEntryType
	}
	return ThreatEntryType_THREAT_ENTRY_TYPE_UNSPECIFIED
}

// A collection of lists available for download by the client.
type ListThreatListsResponse struct {
	// The lists available for download by the client.
	ThreatLists          []*ThreatListDescriptor `protobuf:"bytes,1,rep,name=threat_lists,json=threatLists,proto3" json:"threat_lists,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *ListThreatListsResponse) Reset()         { *m = ListThreatListsResponse{} }
func (m *ListThreatListsResponse) String() string { return proto.CompactTextString(m) }
func (*ListThreatListsResponse) ProtoMessage()    {}
func (*ListThreatListsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_safebrowsing_1402a7ed618646ca, []int{17}
}
func (m *ListThreatListsResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ListThreatListsResponse.Unmarshal(m, b)
}
func (m *ListThreatListsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ListThreatListsResponse.Marshal(b, m, deterministic)
}
func (dst *ListThreatListsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ListThreatListsResponse.Merge(dst, src)
}
func (m *ListThreatListsResponse) XXX_Size() int {
	return xxx_messageInfo_ListThreatListsResponse.Size(m)
}
func (m *ListThreatListsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_ListThreatListsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_ListThreatListsResponse proto.InternalMessageInfo

func (m *ListThreatListsResponse) GetThreatLists() []*ThreatListDescriptor {
	if m != nil {
		return m.ThreatLists
	}
	return nil
}

func init() {
	/*
		proto.RegisterType((*ThreatInfo)(nil), "safebrowsing_proto.ThreatInfo")
		proto.RegisterType((*ThreatMatch)(nil), "safebrowsing_proto.ThreatMatch")
		proto.RegisterType((*FindThreatMatchesRequest)(nil), "safebrowsing_proto.FindThreatMatchesRequest")
		proto.RegisterType((*FindThreatMatchesResponse)(nil), "safebrowsing_proto.FindThreatMatchesResponse")
		proto.RegisterType((*FetchThreatListUpdatesRequest)(nil), "safebrowsing_proto.FetchThreatListUpdatesRequest")
		proto.RegisterType((*FetchThreatListUpdatesRequest_ListUpdateRequest)(nil), "safebrowsing_proto.FetchThreatListUpdatesRequest.ListUpdateRequest")
		proto.RegisterType((*FetchThreatListUpdatesRequest_ListUpdateRequest_Constraints)(nil), "safebrowsing_proto.FetchThreatListUpdatesRequest.ListUpdateRequest.Constraints")
		proto.RegisterType((*FetchThreatListUpdatesResponse)(nil), "safebrowsing_proto.FetchThreatListUpdatesResponse")
		proto.RegisterType((*FetchThreatListUpdatesResponse_ListUpdateResponse)(nil), "safebrowsing_proto.FetchThreatListUpdatesResponse.ListUpdateResponse")
		proto.RegisterType((*FindFullHashesRequest)(nil), "safebrowsing_proto.FindFullHashesRequest")
		proto.RegisterType((*FindFullHashesResponse)(nil), "safebrowsing_proto.FindFullHashesResponse")
		proto.RegisterType((*ClientInfo)(nil), "safebrowsing_proto.ClientInfo")
		proto.RegisterType((*Checksum)(nil), "safebrowsing_proto.Checksum")
		proto.RegisterType((*ThreatEntry)(nil), "safebrowsing_proto.ThreatEntry")
		proto.RegisterType((*ThreatEntrySet)(nil), "safebrowsing_proto.ThreatEntrySet")
		proto.RegisterType((*RawIndices)(nil), "safebrowsing_proto.RawIndices")
		proto.RegisterType((*RawHashes)(nil), "safebrowsing_proto.RawHashes")
		proto.RegisterType((*RiceDeltaEncoding)(nil), "safebrowsing_proto.RiceDeltaEncoding")
		proto.RegisterType((*ThreatEntryMetadata)(nil), "safebrowsing_proto.ThreatEntryMetadata")
		proto.RegisterType((*ThreatEntryMetadata_MetadataEntry)(nil), "safebrowsing_proto.ThreatEntryMetadata.MetadataEntry")
		proto.RegisterType((*ThreatListDescriptor)(nil), "safebrowsing_proto.ThreatListDescriptor")
		proto.RegisterType((*ListThreatListsResponse)(nil), "safebrowsing_proto.ListThreatListsResponse")
		proto.RegisterEnum("safebrowsing_proto.ThreatType", ThreatType_name, ThreatType_value)
		proto.RegisterEnum("safebrowsing_proto.PlatformType", PlatformType_name, PlatformType_value)
		proto.RegisterEnum("safebrowsing_proto.CompressionType", CompressionType_name, CompressionType_value)
		proto.RegisterEnum("safebrowsing_proto.ThreatEntryType", ThreatEntryType_name, ThreatEntryType_value)
		proto.RegisterEnum("safebrowsing_proto.FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType", FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType_name, FetchThreatListUpdatesResponse_ListUpdateResponse_ResponseType_value)
	*/
}

func init() { /*proto.RegisterFile("safebrowsing.proto", fileDescriptor_safebrowsing_1402a7ed618646ca) */
}

var fileDescriptor_safebrowsing_1402a7ed618646ca = []byte{
	// 1636 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x58, 0x4f, 0x8f, 0xdb, 0xc6,
	0x15, 0x2f, 0xc5, 0x95, 0xb4, 0xfb, 0xf4, 0x67, 0xb9, 0x63, 0xaf, 0x2d, 0x6f, 0x62, 0x7b, 0x43,
	0x23, 0xad, 0x61, 0x14, 0x4a, 0xe1, 0x20, 0x49, 0x0b, 0x14, 0x6d, 0x58, 0x89, 0xf2, 0x12, 0xd1,
	0x92, 0xca, 0x48, 0xf2, 0xda, 0xcd, 0x81, 0xa0, 0xa9, 0xd1, 0x8a, 0x88, 0x44, 0x2a, 0x9c, 0x91,
	0xe5, 0xcd, 0x47, 0x28, 0xd0, 0x53, 0xd1, 0x73, 0x7b, 0xee, 0xb5, 0xe8, 0xb5, 0xdf, 0xa3, 0xc7,
	0x7e, 0x82, 0xa2, 0x87, 0xde, 0x8b, 0x19, 0x0e, 0x25, 0xea, 0xcf, 0x7a, 0xd7, 0x59, 0x1f, 0x72,
	0x9b, 0x79, 0x7c, 0xf3, 0x9b, 0xc7, 0xdf, 0xfb, 0x4b, 0x02, 0xa2, 0xde, 0x90, 0xbc, 0x8a, 0xa3,
	0x39, 0x0d, 0xc2, 0xf3, 0xfa, 0x34, 0x8e, 0x58, 0x84, 0x56, 0x64, 0xae, 0x90, 0x1d, 0x3d, 0x38,
	0x8f, 0xa2, 0xf3, 0x31, 0xf9, 0x44, 0xec, 0x5e, 0xcd, 0x86, 0x9f, 0x0c, 0x66, 0xb1, 0xc7, 0x82,
	0x28, 0x4c, 0xce, 0xe8, 0x7f, 0xcf, 0x01, 0xf4, 0x46, 0x31, 0xf1, 0x98, 0x15, 0x0e, 0x23, 0x64,
	0x40, 0x99, 0x89, 0x9d, 0xcb, 0x2e, 0xa6, 0x84, 0xd6, 0x94, 0x63, 0xf5, 0x71, 0xf5, 0xe9, 0x83,
	0xfa, 0x26, 0x72, 0x3d, 0x39, 0xd5, 0xbb, 0x98, 0x12, 0x5c, 0x62, 0x8b, 0x35, 0x45, 0xcf, 0xa0,
	0x3a, 0x1d, 0x7b, 0x6c, 0x18, 0xc5, 0x13, 0x09, 0x92, 0x13, 0x20, 0xc7, 0xdb, 0x40, 0x3a, 0x52,
	0x53, 0xc0, 0x54, 0xa6, 0x99, 0x1d, 0x45, 0x5f, 0x03, 0x92, 0xb6, 0x90, 0x90, 0xc5, 0x17, 0x12,
	0x6c, 0x47, 0x80, 0x3d, 0xba, 0xdc, 0x22, 0x93, 0x2b, 0x0b, 0x3c, 0x8d, 0xad, 0x0a, 0x28, 0x6a,
	0x41, 0x35, 0x03, 0x19, 0x10, 0x5a, 0x53, 0x8f, 0xd5, 0xc7, 0xa5, 0xa7, 0x0f, 0xaf, 0x80, 0xc3,
	0x95, 0x25, 0x54, 0x40, 0xa8, 0xfe, 0x4f, 0x15, 0x4a, 0xc9, 0xe3, 0x53, 0x8f, 0xf9, 0x23, 0xf4,
	0x5b, 0x28, 0x65, 0x68, 0xab, 0x29, 0xc7, 0xca, 0x35, 0x58, 0x83, 0x25, 0x6b, 0xc8, 0x84, 0xca,
	0x0a, 0x69, 0xb5, 0x9c, 0x80, 0xb8, 0x9a, 0xb3, 0x72, 0x96, 0x33, 0xe4, 0xc0, 0xc1, 0x06, 0x65,
	0xb5, 0x82, 0x80, 0xba, 0x16, 0x63, 0xfb, 0x6b, 0x8c, 0xa1, 0x2f, 0xa0, 0x90, 0x88, 0x6a, 0xea,
	0xb1, 0x72, 0x1d, 0xa2, 0xa4, 0x3a, 0xfa, 0x06, 0x0e, 0x57, 0x2c, 0x99, 0x10, 0xe6, 0x0d, 0x3c,
	0xe6, 0xd5, 0x76, 0x04, 0xce, 0xcf, 0xae, 0xc0, 0x39, 0x95, 0xea, 0xf8, 0x16, 0xdb, 0x14, 0xa2,
	0x2f, 0xa1, 0xea, 0x7b, 0xfe, 0x88, 0xb8, 0x69, 0x30, 0xd7, 0xf2, 0x02, 0xf5, 0x5e, 0x3d, 0x89,
	0xf6, 0x7a, 0x1a, 0xed, 0xf5, 0xa6, 0x54, 0xc0, 0x15, 0x71, 0x20, 0xdd, 0xea, 0x7f, 0x52, 0xa0,
	0xd6, 0x0a, 0xc2, 0x41, 0xc6, 0x89, 0x84, 0x62, 0xf2, 0xdd, 0x8c, 0x50, 0x86, 0x3e, 0x87, 0x82,
	0x3f, 0x0e, 0x48, 0xc8, 0x84, 0x23, 0x4b, 0xdb, 0x1d, 0xd9, 0x10, 0x1a, 0x3c, 0x69, 0xb0, 0xd4,
	0xce, 0x44, 0x41, 0x10, 0x0e, 0x23, 0xe1, 0xc2, 0xd2, 0xdb, 0xa2, 0x40, 0x1c, 0x96, 0x51, 0xc0,
	0xd7, 0xfa, 0x73, 0xb8, 0xb7, 0xc5, 0x28, 0x3a, 0x8d, 0x42, 0x4a, 0xd0, 0xaf, 0xa0, 0x38, 0x49,
	0x44, 0x22, 0x2b, 0xdf, 0xea, 0x0b, 0x71, 0x16, 0xa7, 0xfa, 0xfa, 0xdf, 0x0a, 0x70, 0xbf, 0x45,
	0x98, 0x3f, 0x4a, 0x9e, 0xb6, 0x03, 0xca, 0xfa, 0xd3, 0x81, 0xc7, 0x6e, 0xfe, 0xca, 0x33, 0xb8,
	0x3d, 0x0e, 0x28, 0x73, 0x67, 0x02, 0xce, 0x8d, 0x13, 0xb8, 0x34, 0xad, 0x1a, 0xdb, 0x50, 0xde,
	0x6a, 0x48, 0x7d, 0x29, 0x92, 0x12, 0x8c, 0xc6, 0xeb, 0x22, 0x7a, 0xf4, 0xaf, 0x1d, 0x38, 0xd8,
	0xd0, 0xfc, 0x71, 0x67, 0x61, 0xfe, 0x06, 0x59, 0x78, 0x1b, 0xf2, 0x94, 0x79, 0x8c, 0x88, 0x24,
	0x2c, 0xe3, 0x64, 0x83, 0xbe, 0x83, 0x92, 0x1f, 0x85, 0x94, 0xc5, 0x5e, 0x10, 0x32, 0x2a, 0x13,
	0xcb, 0x79, 0x0f, 0x94, 0xd7, 0x1b, 0x4b, 0x58, 0x9c, 0xbd, 0xe3, 0xe8, 0xdf, 0x0a, 0x94, 0x32,
	0x0f, 0xd1, 0xcf, 0x01, 0x4d, 0xbc, 0x37, 0xa9, 0xf7, 0xd3, 0x9a, 0xca, 0x89, 0xcf, 0x63, 0x6d,
	0xe2, 0xbd, 0x49, 0x60, 0x65, 0xd5, 0x44, 0xbf, 0x80, 0xdb, 0x5c, 0x9b, 0xa7, 0xf0, 0x2b, 0x8f,
	0x2e, 0xf5, 0x73, 0x42, 0x9f, 0x23, 0x35, 0xe5, 0xa3, 0xf4, 0xc4, 0x1d, 0x28, 0xc4, 0xe4, 0x9c,
	0x27, 0x38, 0x7f, 0xf3, 0x3d, 0x2c, 0x77, 0xe8, 0xf7, 0x70, 0x87, 0xce, 0xa6, 0xd3, 0x28, 0x66,
	0x64, 0xe0, 0xfa, 0xd1, 0x64, 0x1a, 0x13, 0x4a, 0x83, 0x28, 0x7c, 0x6b, 0x7b, 0x68, 0x2c, 0xf5,
	0x04, 0xcd, 0x87, 0x0b, 0x88, 0xcc, 0x13, 0xaa, 0xff, 0xb1, 0x08, 0x0f, 0x2e, 0x23, 0x4c, 0xa6,
	0xe2, 0x05, 0x1c, 0xae, 0x46, 0x7d, 0x22, 0x4f, 0x13, 0xd3, 0x7c, 0x17, 0x1f, 0x24, 0x47, 0x57,
	0x9c, 0x90, 0x88, 0xf0, 0xad, 0xf1, 0x86, 0x8c, 0xa2, 0x53, 0x38, 0x9c, 0x04, 0x61, 0x30, 0x99,
	0x4d, 0xdc, 0xb9, 0x17, 0xb0, 0x65, 0x05, 0xcc, 0x5d, 0x55, 0x01, 0x6f, 0xc9, 0x73, 0x67, 0x5e,
	0xc0, 0x52, 0xe1, 0xd1, 0x5f, 0xf3, 0x80, 0x36, 0xaf, 0xbe, 0x79, 0x26, 0x6d, 0x4d, 0x81, 0xdc,
	0x0d, 0x52, 0x60, 0x23, 0x35, 0xd5, 0x1f, 0x94, 0x9a, 0x73, 0xa8, 0xa4, 0xde, 0x4a, 0x60, 0x76,
	0x04, 0x0c, 0x7e, 0x2f, 0x1e, 0xab, 0xa7, 0x8b, 0xe4, 0xe2, 0x38, 0xb3, 0x43, 0x5f, 0xc2, 0x9e,
	0x37, 0x18, 0x04, 0x4c, 0x04, 0x69, 0x5e, 0x84, 0x89, 0x7e, 0x05, 0x11, 0x5d, 0xc2, 0xf0, 0xf2,
	0x10, 0xfa, 0x0d, 0xec, 0xc6, 0x64, 0x12, 0xbd, 0xf6, 0xc6, 0xb4, 0x56, 0xb8, 0x36, 0xc0, 0xe2,
	0x0c, 0x7a, 0x0c, 0x5a, 0x48, 0xe6, 0x6e, 0x52, 0xb8, 0xdd, 0xa4, 0x9e, 0x14, 0x45, 0x3d, 0xa9,
	0x86, 0x64, 0x9e, 0xd4, 0xf6, 0xae, 0x28, 0x2c, 0xbf, 0x84, 0x5d, 0x7f, 0x44, 0xfc, 0x6f, 0xe9,
	0x6c, 0x52, 0xdb, 0x15, 0x61, 0xf5, 0xe1, 0xd6, 0x7c, 0x92, 0x3a, 0x78, 0xa1, 0xad, 0x63, 0x28,
	0x67, 0x39, 0x40, 0xf7, 0xe1, 0x1e, 0x36, 0xbb, 0x1d, 0xc7, 0xee, 0x9a, 0x6e, 0xef, 0x65, 0xc7,
	0x74, 0xfb, 0x76, 0xb7, 0x63, 0x36, 0xac, 0x96, 0x65, 0x36, 0xb5, 0x9f, 0x20, 0x04, 0xd5, 0x8e,
	0x81, 0x7b, 0x96, 0xd1, 0x76, 0xfb, 0x9d, 0xa6, 0xd1, 0x33, 0x35, 0x05, 0xed, 0x43, 0xa9, 0xd5,
	0x6f, 0x2f, 0x04, 0x39, 0xfd, 0x1f, 0x0a, 0x1c, 0xf2, 0xae, 0xd8, 0x9a, 0x8d, 0xc7, 0x27, 0x1e,
	0x7d, 0x0f, 0x7d, 0xfa, 0x11, 0x54, 0xb2, 0x2c, 0x24, 0x03, 0x6a, 0x19, 0x97, 0xfd, 0x25, 0x07,
	0x74, 0xbd, 0x99, 0xab, 0xef, 0xdc, 0xcc, 0xff, 0xa7, 0xc0, 0x9d, 0x75, 0xbb, 0x6f, 0xdc, 0xca,
	0xdf, 0x73, 0xfe, 0xa3, 0xaf, 0xe1, 0x6e, 0x48, 0xce, 0x3d, 0x16, 0xbc, 0x26, 0xee, 0xda, 0x48,
	0xa5, 0x5e, 0x05, 0x78, 0x98, 0x9e, 0x6c, 0xac, 0x8c, 0x56, 0x1d, 0x80, 0x25, 0xe7, 0xe8, 0x03,
	0xd8, 0x93, 0x5c, 0x07, 0x03, 0xe1, 0xa6, 0x3d, 0xbc, 0x9b, 0x08, 0xac, 0x01, 0xfa, 0x18, 0xaa,
	0xf2, 0xe1, 0x6b, 0x12, 0xd3, 0xf4, 0x2d, 0xf6, 0xb0, 0x74, 0xcf, 0xf3, 0x44, 0xa8, 0xeb, 0xb0,
	0x9b, 0xc6, 0x1a, 0xef, 0x08, 0x74, 0xe4, 0x3d, 0xfd, 0xec, 0x73, 0x01, 0x56, 0xc6, 0x72, 0xa7,
	0x7f, 0x9a, 0x0e, 0xe4, 0x22, 0xf2, 0x11, 0x82, 0x9d, 0x91, 0x47, 0x47, 0x52, 0x49, 0xac, 0x91,
	0x06, 0xea, 0x2c, 0x1e, 0xcb, 0x2b, 0xf8, 0x52, 0xff, 0x6f, 0x0e, 0xaa, 0xab, 0xf9, 0x82, 0x6c,
	0xd0, 0x32, 0xfd, 0x24, 0x5b, 0xfe, 0xae, 0xd5, 0x53, 0xf6, 0xfd, 0x55, 0x01, 0xfa, 0x35, 0x40,
	0xec, 0xcd, 0xdd, 0x91, 0x08, 0x00, 0xe9, 0xa4, 0xfb, 0xdb, 0x90, 0xb0, 0x37, 0x97, 0x51, 0xb2,
	0x17, 0xa7, 0x4b, 0x1e, 0x84, 0xfc, 0x74, 0x10, 0x0e, 0x02, 0x5f, 0x7c, 0xac, 0x5c, 0x1a, 0x84,
	0xd8, 0x9b, 0x5b, 0x89, 0x16, 0xe6, 0x17, 0xca, 0x35, 0x6a, 0x41, 0x29, 0x0e, 0x7c, 0x92, 0xde,
	0x9f, 0xcc, 0x08, 0x1f, 0x6f, 0x05, 0x08, 0x7c, 0xd2, 0x24, 0x63, 0xe6, 0x99, 0xa1, 0x1f, 0x0d,
	0x82, 0xf0, 0x1c, 0x03, 0x3f, 0x29, 0x0d, 0x39, 0x81, 0xb2, 0xc0, 0x49, 0x2d, 0xc9, 0xbf, 0x0b,
	0x90, 0x30, 0x41, 0x5a, 0xa4, 0xff, 0x14, 0x60, 0x69, 0x2b, 0xaa, 0x41, 0x31, 0x85, 0xe4, 0x99,
	0x90, 0xc7, 0xe9, 0x56, 0xff, 0x0a, 0xf6, 0x16, 0x94, 0xa0, 0x87, 0x50, 0x9a, 0xc6, 0x64, 0x18,
	0xbc, 0x71, 0x69, 0xf0, 0x3d, 0x91, 0x03, 0x06, 0x24, 0xa2, 0x6e, 0xf0, 0x3d, 0x2f, 0x34, 0xeb,
	0x34, 0x97, 0x33, 0x3c, 0xea, 0x7f, 0x51, 0xe0, 0x60, 0xc3, 0x2e, 0x8e, 0x3a, 0x0c, 0x62, 0xca,
	0xdc, 0xd7, 0xde, 0x78, 0x96, 0xa0, 0xaa, 0x18, 0x84, 0xe8, 0x39, 0x97, 0xf0, 0xf8, 0x14, 0x6f,
	0x3d, 0xf5, 0x62, 0x6f, 0x42, 0x18, 0x89, 0xe5, 0xa8, 0x52, 0xe1, 0xd2, 0x4e, 0x2a, 0xe4, 0x38,
	0xe1, 0x6c, 0x92, 0xf9, 0xa4, 0x14, 0xd6, 0x85, 0xb3, 0x49, 0x3a, 0xc6, 0x7c, 0x04, 0x65, 0xc2,
	0x2f, 0x25, 0x03, 0x77, 0xf1, 0x0d, 0x54, 0xc6, 0x25, 0x29, 0xe3, 0x43, 0x0f, 0xb7, 0xf0, 0xd6,
	0x96, 0xef, 0x1f, 0xe4, 0x40, 0x71, 0x39, 0x56, 0xf1, 0x52, 0xf1, 0xd9, 0x35, 0xbf, 0x9c, 0xea,
	0xe9, 0x22, 0xf9, 0x2e, 0x4b, 0x51, 0x8e, 0xbe, 0x80, 0xca, 0xca, 0x13, 0x9e, 0x16, 0xdf, 0x92,
	0x0b, 0x99, 0x29, 0x7c, 0xc9, 0xc7, 0xcd, 0x84, 0x91, 0x84, 0xc7, 0x64, 0xa3, 0xff, 0x47, 0x81,
	0xdb, 0xcb, 0x6e, 0xd8, 0x24, 0xd4, 0x8f, 0x83, 0x29, 0x8b, 0xe2, 0x1f, 0xf7, 0xd8, 0xad, 0xfe,
	0xf0, 0x99, 0x43, 0x1f, 0xc2, 0x5d, 0xfe, 0xaa, 0xcb, 0x97, 0x5e, 0x56, 0xf0, 0xaf, 0x16, 0xff,
	0x49, 0xf8, 0x90, 0x96, 0xfa, 0xe6, 0xf1, 0xe5, 0xd7, 0xac, 0x72, 0x96, 0xfe, 0x31, 0x11, 0xa0,
	0x4f, 0xfe, 0xa0, 0xa4, 0xff, 0x60, 0xc4, 0x7b, 0x7c, 0x00, 0x77, 0x7b, 0x27, 0xd8, 0x34, 0x7a,
	0xdb, 0x5a, 0x66, 0x09, 0x8a, 0xa7, 0x46, 0xfb, 0xcc, 0xc0, 0xbc, 0x57, 0xde, 0x01, 0xd4, 0x75,
	0x1a, 0xbc, 0x7d, 0x9a, 0xf6, 0x33, 0xcb, 0x36, 0x4d, 0x6c, 0xd9, 0xcf, 0xb4, 0x1c, 0x3a, 0x84,
	0x83, 0xbe, 0x7d, 0x66, 0xd8, 0x3d, 0xb3, 0xe9, 0x76, 0x9d, 0x56, 0x4f, 0xa8, 0xab, 0xe8, 0x11,
	0x3c, 0xec, 0x38, 0x3d, 0xd3, 0xe6, 0x0d, 0xb7, 0xfd, 0xd2, 0x3d, 0x31, 0xf0, 0x69, 0xab, 0xdf,
	0x76, 0x8d, 0x4e, 0xa7, 0x6d, 0x35, 0x8c, 0x9e, 0xe5, 0xd8, 0xda, 0xce, 0x93, 0x3f, 0x2b, 0x50,
	0xce, 0x92, 0xcc, 0x7b, 0x78, 0xa7, 0x6d, 0xf4, 0x5a, 0x0e, 0x3e, 0xbd, 0xc4, 0xa0, 0x33, 0xcb,
	0x6e, 0x3a, 0x67, 0x5d, 0x4d, 0x41, 0x7b, 0x90, 0x6f, 0x5b, 0x76, 0xff, 0x85, 0x96, 0xe3, 0x72,
	0xc3, 0x6e, 0x62, 0xc7, 0x6a, 0x6a, 0x2a, 0x2a, 0x82, 0xea, 0x74, 0x5f, 0x68, 0x3b, 0x7c, 0x61,
	0x39, 0x5d, 0x2d, 0x8f, 0x34, 0x28, 0x1b, 0xf6, 0x4b, 0x37, 0x45, 0xd6, 0x0a, 0xe8, 0x00, 0x2a,
	0x46, 0xbb, 0xbd, 0x90, 0x74, 0xb5, 0x22, 0x02, 0x28, 0x34, 0x4e, 0xb0, 0x73, 0x6a, 0x6a, 0xbb,
	0x4f, 0x5a, 0xb0, 0xbf, 0x56, 0x6c, 0xd1, 0x31, 0x7c, 0xd8, 0x70, 0x4e, 0x3b, 0xd8, 0xec, 0x76,
	0x2d, 0xc7, 0xde, 0x66, 0x5c, 0x11, 0x54, 0x6c, 0x9c, 0x69, 0x0a, 0xda, 0x85, 0x1d, 0x6c, 0x35,
	0x4c, 0x2d, 0xf7, 0xe4, 0x1b, 0xd8, 0x5f, 0x73, 0x3c, 0xfa, 0x08, 0xee, 0x4b, 0xc2, 0x4d, 0xbb,
	0x87, 0x5f, 0x5e, 0x02, 0xd4, 0xc7, 0x6d, 0x4d, 0x41, 0x55, 0x00, 0xf3, 0x85, 0xd9, 0xe8, 0xf7,
	0x8c, 0xdf, 0xb5, 0x4d, 0x2d, 0x87, 0xca, 0xb0, 0x6b, 0x75, 0x5c, 0x6c, 0xd8, 0xcf, 0x4c, 0x4d,
	0x7d, 0x55, 0x10, 0x2e, 0xff, 0xf4, 0xff, 0x01, 0x00, 0x00, 0xff, 0xff, 0x68, 0x6b, 0x92, 0x80,
	0x9e, 0x13, 0x00, 0x00,
}
