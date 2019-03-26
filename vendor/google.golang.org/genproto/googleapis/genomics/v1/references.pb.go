// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/genomics/v1/references.proto

package genomics

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// A reference is a canonical assembled DNA sequence, intended to act as a
// reference coordinate space for other genomic annotations. A single reference
// might represent the human chromosome 1 or mitochandrial DNA, for instance. A
// reference belongs to one or more reference sets.
//
// For more genomics resource definitions, see [Fundamentals of Google
// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
type Reference struct {
	// The server-generated reference ID, unique across all references.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// The length of this reference's sequence.
	Length int64 `protobuf:"varint,2,opt,name=length" json:"length,omitempty"`
	// MD5 of the upper-case sequence excluding all whitespace characters (this
	// is equivalent to SQ:M5 in SAM). This value is represented in lower case
	// hexadecimal format.
	Md5Checksum string `protobuf:"bytes,3,opt,name=md5checksum" json:"md5checksum,omitempty"`
	// The name of this reference, for example `22`.
	Name string `protobuf:"bytes,4,opt,name=name" json:"name,omitempty"`
	// The URI from which the sequence was obtained. Typically specifies a FASTA
	// format file.
	SourceUri string `protobuf:"bytes,5,opt,name=source_uri,json=sourceUri" json:"source_uri,omitempty"`
	// All known corresponding accession IDs in INSDC (GenBank/ENA/DDBJ) ideally
	// with a version number, for example `GCF_000001405.26`.
	SourceAccessions []string `protobuf:"bytes,6,rep,name=source_accessions,json=sourceAccessions" json:"source_accessions,omitempty"`
	// ID from https://www.ncbi.nlm.nih.gov/taxonomy. For example, 9606 for human.
	NcbiTaxonId int32 `protobuf:"varint,7,opt,name=ncbi_taxon_id,json=ncbiTaxonId" json:"ncbi_taxon_id,omitempty"`
}

func (m *Reference) Reset()                    { *m = Reference{} }
func (m *Reference) String() string            { return proto.CompactTextString(m) }
func (*Reference) ProtoMessage()               {}
func (*Reference) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{0} }

func (m *Reference) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Reference) GetLength() int64 {
	if m != nil {
		return m.Length
	}
	return 0
}

func (m *Reference) GetMd5Checksum() string {
	if m != nil {
		return m.Md5Checksum
	}
	return ""
}

func (m *Reference) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Reference) GetSourceUri() string {
	if m != nil {
		return m.SourceUri
	}
	return ""
}

func (m *Reference) GetSourceAccessions() []string {
	if m != nil {
		return m.SourceAccessions
	}
	return nil
}

func (m *Reference) GetNcbiTaxonId() int32 {
	if m != nil {
		return m.NcbiTaxonId
	}
	return 0
}

// A reference set is a set of references which typically comprise a reference
// assembly for a species, such as `GRCh38` which is representative
// of the human genome. A reference set defines a common coordinate space for
// comparing reference-aligned experimental data. A reference set contains 1 or
// more references.
//
// For more genomics resource definitions, see [Fundamentals of Google
// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
type ReferenceSet struct {
	// The server-generated reference set ID, unique across all reference sets.
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	// The IDs of the reference objects that are part of this set.
	// `Reference.md5checksum` must be unique within this set.
	ReferenceIds []string `protobuf:"bytes,2,rep,name=reference_ids,json=referenceIds" json:"reference_ids,omitempty"`
	// Order-independent MD5 checksum which identifies this reference set. The
	// checksum is computed by sorting all lower case hexidecimal string
	// `reference.md5checksum` (for all reference in this set) in
	// ascending lexicographic order, concatenating, and taking the MD5 of that
	// value. The resulting value is represented in lower case hexadecimal format.
	Md5Checksum string `protobuf:"bytes,3,opt,name=md5checksum" json:"md5checksum,omitempty"`
	// ID from https://www.ncbi.nlm.nih.gov/taxonomy (for example, 9606 for human)
	// indicating the species which this reference set is intended to model. Note
	// that contained references may specify a different `ncbiTaxonId`, as
	// assemblies may contain reference sequences which do not belong to the
	// modeled species, for example EBV in a human reference genome.
	NcbiTaxonId int32 `protobuf:"varint,4,opt,name=ncbi_taxon_id,json=ncbiTaxonId" json:"ncbi_taxon_id,omitempty"`
	// Free text description of this reference set.
	Description string `protobuf:"bytes,5,opt,name=description" json:"description,omitempty"`
	// Public id of this reference set, such as `GRCh37`.
	AssemblyId string `protobuf:"bytes,6,opt,name=assembly_id,json=assemblyId" json:"assembly_id,omitempty"`
	// The URI from which the references were obtained.
	SourceUri string `protobuf:"bytes,7,opt,name=source_uri,json=sourceUri" json:"source_uri,omitempty"`
	// All known corresponding accession IDs in INSDC (GenBank/ENA/DDBJ) ideally
	// with a version number, for example `NC_000001.11`.
	SourceAccessions []string `protobuf:"bytes,8,rep,name=source_accessions,json=sourceAccessions" json:"source_accessions,omitempty"`
}

