package mqttbase

import "errors"

// TopicFilter - TopicFilter  structure
type TopicFilter struct {
	Filter string
	Qos    byte
}

// SubscribePacket - Subscribe packet structure
type SubscribePacket struct {
	FixedHeader *FixedHeader
	Id          uint16
	Topics      []*TopicFilter
}

// PacketType - Returns packet type
func (p *SubscribePacket) PacketType() byte {
	return Subscribe
}

// NewSubscribePacket - Creates a new Subscribe Packet
func NewSubscribePacket() *SubscribePacket {
	packet := new(SubscribePacket)
	packet.FixedHeader = new(FixedHeader)
	packet.FixedHeader.cntrlPacketType = Subscribe
	packet.FixedHeader.remaingLength = 2
	// TODO - Fix this.......
	packet.Topics = make([]*TopicFilter, 0, 32)
	return packet
}

func (p *SubscribePacket) AddTopic(topic *TopicFilter) {
	p.Topics = append(p.Topics, topic)
}

func (p *SubscribePacket) Marshal() ([]byte, error) {
	totalLen := 0
	for _, topic := range p.Topics {
		// lentgh of string plus 2 (len) + qos
		totalLen = totalLen + len(topic.Filter) + 3
	}
	p.FixedHeader.remaingLength = uint32(totalLen + 2)
	fixedHeader := p.FixedHeader.Marshal()
	// 2 is variable header
	data := make([]byte, 0, len(fixedHeader)+totalLen+2)
	data = append(data, fixedHeader...)
	// append ID
	data = append(data, byte(p.Id>>8))
	data = append(data, byte(p.Id))

	for _, topic := range p.Topics {
		str, _ := EncodeString(topic.Filter)
		data = append(data, str...)
		data = append(data, topic.Qos)
	}
	return data, nil
}

func (p *SubscribePacket) unmarshal(data []byte) error {
	var err error
	fh := new(FixedHeader)
	fh.unmarshal(data)
	p.FixedHeader = fh
	if fh.remaingLength < 2 {
		return errors.New("No subscribe packet id")
	}
	p.Id = uint16(data[2])<<8 | uint16(data[3])
	// TODO - Fix this.......
	p.Topics = make([]*TopicFilter, 0, 32)
	pos := 4
	for uint32(pos) < fh.remaingLength+2 {
		topic := new(TopicFilter)
		topic.Filter, err = UnencodeString(data[pos:])
		if err != nil {
			return err
		}
		pos = pos + len(topic.Filter) + 2
		topic.Qos = data[pos]
		p.AddTopic(topic)
		pos++
	}
	return nil
}
