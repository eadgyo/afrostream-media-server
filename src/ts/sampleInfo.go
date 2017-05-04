package ts

// Hold information on Time, data location and type
type SampleInfo struct {
	PCR uint64
	CTS uint64
	DTS uint64

	isIFrameType bool

	mdatOffset int64
	mdatSize   uint32

	size uint32

	pesStream uint32

	hasPCR 		bool
	hasDTS		bool
	hasCTS		bool

	NALUnits []NALUnit
}

// Hold NAL unit location in the mdat
type NALUnit struct {
	mdatOffset int64
	mdatSize   uint32
}


func (info SampleInfo) IsIframe() (bool) {
	return info.isIFrameType
}

func (info SampleInfo) HasAdaptationField() (bool) {
	return info.hasPCR
}