func (m *ReferenceSet) Reset()                    { *m = ReferenceSet{} }
func (m *ReferenceSet) String() string            { return proto.CompactTextString(m) }
func (*ReferenceSet) ProtoMessage()               {}
func (*ReferenceSet) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{1} }

func (m *ReferenceSet) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *ReferenceSet) GetReferenceIds() []string {
	if m != nil {
		return m.ReferenceIds
	}
	return nil
}

func (m *ReferenceSet) GetMd5Checksum() string {
	if m != nil {
		return m.Md5Checksum
	}
	return ""
}

func (m *ReferenceSet) GetNcbiTaxonId() int32 {
	if m != nil {
		return m.NcbiTaxonId
	}
	return 0
}

func (m *ReferenceSet) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func (m *ReferenceSet) GetAssemblyId() string {
	if m != nil {
		return m.AssemblyId
	}
	return ""
}

func (m *ReferenceSet) GetSourceUri() string {
	if m != nil {
		return m.SourceUri
	}
	return ""
}

func (m *ReferenceSet) GetSourceAccessions() []string {
	if m != nil {
		return m.SourceAccessions
	}
	return nil
}

type SearchReferenceSetsRequest struct {
	// If present, return reference sets for which the
	// [md5checksum][google.genomics.v1.ReferenceSet.md5checksum] matches exactly.
	Md5Checksums []string `protobuf:"bytes,1,rep,name=md5checksums" json:"md5checksums,omitempty"`
	// If present, return reference sets for which a prefix of any of
	// [sourceAccessions][google.genomics.v1.ReferenceSet.source_accessions]
	// match any of these strings. Accession numbers typically have a main number
	// and a version, for example `NC_000001.11`.
	Accessions []string `protobuf:"bytes,2,rep,name=accessions" json:"accessions,omitempty"`
	// If present, return reference sets for which a substring of their
	// `assemblyId` matches this string (case insensitive).
	AssemblyId string `protobuf:"bytes,3,opt,name=assembly_id,json=assemblyId" json:"assembly_id,omitempty"`
	// The continuation token, which is used to page through large result sets.
	// To get the next page of results, set this parameter to the value of
	// `nextPageToken` from the previous response.
	PageToken string `protobuf:"bytes,4,opt,name=page_token,json=pageToken" json:"page_token,omitempty"`
	// The maximum number of results to return in a single page. If unspecified,
	// defaults to 1024. The maximum value is 4096.
	PageSize int32 `protobuf:"varint,5,opt,name=page_size,json=pageSize" json:"page_size,omitempty"`
}

func (m *SearchReferenceSetsRequest) Reset()                    { *m = SearchReferenceSetsRequest{} }
func (m *SearchReferenceSetsRequest) String() string            { return proto.CompactTextString(m) }
func (*SearchReferenceSetsRequest) ProtoMessage()               {}
func (*SearchReferenceSetsRequest) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{2} }

func (m *SearchReferenceSetsRequest) GetMd5Checksums() []string {
	if m != nil {
		return m.Md5Checksums
	}
	return nil
}

func (m *SearchReferenceSetsRequest) GetAccessions() []string {
	if m != nil {
		return m.Accessions
	}
	return nil
}

func (m *SearchReferenceSetsRequest) GetAssemblyId() string {
	if m != nil {
		return m.AssemblyId
	}
	return ""
}

func (m *SearchReferenceSetsRequest) GetPageToken() string {
	if m != nil {
		return m.PageToken
	}
	return ""
}

func (m *SearchReferenceSetsRequest) GetPageSize() int32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

type SearchReferenceSetsResponse struct {
	// The matching references sets.
	ReferenceSets []*ReferenceSet `protobuf:"bytes,1,rep,name=reference_sets,json=referenceSets" json:"reference_sets,omitempty"`
	// The continuation token, which is used to page through large result sets.
	// Provide this value in a subsequent request to return the next page of
	// results. This field will be empty if there aren't any additional results.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken" json:"next_page_token,omitempty"`
}

func (m *SearchReferenceSetsResponse) Reset()                    { *m = SearchReferenceSetsResponse{} }
func (m *SearchReferenceSetsResponse) String() string            { return proto.CompactTextString(m) }
func (*SearchReferenceSetsResponse) ProtoMessage()               {}
func (*SearchReferenceSetsResponse) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{3} }

