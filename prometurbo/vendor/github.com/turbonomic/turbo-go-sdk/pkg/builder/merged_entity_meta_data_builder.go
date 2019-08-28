package builder

import (
	"fmt"
	"github.com/turbonomic/turbo-go-sdk/pkg/proto"
)

type ReturnType string

const (
	MergedEntityMetadata_STRING      ReturnType = "String"
	MergedEntityMetadata_LIST_STRING ReturnType = "List"
)

var (
	returnTypeMapping = map[ReturnType]proto.MergedEntityMetadata_ReturnType{
		MergedEntityMetadata_STRING:      proto.MergedEntityMetadata_STRING,
		MergedEntityMetadata_LIST_STRING: proto.MergedEntityMetadata_LIST_STRING,
	}
)

// ============================ MergedEntityMetadata_MatchingMetadata ==============================
// matchingData is structure to hold all the fields of the MatchingData proto message.
// Either a property of an entity or a field or entity OID will be used in the MatchingData message
type matchingData struct {
	propertyName string
	delimiter    string
	fieldName    string
	fieldPaths   []string
	useEntityOid bool
}

// MatchingData field represents the kind of data we will extract for matching the entities.
// It can be a property which is extracted from the entity property map,
// or it can be a field which is named within the entityDTO itself,
// or the entity OID in XL repository.
// In some cases, we encode a List of Strings as a single string.
// In that case, one can specify a delimiter that separates different strings in the value.
// For example, we have a PM_UUID_LIST property where we have a comma separated list of UUIDs in a single string.
//	message MatchingData {
//		oneof matching_data {
//			EntityPropertyName matching_property = 100;
//			EntityField matching_field = 101;
//			EntityOid matching_entity_oid = 102;
//		}
//		optional string delimiter = 200;
//	}
func newMatchingData(matchingData *matchingData) *proto.MergedEntityMetadata_MatchingData {
	// Create MergedEntityMetadata/MatchingMetadata/MatchingData for the internal property
	matchingDataBuilder := &proto.MergedEntityMetadata_MatchingData{}

	propertyName := matchingData.propertyName
	if propertyName != "" {
		entityPropertyNameBuilder := &proto.MergedEntityMetadata_EntityPropertyName{}
		entityPropertyNameBuilder.PropertyName = &propertyName

		matchingDataProperty := &proto.MergedEntityMetadata_MatchingData_MatchingProperty{}
		matchingDataProperty.MatchingProperty = entityPropertyNameBuilder

		matchingDataBuilder.MatchingData = matchingDataProperty
	}

	fieldName := matchingData.fieldName
	if fieldName != "" {
		entityFieldBuilder := &proto.MergedEntityMetadata_EntityField{}
		entityFieldBuilder.FieldName = &fieldName
		entityFieldBuilder.MessagePath = matchingData.fieldPaths

		matchingDataField := &proto.MergedEntityMetadata_MatchingData_MatchingField{}
		matchingDataField.MatchingField = entityFieldBuilder

		matchingDataBuilder.MatchingData = matchingDataField
	}

	if matchingData.useEntityOid {
		oidBuilder := &proto.MergedEntityMetadata_EntityOid{}

		matchingDataOid := &proto.MergedEntityMetadata_MatchingData_MatchingEntityOid{}
		matchingDataOid.MatchingEntityOid = oidBuilder

		matchingDataBuilder.MatchingData = matchingDataOid
	}

	if matchingData.delimiter != "" {
		matchingDataBuilder.Delimiter = &matchingData.delimiter
	}

	return matchingDataBuilder
}

type matchingMetadataBuilder struct {
	internalReturnType   ReturnType
	externalReturnType   ReturnType
	internalMatchingData []*matchingData
	externalMatchingData []*matchingData
}

func newMatchingMetadataBuilder() *matchingMetadataBuilder {
	builder := &matchingMetadataBuilder{
		internalMatchingData: []*matchingData{},
		externalMatchingData: []*matchingData{},
	}
	return builder
}

