package membership

type DatagramBuilder struct {
	maxDatagramLength int
}

func NewDatagramBuilder(maxDatagramLength int) *DatagramBuilder {
	return &DatagramBuilder{
		maxDatagramLength: maxDatagramLength,
	}
}

func (b *DatagramBuilder) AppendToBuffer(buffer []byte, message Message, gossip *GossipQueue) ([]byte, int, error) {
	datagramBuffer, datagramN, err := message.AppendToBuffer(buffer)
	if err != nil {
		return buffer, 0, err
	}

	gossip.Prepare()
	for i := 0; i < gossip.Len(); i++ {
		gossipBuffer, gossipN, err := gossip.Get(i).AppendToBuffer(datagramBuffer)
		if err != nil {
			return buffer, 0, err
		}

		if len(datagramBuffer) > b.maxDatagramLength {
			break
		}

		gossip.MarkGossiped(i)
		datagramBuffer = gossipBuffer
		datagramN += gossipN
	}

	return datagramBuffer, datagramN, nil
}
