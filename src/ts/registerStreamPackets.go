package ts

// Create Elementary stream packets containing our stream src
func RegisterStreamPackets(streamInfo StreamInfo, samplesInfo []SampleInfo, fragment *FragmentData) {

	// For each sample
	for _, sample := range samplesInfo {
		// Create the elementary stream
		elementaryStream := CreateElementaryStream(streamInfo, sample)

		// Create packets stream
		pes := createPackets(streamInfo, sample, uint32(len(elementaryStream)))

		// Fill packets payload
		fillPackets(&pes, elementaryStream)

		// Append fragment to PES
		fragment.pes = append(fragment.pes, pes...)
	}
}

func CreateElementaryStream(stream StreamInfo, sample SampleInfo) ([]byte) {

	size := 24 + 8 + 16 // startCode + stream id + ps packet length

	// Optional PES Header
	size += 16 + 8 // Optional + PES Header length + PTS + DTS

	flag7 := uint32(0)

	// If CTS needed
	if sample.hasCTS {
		size += 33
		flag7 |= 1 << 7
	}

	// If DTS needed
	if sample.hasDTS {
		size += 33
		flag7 |= 1 << 6
	}

	headerLength := RoundDivision32(uint32(size), 8)  // rounded to upper byte
	packetLength := headerLength + sample.size

	data := NewData(int(packetLength))
	data.PushUInt(1, 24)
	data.PushUInt(stream.streamId, 8)
	data.PushUInt(sample.size, 16)
	data.PushUInt(1, 1)
	data.PushUInt(0, 7)
	data.PushUInt(flag7, 8)
	data.PushUInt(headerLength - 8, 8)

	// If CTS needed
	if sample.hasCTS {
		pushTimestamp(sample.CTS, data)
	}

	// If DTS needed
	if sample.hasDTS {
		pushTimestamp(sample.DTS, data)
	}

	// Fill to the end of header
	data.FillTo(0xff, int(headerLength * 8))

	// Fill with the sample data
	stream.mdat.Offset = sample.mdatOffset
	stream.mdat.Size = sample.mdatSize
	data.PushAll(stream.mdat.ToBytes())
	return data.Data
}

func pushTimestamp(timestamp uint64, data *Data) {
	data.PushUInt64(timestamp >> 30, 3) 	// timestamp [32..30]
	data.PushUInt(1, 1) 			// marker_bit
	data.PushUInt64(timestamp >> 15, 15) 	// timestamp [29..15]
	data.PushUInt(1, 1) 			// marker_bit
	data.PushUInt64(timestamp, 15) 		// timestamp [14..0]
	data.PushUInt(1, 1) 			// marker_bit
}

func createPackets(info StreamInfo, sample SampleInfo, elementaryStreamSize uint32) (packets []PES){
	// Create the first fragment with adaptation field
	pid := info.PID
	streamId := info.streamId

	// Create the first packet
	var firstPacket PES = *NewStartStream(pid, streamId)
	if sample.HasAdaptationField() {
		firstPacket.AdaptationFieldControl = 0x03
	} else {
		firstPacket = *NewStream(pid)
	}

	if sample.hasPCR {
		pcr := PCR{}
		pcr.BaseMediaDecodeTime = sample.PCR
		firstPacket.setPCR(pcr)
	}

	// IF isIFrame set RAP
	if sample.IsIframe() {
		firstPacket.RandomAccessIndicator = 1
	}
	firstPacket.setAdaptationControl(true, true)

	// Compute the number of fragments needed
	restingSize := uint32(188 - firstPacket.RestingSize())
	firstPacket.Payload.EmptySize = restingSize

	neededPackets := uint32(0)
	if restingSize < elementaryStreamSize {
		neededPackets = RoundDivision32(elementaryStreamSize - restingSize, 184) // rounded
	}

	// Create packets
	packets = make([]PES, neededPackets + 1)
	packets[0] = firstPacket

	for i := uint32(1); i < neededPackets + 1; i++ {
		packets[i] = *NewStream(pid)
		packets[i].setAdaptationControl(false, true)
		packets[i].Payload.EmptySize = 184
	}

	return
}

func fillPackets(packets *[]PES, elementaryStream []byte) {

	offset := uint32(0)
	finalSize := uint32(len(elementaryStream))
	packetId := uint32(0)

	var extractedSize uint32
	var payloadSize uint32

	// While there is data left
	for offset != finalSize {
		// Get the corresponding part of the packet
		payloadSize = uint32((*packets)[packetId].Payload.Size())
		extractedSize = Min32(payloadSize, finalSize - offset)

		// Register the payload
		(*packets)[packetId].Data = elementaryStream[offset:offset+extractedSize]

		// Go to next packet
		offset += extractedSize
		packetId++
	}

	// Fill the end of the last packet
	if packetId != 0 {
		packetId--
	}
	lastRestingSize := payloadSize - extractedSize
	if lastRestingSize > 0  {
		fillingBytes := NewData(int(lastRestingSize*8))
		fillingBytes.FillRemaining(0xff)
		(*packets)[packetId].Data = append((*packets)[packetId].Data, fillingBytes.Data...)
	}

}