func (builder *matchingMetadataBuilder) build() (*proto.MergedEntityMetadata_MatchingMetadata, error) {
	// validate internal property return type
	if builder.internalReturnType == "" {
		return nil, fmt.Errorf("internal entity metadata return type not set")
	}
	internalReturnType, exists := returnTypeMapping[builder.internalReturnType]
	if !exists {
		return nil, fmt.Errorf("unknown internal entity metadata return type")
	}
	// validate external property return type
	if builder.externalReturnType == "" {
		return nil, fmt.Errorf("external entity metadata return type not set")
	}
	externalReturnType, exists := returnTypeMapping[builder.externalReturnType]
	if !exists {
		return nil, fmt.Errorf("unknown external entity metadata return type")
	}
	matchingMetadata := &proto.MergedEntityMetadata_MatchingMetadata{
		ReturnType:               &internalReturnType,
		ExternalEntityReturnType: &externalReturnType,
	}
	// create internal property matching data
	for _, internalData := range builder.internalMatchingData {
		matchingMetadata.MatchingData =
			append(matchingMetadata.MatchingData, newMatchingData(internalData))
	}
	// create external property matching data
	for _, externalData := range builder.externalMatchingData {
		matchingMetadata.ExternalEntityMatchingProperty =
			append(matchingMetadata.ExternalEntityMatchingProperty, newMatchingData(externalData))
	}
	return matchingMetadata, nil
}

func (builder *matchingMetadataBuilder) addInternalMatchingData(internal *matchingData) *matchingMetadataBuilder {
	builder.internalMatchingData = append(builder.internalMatchingData, internal)
	return builder
}

func (builder *matchingMetadataBuilder) addExternalMatchingData(external *matchingData) *matchingMetadataBuilder {
	builder.externalMatchingData = append(builder.externalMatchingData, external)
	return builder
}

// ============================= MergedEntityMetadata_EntityPropertyName ===========================
type propertyBuilder struct {
	propertyName string
}

func newPropertyBuilder(propertyName string) *propertyBuilder {
	return &propertyBuilder{
		propertyName: propertyName,
	}
}

func (builder *propertyBuilder) build() *proto.MergedEntityMetadata_EntityPropertyName {
	return &proto.MergedEntityMetadata_EntityPropertyName{
		PropertyName: &builder.propertyName,
	}
}

// =========================== MergedEntityMetadata_EntityField ===================================
type fieldBuilder struct {
	fieldName  string
	fieldPaths []string
}

func newFieldBuilder(fieldName string, fieldPaths []string) *fieldBuilder {
	return &fieldBuilder{
		fieldName:  fieldName,
		fieldPaths: fieldPaths,
	}
}

func (builder *fieldBuilder) build() *proto.MergedEntityMetadata_EntityField {
	return &proto.MergedEntityMetadata_EntityField{
		FieldName:   &builder.fieldName,
		MessagePath: builder.fieldPaths,
	}
}

// ============================== MergedEntityMetadata_CommodityBoughtMetadata =====================
type commodityBoughtMetadataBuilder struct {
	providerType proto.EntityDTO_EntityType
	// nil providerToReplace indicates that there is no need to replace provider
	providerToReplace *proto.EntityDTO_EntityType
	commBoughtList    []proto.CommodityDTO_CommodityType
}

func newCommodityBoughtMetadataBuilder(providerType proto.EntityDTO_EntityType) *commodityBoughtMetadataBuilder {
	return &commodityBoughtMetadataBuilder{
		providerType: providerType,
	}
}

func (builder *commodityBoughtMetadataBuilder) addBought(commType proto.CommodityDTO_CommodityType) *commodityBoughtMetadataBuilder {
	builder.commBoughtList = append(builder.commBoughtList, commType)
	return builder
}

func (builder *commodityBoughtMetadataBuilder) addBoughtList(commType []proto.CommodityDTO_CommodityType) *commodityBoughtMetadataBuilder {
	builder.commBoughtList = append(builder.commBoughtList, commType...)
	return builder
}

func (builder *commodityBoughtMetadataBuilder) build() *proto.MergedEntityMetadata_CommodityBoughtMetadata {
	return &proto.MergedEntityMetadata_CommodityBoughtMetadata{
		CommodityMetadata: builder.commBoughtList,
		ProviderType:      &builder.providerType,
		ReplacesProvider:  builder.providerToReplace,
	}
}

