package ts

import (
	"mp4"
)


func CreateHLSFragmentWithConf(dConf mp4.Conf, filename string, fragmentNumber uint32, fragmentDuration uint32) ([]byte) {

	// Variables data used to create our modifiedFragment
	modifiedFragment := FragmentData{}

	// 1) analyse the stream and found get main information
	streamInfo := AnalyseStream(dConf, filename)

	// 2) Create program packets
	RegisterProgramPackets(streamInfo, modifiedFragment)

	// 3) Retrieve information on the created modifiedFragment
	fragmentInfo := GetFragmentInfo(streamInfo, fragmentNumber, fragmentDuration)

	// 4) Retrieve information on all contained samples
	samplesInfo := GetSamplesInfo(streamInfo, fragmentInfo, modifiedFragment)

	// 5) Create PES packets
	RegisterStreamPackets(streamInfo, samplesInfo, modifiedFragment)

	// 6) Create our fragment assembling all created packets
	return FinaliseFragment(modifiedFragment)
}

