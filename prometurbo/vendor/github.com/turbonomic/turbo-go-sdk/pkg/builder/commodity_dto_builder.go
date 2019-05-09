package builder

import "github.com/turbonomic/turbo-go-sdk/pkg/proto"

type CommodityDTOBuilder struct {
	commodityType           *proto.CommodityDTO_CommodityType
	key                     *string
	used                    *float64
	reservation             *float64
	capacity                *float64
	limit                   *float64
	peak                    *float64
	active                  *bool
	resizable               *bool
	displayName             *string
	thin                    *bool
	computedUsed            *bool
	usedIncrement           *float64
	propMap                 map[string][]string
	isUsedPct               *bool
	utilizationThresholdPct *float64
	pricingMetadata         *proto.CommodityDTO_PricingMetadata

	storageLatencyData    *proto.CommodityDTO_StorageLatencyData
	storageAccessData     *proto.CommodityDTO_StorageAccessData
	vstoragePartitionData *proto.VStoragePartitionData

	vMemData *proto.CommodityDTO_VMemData
	vCpuData *proto.CommodityDTO_VCpuData

	err error
}

func NewCommodityDTOBuilder(commodityType proto.CommodityDTO_CommodityType) *CommodityDTOBuilder {
	return &CommodityDTOBuilder{
		commodityType: &commodityType,
	}
}

func (cb *CommodityDTOBuilder) Create() (*proto.CommodityDTO, error) {
	if cb.err != nil {
		return nil, cb.err
	}
	commodityDTO := &proto.CommodityDTO{
		CommodityType: cb.commodityType,
		Key:           cb.key,
		Used:          cb.used,
		Reservation:   cb.reservation,
		Capacity:      cb.capacity,
		Limit:         cb.limit,
		Peak:          cb.peak,
		Active:        cb.active,
		Resizable:     cb.resizable,
		DisplayName:   cb.displayName,
		Thin:          cb.thin,
		ComputedUsed:  cb.computedUsed,
		UsedIncrement: cb.usedIncrement,
		PropMap:       buildPropertyMap(cb.propMap),
	}

	if cb.storageLatencyData != nil {
		commodityDTO.CommodityData = &proto.CommodityDTO_StorageLatencyData_{cb.storageLatencyData}
	} else if cb.storageAccessData != nil {
		commodityDTO.CommodityData = &proto.CommodityDTO_StorageAccessData_{cb.storageAccessData}
	} else if cb.vstoragePartitionData != nil {
		commodityDTO.CommodityData = &proto.CommodityDTO_VstoragePartitionData{cb.vstoragePartitionData}
	}

	if cb.vCpuData != nil {
		commodityDTO.HotresizeData = &proto.CommodityDTO_VcpuData{cb.vCpuData}
	} else if cb.vMemData != nil {
		commodityDTO.HotresizeData = &proto.CommodityDTO_VmemData{cb.vMemData}
	}

	return commodityDTO, nil
}

func (cb *CommodityDTOBuilder) Key(key string) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.key = &key
	return cb
}

func (cb *CommodityDTOBuilder) Capacity(capacity float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.capacity = &capacity
	return cb
}

func (cb *CommodityDTOBuilder) Used(used float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.used = &used
	return cb
}

func (cb *CommodityDTOBuilder) Peak(peak float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.peak = &peak
	return cb
}

func (cb *CommodityDTOBuilder) Reservation(reservation float64) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.reservation = &reservation
	return cb
}

func (cb *CommodityDTOBuilder) Resizable(resizable bool) *CommodityDTOBuilder {
	if cb.err != nil {
		return cb
	}
	cb.resizable = &resizable
	return cb
}

func buildPropertyMap(propMap map[string][]string) []*proto.CommodityDTO_PropertiesList {
	if propMap == nil {
		return nil
	}
	propList := []*proto.CommodityDTO_PropertiesList{}
	for name, values := range propMap {
		propList = append(propList, &proto.CommodityDTO_PropertiesList{
			Name:   &name,
			Values: values,
		})
	}
	return propList
}