// ============================== MergedEntityMetadata_CommoditySoldMetadata =====================
type commoditySoldMetadataBuilder struct {
	soldMetadata    *proto.MergedEntityMetadata_CommoditySoldMetadata
	commodityType   proto.CommodityDTO_CommodityType
	ignoreIfPresent bool
	fieldBuilders   []*fieldBuilder
}

func newCommoditySoldMetadataBuilder(commType proto.CommodityDTO_CommodityType, ignoreIfPresent bool) *commoditySoldMetadataBuilder {
	return &commoditySoldMetadataBuilder{
		commodityType:   commType,
		ignoreIfPresent: ignoreIfPresent,
	}
}

func (builder *commoditySoldMetadataBuilder) addField(fieldName string, fieldPaths []string) *commoditySoldMetadataBuilder {
	if fieldName == "" {
		return builder
	}
	builder.fieldBuilders = append(builder.fieldBuilders,
		newFieldBuilder(fieldName, fieldPaths))
	return builder
}

func (builder *commoditySoldMetadataBuilder) build() *proto.MergedEntityMetadata_CommoditySoldMetadata {
	soldMetadata := &proto.MergedEntityMetadata_CommoditySoldMetadata{
		CommodityType:   &builder.commodityType,
		IgnoreIfPresent: &builder.ignoreIfPresent,
	}
	for _, fieldBuilder := range builder.fieldBuilders {
		soldMetadata.PatchedFields = append(soldMetadata.PatchedFields, fieldBuilder.build())
	}
	return soldMetadata
}

// MergedEntityMetadataBuilder is used to create an MergedEntityMetadata object.
// MergedEntityMetadata is a message in the TemplateDTO of the supply chain.  It provides data that
// defines the stitching behavior of entities discovered by a probe.  There should be a
// MergedEntityMetadata entry for each entity type in the probe that is reported as origin "proxy".
// The MergedEntityMetadata is created for stitching in XL server. It combines information that was previously
// contained in various places used for stitching in Classic OpsManager server (e.g. external entity link, replacement
// entity metadata, and etc.) and stores it in one place.
type MergedEntityMetadataBuilder struct {
	// MergedEntityMetadata - the main protobuf structure to return
	metadata *proto.MergedEntityMetadata
	// Deprecated: commoditiesSold exits for backward compatibility and should not be used any more
	commoditiesSold []proto.CommodityDTO_CommodityType
	// MergedEntityMetadata consists of MatchingMetadata for proxy (internal) and external entity
	*matchingMetadataBuilder
	keepStandAlone              bool
	propertyBuilders            []*propertyBuilder
	fieldBuilders               []*fieldBuilder
	commBoughtMetadataList      []*commodityBoughtMetadataBuilder // for different providers
	commoditiesSoldMetadataList []*commoditySoldMetadataBuilder
}

// NewMergedEntityMetadataBuilder initializes a MergedEntityMetadataBuilder object
func NewMergedEntityMetadataBuilder() *MergedEntityMetadataBuilder {
	builder := &MergedEntityMetadataBuilder{
		metadata:                &proto.MergedEntityMetadata{},
		matchingMetadataBuilder: newMatchingMetadataBuilder(),
		keepStandAlone:          true,
	}

	return builder
}

// Build creates and gets the MergedEntityMetadata object
func (builder *MergedEntityMetadataBuilder) Build() (*proto.MergedEntityMetadata, error) {
	matchingMetadata, err := builder.matchingMetadataBuilder.build()
	if err != nil {
		return nil, err
	}

	mergedEntityMetadata := &proto.MergedEntityMetadata{
		KeepStandalone: &builder.keepStandAlone,
		// Add the internal and external property matching metadata
		MatchingMetadata: matchingMetadata,
	}

	// Add commodities sold list
	if len(builder.commoditiesSold) > 0 {
		mergedEntityMetadata.CommoditiesSold = append(mergedEntityMetadata.CommoditiesSold, builder.commoditiesSold...)
	}

	// Add patched properties
	for _, propertyBuilder := range builder.propertyBuilders {
		mergedEntityMetadata.PatchedProperties = append(mergedEntityMetadata.PatchedProperties, propertyBuilder.build())
	}

	// Add patched fields
	for _, fieldBuilder := range builder.fieldBuilders {
		mergedEntityMetadata.PatchedFields = append(mergedEntityMetadata.PatchedFields, fieldBuilder.build())
	}

	// Add patched commodities bought
	for _, commodityBoughtMetadataBuilder := range builder.commBoughtMetadataList {
		mergedEntityMetadata.CommoditiesBought = append(mergedEntityMetadata.CommoditiesBought, commodityBoughtMetadataBuilder.build())
	}

	// Add patched commodities sold
	for _, commoditySoldMetadataBuilder := range builder.commoditiesSoldMetadataList {
		mergedEntityMetadata.CommoditiesSoldMetadata = append(mergedEntityMetadata.CommoditiesSoldMetadata, commoditySoldMetadataBuilder.build())
	}

	return mergedEntityMetadata, nil
}