func (m *SearchReferenceSetsResponse) GetReferenceSets() []*ReferenceSet {
	if m != nil {
		return m.ReferenceSets
	}
	return nil
}

func (m *SearchReferenceSetsResponse) GetNextPageToken() string {
	if m != nil {
		return m.NextPageToken
	}
	return ""
}

type GetReferenceSetRequest struct {
	// The ID of the reference set.
	ReferenceSetId string `protobuf:"bytes,1,opt,name=reference_set_id,json=referenceSetId" json:"reference_set_id,omitempty"`
}

func (m *GetReferenceSetRequest) Reset()                    { *m = GetReferenceSetRequest{} }
func (m *GetReferenceSetRequest) String() string            { return proto.CompactTextString(m) }
func (*GetReferenceSetRequest) ProtoMessage()               {}
func (*GetReferenceSetRequest) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{4} }

func (m *GetReferenceSetRequest) GetReferenceSetId() string {
	if m != nil {
		return m.ReferenceSetId
	}
	return ""
}

type SearchReferencesRequest struct {
	// If present, return references for which the
	// [md5checksum][google.genomics.v1.Reference.md5checksum] matches exactly.
	Md5Checksums []string `protobuf:"bytes,1,rep,name=md5checksums" json:"md5checksums,omitempty"`
	// If present, return references for which a prefix of any of
	// [sourceAccessions][google.genomics.v1.Reference.source_accessions] match
	// any of these strings. Accession numbers typically have a main number and a
	// version, for example `GCF_000001405.26`.
	Accessions []string `protobuf:"bytes,2,rep,name=accessions" json:"accessions,omitempty"`
	// If present, return only references which belong to this reference set.
	ReferenceSetId string `protobuf:"bytes,3,opt,name=reference_set_id,json=referenceSetId" json:"reference_set_id,omitempty"`
	// The continuation token, which is used to page through large result sets.
	// To get the next page of results, set this parameter to the value of
	// `nextPageToken` from the previous response.
	PageToken string `protobuf:"bytes,4,opt,name=page_token,json=pageToken" json:"page_token,omitempty"`
	// The maximum number of results to return in a single page. If unspecified,
	// defaults to 1024. The maximum value is 4096.
	PageSize int32 `protobuf:"varint,5,opt,name=page_size,json=pageSize" json:"page_size,omitempty"`
}

func (m *SearchReferencesRequest) Reset()                    { *m = SearchReferencesRequest{} }
func (m *SearchReferencesRequest) String() string            { return proto.CompactTextString(m) }
func (*SearchReferencesRequest) ProtoMessage()               {}
func (*SearchReferencesRequest) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{5} }

func (m *SearchReferencesRequest) GetMd5Checksums() []string {
	if m != nil {
		return m.Md5Checksums
	}
	return nil
}

func (m *SearchReferencesRequest) GetAccessions() []string {
	if m != nil {
		return m.Accessions
	}
	return nil
}

func (m *SearchReferencesRequest) GetReferenceSetId() string {
	if m != nil {
		return m.ReferenceSetId
	}
	return ""
}

func (m *SearchReferencesRequest) GetPageToken() string {
	if m != nil {
		return m.PageToken
	}
	return ""
}

func (m *SearchReferencesRequest) GetPageSize() int32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

type SearchReferencesResponse struct {
	// The matching references.
	References []*Reference `protobuf:"bytes,1,rep,name=references" json:"references,omitempty"`
	// The continuation token, which is used to page through large result sets.
	// Provide this value in a subsequent request to return the next page of
	// results. This field will be empty if there aren't any additional results.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken" json:"next_page_token,omitempty"`
}

func (m *SearchReferencesResponse) Reset()                    { *m = SearchReferencesResponse{} }
func (m *SearchReferencesResponse) String() string            { return proto.CompactTextString(m) }
func (*SearchReferencesResponse) ProtoMessage()               {}
func (*SearchReferencesResponse) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{6} }

func (m *SearchReferencesResponse) GetReferences() []*Reference {
	if m != nil {
		return m.References
	}
	return nil
}

func (m *SearchReferencesResponse) GetNextPageToken() string {
	if m != nil {
		return m.NextPageToken
	}
	return ""
}

type GetReferenceRequest struct {
	// The ID of the reference.
	ReferenceId string `protobuf:"bytes,1,opt,name=reference_id,json=referenceId" json:"reference_id,omitempty"`
}

func (m *GetReferenceRequest) Reset()                    { *m = GetReferenceRequest{} }
func (m *GetReferenceRequest) String() string            { return proto.CompactTextString(m) }
func (*GetReferenceRequest) ProtoMessage()               {}
func (*GetReferenceRequest) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{7} }