// KeepInTopology indicates whether the entity reported by the probe should be kept in the topology or not if no
// stitching match is found.
// By default (if this function is not called) an entity is kept in the topology if no stitching match is found.
func (builder *MergedEntityMetadataBuilder) KeepInTopology(keepInTopology bool) *MergedEntityMetadataBuilder {
	builder.keepStandAlone = keepInTopology
	return builder
}

// InternalMatchingType specifies the type of the matching metadata to look for in the internal entity.
// Currently only MergedEntityMetadata_STRING and MergedEntityMetadata_LIST_STRING are supported.
// If MergedEntityMetadata_LIST_STRING is specified, InternalMatchingPropertyWithDelimiter() or
// InternalMatchingFieldWitDelimiter() must be called to explicitly set the delimiter that separates
// the list of strings
func (builder *MergedEntityMetadataBuilder) InternalMatchingType(returnType ReturnType) *MergedEntityMetadataBuilder {
	builder.matchingMetadataBuilder.internalReturnType = returnType
	return builder
}

// InternalMatchingProperty specifies the property name extracted from the internal entity's property map for matching.
func (builder *MergedEntityMetadataBuilder) InternalMatchingProperty(propertyName string) *MergedEntityMetadataBuilder {
	internal := &matchingData{
		propertyName: propertyName,
	}
	builder.matchingMetadataBuilder.addInternalMatchingData(internal)

	return builder
}

// InternalMatchingPropertyWithDelimiter specifies the property name extracted from the internal entity's property map for matching.
// The property value encodes a list of strings as a single string. The delimiter is used to separate out these strings.
func (builder *MergedEntityMetadataBuilder) InternalMatchingPropertyWithDelimiter(propertyName string,
	delimiter string) *MergedEntityMetadataBuilder {
	internal := &matchingData{
		propertyName: propertyName,
		delimiter:    delimiter,
	}
	builder.matchingMetadataBuilder.addInternalMatchingData(internal)

	return builder
}

// InternalMatchingProperty specifies the field name extracted from the internal entity for matching.
func (builder *MergedEntityMetadataBuilder) InternalMatchingField(fieldName string,
	fieldPaths []string) *MergedEntityMetadataBuilder {
	internal := &matchingData{
		fieldName:  fieldName,
		fieldPaths: fieldPaths,
	}
	builder.matchingMetadataBuilder.addInternalMatchingData(internal)

	return builder
}

// InternalMatchingProperty specifies the field name extracted from the internal entity for matching.
// The field value encodes a list of strings as a single string. The delimiter is used to separate out these strings.
func (builder *MergedEntityMetadataBuilder) InternalMatchingFieldWitDelimiter(fieldName string, fieldPaths []string,
	delimiter string) *MergedEntityMetadataBuilder {
	internal := &matchingData{
		fieldName:  fieldName,
		fieldPaths: fieldPaths,
		delimiter:  delimiter,
	}
	builder.matchingMetadataBuilder.addInternalMatchingData(internal)

	return builder
}

// InternalMatchingOid specifies the entity OID in XL repository for matching.
func (builder *MergedEntityMetadataBuilder) InternalMatchingOid() *MergedEntityMetadataBuilder {
	internal := &matchingData{
		useEntityOid: true,
	}
	builder.matchingMetadataBuilder.addInternalMatchingData(internal)

	return builder
}