func (m *GetReferenceRequest) GetReferenceId() string {
	if m != nil {
		return m.ReferenceId
	}
	return ""
}

type ListBasesRequest struct {
	// The ID of the reference.
	ReferenceId string `protobuf:"bytes,1,opt,name=reference_id,json=referenceId" json:"reference_id,omitempty"`
	// The start position (0-based) of this query. Defaults to 0.
	Start int64 `protobuf:"varint,2,opt,name=start" json:"start,omitempty"`
	// The end position (0-based, exclusive) of this query. Defaults to the length
	// of this reference.
	End int64 `protobuf:"varint,3,opt,name=end" json:"end,omitempty"`
	// The continuation token, which is used to page through large result sets.
	// To get the next page of results, set this parameter to the value of
	// `nextPageToken` from the previous response.
	PageToken string `protobuf:"bytes,4,opt,name=page_token,json=pageToken" json:"page_token,omitempty"`
	// The maximum number of bases to return in a single page. If unspecified,
	// defaults to 200Kbp (kilo base pairs). The maximum value is 10Mbp (mega base
	// pairs).
	PageSize int32 `protobuf:"varint,5,opt,name=page_size,json=pageSize" json:"page_size,omitempty"`
}

func (m *ListBasesRequest) Reset()                    { *m = ListBasesRequest{} }
func (m *ListBasesRequest) String() string            { return proto.CompactTextString(m) }
func (*ListBasesRequest) ProtoMessage()               {}
func (*ListBasesRequest) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{8} }

func (m *ListBasesRequest) GetReferenceId() string {
	if m != nil {
		return m.ReferenceId
	}
	return ""
}

func (m *ListBasesRequest) GetStart() int64 {
	if m != nil {
		return m.Start
	}
	return 0
}

func (m *ListBasesRequest) GetEnd() int64 {
	if m != nil {
		return m.End
	}
	return 0
}

func (m *ListBasesRequest) GetPageToken() string {
	if m != nil {
		return m.PageToken
	}
	return ""
}

func (m *ListBasesRequest) GetPageSize() int32 {
	if m != nil {
		return m.PageSize
	}
	return 0
}

type ListBasesResponse struct {
	// The offset position (0-based) of the given `sequence` from the
	// start of this `Reference`. This value will differ for each page
	// in a paginated request.
	Offset int64 `protobuf:"varint,1,opt,name=offset" json:"offset,omitempty"`
	// A substring of the bases that make up this reference.
	Sequence string `protobuf:"bytes,2,opt,name=sequence" json:"sequence,omitempty"`
	// The continuation token, which is used to page through large result sets.
	// Provide this value in a subsequent request to return the next page of
	// results. This field will be empty if there aren't any additional results.
	NextPageToken string `protobuf:"bytes,3,opt,name=next_page_token,json=nextPageToken" json:"next_page_token,omitempty"`
}

func (m *ListBasesResponse) Reset()                    { *m = ListBasesResponse{} }
func (m *ListBasesResponse) String() string            { return proto.CompactTextString(m) }
func (*ListBasesResponse) ProtoMessage()               {}
func (*ListBasesResponse) Descriptor() ([]byte, []int) { return fileDescriptor10, []int{9} }

func (m *ListBasesResponse) GetOffset() int64 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *ListBasesResponse) GetSequence() string {
	if m != nil {
		return m.Sequence
	}
	return ""
}

func (m *ListBasesResponse) GetNextPageToken() string {
	if m != nil {
		return m.NextPageToken
	}
	return ""
}