// ExternalMatchingType specifies the type of the matching metadata to look for in the external entity.
// Currently only MergedEntityMetadata_STRING and MergedEntityMetadata_LIST_STRING are supported.
// If MergedEntityMetadata_LIST_STRING is specified, ExternalMatchingPropertyWithDelimiter() or
// ExternalMatchingFieldWithDelimiter() must be called to explicitly set the delimiter that separates
// the list of strings
func (builder *MergedEntityMetadataBuilder) ExternalMatchingType(returnType ReturnType) *MergedEntityMetadataBuilder {

	builder.matchingMetadataBuilder.externalReturnType = returnType
	return builder
}

// ExternalMatchingProperty specifies the property name extracted from the external entity's property map for matching.
func (builder *MergedEntityMetadataBuilder) ExternalMatchingProperty(propertyName string) *MergedEntityMetadataBuilder {
	external := &matchingData{
		propertyName: propertyName,
	}

	builder.matchingMetadataBuilder.addExternalMatchingData(external)

	return builder
}

// ExternalMatchingPropertyWithDelimiter specifies the property name extracted from the external entity's property map for matching.
// The property value encodes a list of strings as a single string. The delimiter is used to separate out these strings.
func (builder *MergedEntityMetadataBuilder) ExternalMatchingPropertyWithDelimiter(propertyName string,
	delimiter string) *MergedEntityMetadataBuilder {
	external := &matchingData{
		propertyName: propertyName,
		delimiter:    delimiter,
	}
	builder.matchingMetadataBuilder.addExternalMatchingData(external)

	return builder
}

// ExternalMatchingProperty specifies the field name extracted from the external entity for matching.
func (builder *MergedEntityMetadataBuilder) ExternalMatchingField(fieldName string,
	fieldPaths []string) *MergedEntityMetadataBuilder {
	external := &matchingData{
		fieldName:  fieldName,
		fieldPaths: fieldPaths,
	}
	builder.matchingMetadataBuilder.addExternalMatchingData(external)

	return builder
}

// ExternalMatchingProperty specifies the field name extracted from the external entity for matching.
// The field value encodes a list of strings as a single string. The delimiter is used to separate out these strings.
func (builder *MergedEntityMetadataBuilder) ExternalMatchingFieldWithDelimiter(fieldName string,
	fieldPaths []string, delimiter string) *MergedEntityMetadataBuilder {
	external := &matchingData{
		fieldName:  fieldName,
		fieldPaths: fieldPaths,
		delimiter:  delimiter,
	}
	builder.matchingMetadataBuilder.addExternalMatchingData(external)

	return builder
}

// ExternalMatchingOid specifies the entity OID in XL repository for matching.
func (builder *MergedEntityMetadataBuilder) ExternalMatchingOid() *MergedEntityMetadataBuilder {
	external := &matchingData{
		useEntityOid: true,
	}
	builder.matchingMetadataBuilder.addExternalMatchingData(external)

	return builder
}

// PatchProperty specifies the name of entity property that will be merged onto an external entity.
// The property will be searched from EntityDTO.propMap of the internal entity and written to the external entity.
// If the external entity already has a property with the same key, it will be replaced by the
// internal entity's property value.
// If there are multiple properties which need to be merged, this function should be called multiple times
// to specify multiple different property names.
func (builder *MergedEntityMetadataBuilder) PatchProperty(propertyName string) *MergedEntityMetadataBuilder {
	if propertyName == "" {
		return builder
	}
	builder.propertyBuilders = append(builder.propertyBuilders,
		newPropertyBuilder(propertyName))
	return builder
}

// PatchField specifies the name of the entity field and the path to reach that field in DTO, which will be
// merged onto an external entity. The field will be searched from EntityDTO of the internal entity.
// If the external entity already has a field with the same name and path, it will be replaced by the
// field from the internal entity.
// If there are multiple fields which need to be merged, this function should be called multiple times to
// specify multiple different fields.
func (builder *MergedEntityMetadataBuilder) PatchField(fieldName string, fieldPaths []string) *MergedEntityMetadataBuilder {
	if fieldName == "" {
		return builder
	}
	builder.fieldBuilders = append(builder.fieldBuilders,
		newFieldBuilder(fieldName, fieldPaths))
	return builder
}