func init() {
	proto.RegisterType((*Reference)(nil), "google.genomics.v1.Reference")
	proto.RegisterType((*ReferenceSet)(nil), "google.genomics.v1.ReferenceSet")
	proto.RegisterType((*SearchReferenceSetsRequest)(nil), "google.genomics.v1.SearchReferenceSetsRequest")
	proto.RegisterType((*SearchReferenceSetsResponse)(nil), "google.genomics.v1.SearchReferenceSetsResponse")
	proto.RegisterType((*GetReferenceSetRequest)(nil), "google.genomics.v1.GetReferenceSetRequest")
	proto.RegisterType((*SearchReferencesRequest)(nil), "google.genomics.v1.SearchReferencesRequest")
	proto.RegisterType((*SearchReferencesResponse)(nil), "google.genomics.v1.SearchReferencesResponse")
	proto.RegisterType((*GetReferenceRequest)(nil), "google.genomics.v1.GetReferenceRequest")
	proto.RegisterType((*ListBasesRequest)(nil), "google.genomics.v1.ListBasesRequest")
	proto.RegisterType((*ListBasesResponse)(nil), "google.genomics.v1.ListBasesResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ReferenceServiceV1 service

type ReferenceServiceV1Client interface {
	// Searches for reference sets which match the given criteria.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.searchReferenceSets](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L71)
	SearchReferenceSets(ctx context.Context, in *SearchReferenceSetsRequest, opts ...grpc.CallOption) (*SearchReferenceSetsResponse, error)
	// Gets a reference set.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.getReferenceSet](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L83).
	GetReferenceSet(ctx context.Context, in *GetReferenceSetRequest, opts ...grpc.CallOption) (*ReferenceSet, error)
	// Searches for references which match the given criteria.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.searchReferences](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L146).
	SearchReferences(ctx context.Context, in *SearchReferencesRequest, opts ...grpc.CallOption) (*SearchReferencesResponse, error)
	// Gets a reference.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.getReference](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L158).
	GetReference(ctx context.Context, in *GetReferenceRequest, opts ...grpc.CallOption) (*Reference, error)
	// Lists the bases in a reference, optionally restricted to a range.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.getReferenceBases](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L221).
	ListBases(ctx context.Context, in *ListBasesRequest, opts ...grpc.CallOption) (*ListBasesResponse, error)
}

type referenceServiceV1Client struct {
	cc *grpc.ClientConn
}

func NewReferenceServiceV1Client(cc *grpc.ClientConn) ReferenceServiceV1Client {
	return &referenceServiceV1Client{cc}
}

func (c *referenceServiceV1Client) SearchReferenceSets(ctx context.Context, in *SearchReferenceSetsRequest, opts ...grpc.CallOption) (*SearchReferenceSetsResponse, error) {
	out := new(SearchReferenceSetsResponse)
	err := grpc.Invoke(ctx, "/google.genomics.v1.ReferenceServiceV1/SearchReferenceSets", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *referenceServiceV1Client) GetReferenceSet(ctx context.Context, in *GetReferenceSetRequest, opts ...grpc.CallOption) (*ReferenceSet, error) {
	out := new(ReferenceSet)
	err := grpc.Invoke(ctx, "/google.genomics.v1.ReferenceServiceV1/GetReferenceSet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *referenceServiceV1Client) SearchReferences(ctx context.Context, in *SearchReferencesRequest, opts ...grpc.CallOption) (*SearchReferencesResponse, error) {
	out := new(SearchReferencesResponse)
	err := grpc.Invoke(ctx, "/google.genomics.v1.ReferenceServiceV1/SearchReferences", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *referenceServiceV1Client) GetReference(ctx context.Context, in *GetReferenceRequest, opts ...grpc.CallOption) (*Reference, error) {
	out := new(Reference)
	err := grpc.Invoke(ctx, "/google.genomics.v1.ReferenceServiceV1/GetReference", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *referenceServiceV1Client) ListBases(ctx context.Context, in *ListBasesRequest, opts ...grpc.CallOption) (*ListBasesResponse, error) {
	out := new(ListBasesResponse)
	err := grpc.Invoke(ctx, "/google.genomics.v1.ReferenceServiceV1/ListBases", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ReferenceServiceV1 service

type ReferenceServiceV1Server interface {
	// Searches for reference sets which match the given criteria.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.searchReferenceSets](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L71)
	SearchReferenceSets(context.Context, *SearchReferenceSetsRequest) (*SearchReferenceSetsResponse, error)
	// Gets a reference set.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.getReferenceSet](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L83).
	GetReferenceSet(context.Context, *GetReferenceSetRequest) (*ReferenceSet, error)
	// Searches for references which match the given criteria.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.searchReferences](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L146).
	SearchReferences(context.Context, *SearchReferencesRequest) (*SearchReferencesResponse, error)
	// Gets a reference.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.getReference](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L158).
	GetReference(context.Context, *GetReferenceRequest) (*Reference, error)
	// Lists the bases in a reference, optionally restricted to a range.
	//
	// For the definitions of references and other genomics resources, see
	// [Fundamentals of Google
	// Genomics](https://cloud.google.com/genomics/fundamentals-of-google-genomics)
	//
	// Implements
	// [GlobalAllianceApi.getReferenceBases](https://github.com/ga4gh/schemas/blob/v0.5.1/src/main/resources/avro/referencemethods.avdl#L221).
	ListBases(context.Context, *ListBasesRequest) (*ListBasesResponse, error)
}

func RegisterReferenceServiceV1Server(s *grpc.Server, srv ReferenceServiceV1Server) {
	s.RegisterService(&_ReferenceServiceV1_serviceDesc, srv)
}

func _ReferenceServiceV1_SearchReferenceSets_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchReferenceSetsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReferenceServiceV1Server).SearchReferenceSets(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.genomics.v1.ReferenceServiceV1/SearchReferenceSets",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReferenceServiceV1Server).SearchReferenceSets(ctx, req.(*SearchReferenceSetsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReferenceServiceV1_GetReferenceSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReferenceSetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReferenceServiceV1Server).GetReferenceSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.genomics.v1.ReferenceServiceV1/GetReferenceSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReferenceServiceV1Server).GetReferenceSet(ctx, req.(*GetReferenceSetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReferenceServiceV1_SearchReferences_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SearchReferencesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReferenceServiceV1Server).SearchReferences(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.genomics.v1.ReferenceServiceV1/SearchReferences",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReferenceServiceV1Server).SearchReferences(ctx, req.(*SearchReferencesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReferenceServiceV1_GetReference_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetReferenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReferenceServiceV1Server).GetReference(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.genomics.v1.ReferenceServiceV1/GetReference",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReferenceServiceV1Server).GetReference(ctx, req.(*GetReferenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ReferenceServiceV1_ListBases_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListBasesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ReferenceServiceV1Server).ListBases(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/google.genomics.v1.ReferenceServiceV1/ListBases",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ReferenceServiceV1Server).ListBases(ctx, req.(*ListBasesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _ReferenceServiceV1_serviceDesc = grpc.ServiceDesc{
	ServiceName: "google.genomics.v1.ReferenceServiceV1",
	HandlerType: (*ReferenceServiceV1Server)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SearchReferenceSets",
			Handler:    _ReferenceServiceV1_SearchReferenceSets_Handler,
		},
		{
			MethodName: "GetReferenceSet",
			Handler:    _ReferenceServiceV1_GetReferenceSet_Handler,
		},
		{
			MethodName: "SearchReferences",
			Handler:    _ReferenceServiceV1_SearchReferences_Handler,
		},
		{
			MethodName: "GetReference",
			Handler:    _ReferenceServiceV1_GetReference_Handler,
		},
		{
			MethodName: "ListBases",
			Handler:    _ReferenceServiceV1_ListBases_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "google/genomics/v1/references.proto",
}

func init() { proto.RegisterFile("google/genomics/v1/references.proto", fileDescriptor10) }

var fileDescriptor10 = []byte{
	// 851 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x56, 0x41, 0x6f, 0x1b, 0x45,
	0x14, 0xd6, 0x78, 0x63, 0x37, 0x7e, 0x76, 0x12, 0xf7, 0x15, 0xc2, 0xca, 0x25, 0xd4, 0x6c, 0x9a,
	0x62, 0x35, 0x95, 0x57, 0x29, 0x42, 0x42, 0x45, 0x1c, 0xc8, 0xa5, 0x8a, 0xc4, 0x21, 0xda, 0x14,
	0x0e, 0x5c, 0x56, 0x9b, 0xdd, 0x89, 0x33, 0x34, 0xde, 0x31, 0x3b, 0x93, 0xa8, 0xb4, 0xca, 0x01,
	0x24, 0x8e, 0xc0, 0x81, 0x0b, 0x88, 0xdf, 0xc2, 0x89, 0x9f, 0xc0, 0x09, 0x71, 0xe5, 0x47, 0x70,
	0x44, 0x33, 0x3b, 0xbb, 0x1e, 0xaf, 0x97, 0xd8, 0x52, 0xb9, 0xed, 0x7c, 0xf3, 0xe6, 0xcd, 0xf7,
	0x7d, 0x6f, 0xde, 0xec, 0xc0, 0xee, 0x98, 0xf3, 0xf1, 0x05, 0xf5, 0xc7, 0x34, 0xe5, 0x13, 0x16,
	0x0b, 0xff, 0xea, 0xc0, 0xcf, 0xe8, 0x19, 0xcd, 0x68, 0x1a, 0x53, 0x31, 0x9a, 0x66, 0x5c, 0x72,
	0xc4, 0x3c, 0x68, 0x54, 0x04, 0x8d, 0xae, 0x0e, 0xfa, 0x6f, 0x9b, 0x85, 0xd1, 0x94, 0xf9, 0x51,
	0x9a, 0x72, 0x19, 0x49, 0xc6, 0x53, 0xb3, 0xc2, 0xfb, 0x93, 0x40, 0x3b, 0x28, 0xd2, 0xe0, 0x26,
	0x34, 0x58, 0xe2, 0x92, 0x01, 0x19, 0xb6, 0x83, 0x06, 0x4b, 0x70, 0x1b, 0x5a, 0x17, 0x34, 0x1d,
	0xcb, 0x73, 0xb7, 0x31, 0x20, 0x43, 0x27, 0x30, 0x23, 0x1c, 0x40, 0x67, 0x92, 0x7c, 0x10, 0x9f,
	0xd3, 0xf8, 0xb9, 0xb8, 0x9c, 0xb8, 0x8e, 0x5e, 0x60, 0x43, 0x88, 0xb0, 0x96, 0x46, 0x13, 0xea,
	0xae, 0xe9, 0x29, 0xfd, 0x8d, 0x3b, 0x00, 0x82, 0x5f, 0x66, 0x31, 0x0d, 0x2f, 0x33, 0xe6, 0x36,
	0xf5, 0x4c, 0x3b, 0x47, 0x3e, 0xcb, 0x18, 0xee, 0xc3, 0x6d, 0x33, 0x1d, 0xc5, 0x31, 0x15, 0x42,
	0xb1, 0x74, 0x5b, 0x03, 0x67, 0xd8, 0x0e, 0x7a, 0xf9, 0xc4, 0x27, 0x25, 0x8e, 0x1e, 0x6c, 0xa4,
	0xf1, 0x29, 0x0b, 0x65, 0xf4, 0x82, 0xa7, 0x21, 0x4b, 0xdc, 0x5b, 0x03, 0x32, 0x6c, 0x06, 0x1d,
	0x05, 0x3e, 0x53, 0xd8, 0x51, 0xe2, 0xfd, 0xdc, 0x80, 0x6e, 0xa9, 0xed, 0x84, 0xca, 0x05, 0x79,
	0xbb, 0xb0, 0x51, 0x5a, 0x18, 0xb2, 0x44, 0xb8, 0x0d, 0xbd, 0x5b, 0xb7, 0x04, 0x8f, 0x12, 0xb1,
	0x82, 0xd6, 0x05, 0x2e, 0x6b, 0x0b, 0x5c, 0x54, 0x96, 0x84, 0x8a, 0x38, 0x63, 0x53, 0xe5, 0xbe,
	0x11, 0x6f, 0x43, 0x78, 0x0f, 0x3a, 0x91, 0x10, 0x74, 0x72, 0x7a, 0xf1, 0xb5, 0xca, 0xd1, 0xd2,
	0x11, 0x50, 0x40, 0x47, 0x49, 0xc5, 0xbe, 0x5b, 0x2b, 0xd9, 0xb7, 0x5e, 0x6f, 0x9f, 0xf7, 0x1b,
	0x81, 0xfe, 0x09, 0x8d, 0xb2, 0xf8, 0xdc, 0x36, 0x48, 0x04, 0xf4, 0xab, 0x4b, 0x2a, 0x24, 0x7a,
	0xd0, 0xb5, 0x04, 0x0a, 0x97, 0xe4, 0xbe, 0xd8, 0x18, 0xbe, 0x03, 0x60, 0x6d, 0x94, 0x3b, 0x67,
	0x21, 0x55, 0x3d, 0x4e, 0x9d, 0x9e, 0x69, 0x34, 0xa6, 0xa1, 0xe4, 0xcf, 0x69, 0x6a, 0x0e, 0x4a,
	0x5b, 0x21, 0xcf, 0x14, 0x80, 0x77, 0x41, 0x0f, 0x42, 0xc1, 0x5e, 0x52, 0xed, 0x57, 0x33, 0x58,
	0x57, 0xc0, 0x09, 0x7b, 0x49, 0xbd, 0x1f, 0x08, 0xdc, 0xad, 0xe5, 0x2f, 0xa6, 0x3c, 0x15, 0x14,
	0x9f, 0xc2, 0xe6, 0xac, 0xb2, 0x82, 0xca, 0x5c, 0x42, 0xe7, 0xf1, 0x60, 0xb4, 0xd8, 0x21, 0x23,
	0x3b, 0x45, 0x30, 0x3b, 0x11, 0x2a, 0x21, 0x3e, 0x80, 0xad, 0x94, 0xbe, 0x90, 0xa1, 0xc5, 0xb4,
	0xa1, 0x99, 0x6e, 0x28, 0xf8, 0xb8, 0x60, 0xeb, 0x1d, 0xc2, 0xf6, 0x53, 0x2a, 0xe7, 0x32, 0x19,
	0x2f, 0x87, 0xd0, 0x9b, 0xa3, 0x12, 0x96, 0x47, 0x70, 0xd3, 0xde, 0xea, 0x28, 0xf1, 0x7e, 0x27,
	0xf0, 0x56, 0x45, 0xd4, 0xff, 0x5a, 0x91, 0x3a, 0x26, 0x4e, 0x1d, 0x93, 0xd7, 0x2a, 0xcd, 0x37,
	0x04, 0xdc, 0x45, 0x15, 0xa6, 0x2e, 0x1f, 0x03, 0xcc, 0x2e, 0x2d, 0x53, 0x93, 0x9d, 0x1b, 0x6b,
	0x12, 0x58, 0x0b, 0x56, 0xae, 0xc6, 0x87, 0x70, 0xc7, 0xae, 0x46, 0x61, 0xe2, 0xbb, 0xd0, 0xb5,
	0xfb, 0xdd, 0x94, 0xa1, 0x63, 0xb5, 0xbb, 0xf7, 0x0b, 0x81, 0xde, 0xa7, 0x4c, 0xc8, 0xc3, 0x48,
	0xcc, 0xcc, 0x5f, 0xbe, 0x0e, 0xdf, 0x80, 0xa6, 0x90, 0x51, 0x26, 0xcd, 0x45, 0x99, 0x0f, 0xb0,
	0x07, 0x0e, 0x4d, 0x73, 0x93, 0x9d, 0x40, 0x7d, 0xbe, 0x96, 0xb3, 0x1c, 0x6e, 0x5b, 0xd4, 0x8c,
	0xa3, 0xdb, 0xd0, 0xe2, 0x67, 0x67, 0x82, 0x4a, 0xcd, 0xca, 0x09, 0xcc, 0x08, 0xfb, 0xb0, 0x2e,
	0x14, 0xfd, 0x34, 0xa6, 0xc6, 0xa3, 0x72, 0x5c, 0x67, 0xa3, 0x53, 0x63, 0xe3, 0xe3, 0xbf, 0x9a,
	0x80, 0xd6, 0x91, 0xce, 0xae, 0x58, 0x4c, 0x3f, 0x3f, 0xc0, 0x5f, 0x09, 0xdc, 0xa9, 0x69, 0x3e,
	0x1c, 0xd5, 0x15, 0xf2, 0xbf, 0x6f, 0x99, 0xbe, 0xbf, 0x72, 0x7c, 0xae, 0xd5, 0xdb, 0xfd, 0xf6,
	0x8f, 0xbf, 0x7f, 0x6a, 0xec, 0x78, 0xee, 0xfc, 0xcf, 0x8f, 0x4a, 0xe1, 0x0b, 0xbd, 0xec, 0x09,
	0x79, 0x88, 0xdf, 0x13, 0xd8, 0xaa, 0xb4, 0x22, 0x3e, 0xac, 0xdb, 0xa9, 0xbe, 0x5f, 0xfb, 0x4b,
	0xaf, 0x08, 0xef, 0x91, 0xa6, 0xf1, 0x00, 0xef, 0x2f, 0xd2, 0x78, 0x55, 0x6d, 0xb0, 0x6b, 0xfc,
	0x91, 0x40, 0xaf, 0xda, 0x0f, 0xb8, 0xbf, 0x82, 0xf4, 0xd2, 0xa7, 0x47, 0xab, 0x05, 0x1b, 0x93,
	0x06, 0x9a, 0x5d, 0xdf, 0x7b, 0x73, 0x9e, 0x9d, 0xe5, 0xd0, 0x35, 0x74, 0x6d, 0xed, 0xf8, 0xde,
	0x32, 0x77, 0x0a, 0x22, 0x37, 0x77, 0xaa, 0xb7, 0xa7, 0x77, 0xbe, 0x87, 0x3b, 0x95, 0x9d, 0x5f,
	0xd9, 0xcd, 0x73, 0x8d, 0xdf, 0x11, 0x68, 0x97, 0xe7, 0x18, 0xef, 0xd7, 0xe5, 0xac, 0x76, 0x60,
	0x7f, 0x6f, 0x49, 0x94, 0xd1, 0xbe, 0xaf, 0x19, 0xec, 0xe1, 0xee, 0x8d, 0x0c, 0xfc, 0x53, 0xb5,
	0xe8, 0xf0, 0x4b, 0xd8, 0x8e, 0xf9, 0xa4, 0x26, 0xf1, 0xe1, 0xd6, 0xcc, 0xd6, 0x63, 0xf5, 0x4a,
	0x3a, 0x26, 0x5f, 0x3c, 0x29, 0xc2, 0xf8, 0x45, 0x94, 0x8e, 0x47, 0x3c, 0x1b, 0xab, 0x97, 0x98,
	0x7e, 0x43, 0xf9, 0xf9, 0x54, 0x34, 0x65, 0xc2, 0x7e, 0x9d, 0x7d, 0x54, 0x7c, 0xff, 0x43, 0xc8,
	0x69, 0x4b, 0x47, 0xbe, 0xff, 0x6f, 0x00, 0x00, 0x00, 0xff, 0xff, 0xd7, 0x9f, 0xb6, 0x11, 0xc6,
	0x09, 0x00, 0x00,
}