// Deprecated: Use PatchSoldMetadata.
func (builder *MergedEntityMetadataBuilder) PatchSold(commType proto.CommodityDTO_CommodityType) *MergedEntityMetadataBuilder {
	builder.commoditiesSold = append(builder.commoditiesSold, commType)
	return builder
}

// Deprecated: Use PatchSoldMetadata, and call it multiple times to specify multiple commodities.
func (builder *MergedEntityMetadataBuilder) PatchSoldList(commType []proto.CommodityDTO_CommodityType) *MergedEntityMetadataBuilder {
	builder.commoditiesSold = append(builder.commoditiesSold, commType...)
	return builder
}

func (builder *MergedEntityMetadataBuilder) patchBoughtList(
	providerType proto.EntityDTO_EntityType,
	commType []proto.CommodityDTO_CommodityType,
	replacesProvider *proto.EntityDTO_EntityType) *MergedEntityMetadataBuilder {
	commBoughtMetadata := newCommodityBoughtMetadataBuilder(providerType)
	commBoughtMetadata.addBoughtList(commType)
	commBoughtMetadata.providerToReplace = replacesProvider
	builder.commBoughtMetadataList = append(builder.commBoughtMetadataList, commBoughtMetadata)
	return builder
}

// PatchBoughtList specifies the provider type and types of the bought commodities that need to be merged
// onto an external entity. Attributes defined in the internal bought commodity DTO from the specified provider
// type will overwrite those of the external entity.
func (builder *MergedEntityMetadataBuilder) PatchBoughtList(providerType proto.EntityDTO_EntityType,
	commType []proto.CommodityDTO_CommodityType) *MergedEntityMetadataBuilder {
	return builder.patchBoughtList(providerType, commType, nil)
}

// PatchBoughtAndReplaceProvider specifies the provider type and types of the bought commodities that need to be merged
// onto an external entity, and also sets the provider type of the external entity which will be replaced by current
// provider. Attributes defined in the internal bought commodity DTO from the specified provider type will overwrite
// those of the external entity.
func (builder *MergedEntityMetadataBuilder) PatchBoughtAndReplaceProvider(providerType proto.EntityDTO_EntityType,
	commType []proto.CommodityDTO_CommodityType, replacesProvider proto.EntityDTO_EntityType) *MergedEntityMetadataBuilder {
	return builder.patchBoughtList(providerType, commType, &replacesProvider)
}

func (builder *MergedEntityMetadataBuilder) patchSoldMetadata(
	commType proto.CommodityDTO_CommodityType, ignoreIfPresent bool,
	fields map[string][]string) *MergedEntityMetadataBuilder {
	commoditySoldMetadataBuilder := newCommoditySoldMetadataBuilder(commType, ignoreIfPresent)
	for name, paths := range fields {
		commoditySoldMetadataBuilder.addField(name, paths)
	}
	builder.commoditiesSoldMetadataList =
		append(builder.commoditiesSoldMetadataList, commoditySoldMetadataBuilder)
	return builder
}

// PatchSoldMetadata specifies the type of the sold commodity that needs to be merged onto an external entity, and
// defines the fields that should overwrite those of the external entity.
// If there are multiple sold commodities which need to be merged, this function should be called multiple times.
func (builder *MergedEntityMetadataBuilder) PatchSoldMetadata(commType proto.CommodityDTO_CommodityType,
	fields map[string][]string) *MergedEntityMetadataBuilder {
	return builder.patchSoldMetadata(commType, false, fields)
}

// PatchSoldMetadataIgnorePresent specifies the type of the sold commodity that needs to be merged onto an external
// entity, and defines the fields that should overwrite those of the external entity. If a commodity with the
// same type already exists in the external entity, but has a different key, the merge is skipped.
// If there are multiple sold commodities which need to be merged, this function should be called multiple times.
func (builder *MergedEntityMetadataBuilder) PatchSoldMetadataIgnorePresent(commType proto.CommodityDTO_CommodityType,
	fields map[string][]string) *MergedEntityMetadataBuilder {
	return builder.patchSoldMetadata(commType, true, fields)
}
