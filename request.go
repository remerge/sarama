package sarama

import (
	"encoding/binary"
	"fmt"
	"io"
)

type protocolBody interface {
	encoder
	versionedDecoder
	key() int16
	version() int16
	requiredVersion() KafkaVersion
}

type request struct {
	correlationID int32
	clientID      string
	body          protocolBody
}

func (r *request) encode(pe packetEncoder) (err error) {
	pe.push(&lengthField{})
	pe.putInt16(r.body.key())
	pe.putInt16(r.body.version())
	pe.putInt32(r.correlationID)
	err = pe.putString(r.clientID)
	if err != nil {
		return err
	}
	err = r.body.encode(pe)
	if err != nil {
		return err
	}
	return pe.pop()
}

func (r *request) decode(pd packetDecoder) (err error) {
	var key int16
	if key, err = pd.getInt16(); err != nil {
		return err
	}
	var version int16
	if version, err = pd.getInt16(); err != nil {
		return err
	}
	if r.correlationID, err = pd.getInt32(); err != nil {
		return err
	}
	r.clientID, err = pd.getString()

	r.body = allocateBody(key, version)
	if r.body == nil {
		return PacketDecodingError{fmt.Sprintf("unknown request key (%d)", key)}
	}
	return r.body.decode(pd, version)
}

func decodeRequest(r io.Reader) (req *request, bytesRead int, err error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBytes); err != nil {
		return nil, bytesRead, err
	}
	bytesRead += len(lengthBytes)

	length := int32(binary.BigEndian.Uint32(lengthBytes))
	if length <= 4 || length > MaxRequestSize {
		return nil, bytesRead, PacketDecodingError{fmt.Sprintf("message of length %d too large or too small", length)}
	}

	encodedReq := make([]byte, length)
	if _, err := io.ReadFull(r, encodedReq); err != nil {
		return nil, bytesRead, err
	}
	bytesRead += len(encodedReq)

	req = &request{}
	if err := decode(encodedReq, req); err != nil {
		return nil, bytesRead, err
	}
	return req, bytesRead, nil
}

const produceCommandNumber = 0
const fetchCommandNumber = 1
const offsetCommandNumber = 2
const metaCommandNumber = 3
const offsetCommitCommandNumber = 8
const offsetFetchCommandNumber = 9
const consumerMetadataCommandNumber = 10
const joinGroupCommandNumber = 11
const heartbeatCommandNumber = 12
const leaveGroupCommandNumber = 13
const syncGroupCommandNumber = 14
const describeGroupsCommandNumber = 15
const listGroupsCommandNumber = 16
const saslHandshakeCommandNumber = 17
const apiVersionCommandNumber = 18

func allocateBody(key, version int16) protocolBody {
	switch key {
	case produceCommandNumber:
		return &ProduceRequest{}
	case fetchCommandNumber:
		return &FetchRequest{}
	case offsetCommandNumber:
		return &OffsetRequest{Version: version}
	case metaCommandNumber:
		return &MetadataRequest{}
	case offsetCommitCommandNumber:
		return &OffsetCommitRequest{Version: version}
	case offsetFetchCommandNumber:
		return &OffsetFetchRequest{}
	case consumerMetadataCommandNumber:
		return &ConsumerMetadataRequest{}
	case joinGroupCommandNumber:
		return &JoinGroupRequest{}
	case heartbeatCommandNumber:
		return &HeartbeatRequest{}
	case leaveGroupCommandNumber:
		return &LeaveGroupRequest{}
	case syncGroupCommandNumber:
		return &SyncGroupRequest{}
	case describeGroupsCommandNumber:
		return &DescribeGroupsRequest{}
	case listGroupsCommandNumber:
		return &ListGroupsRequest{}
	case saslHandshakeCommandNumber:
		return &SaslHandshakeRequest{}
	case apiVersionCommandNumber:
		return &ApiVersionsRequest{}
	}
	return nil
}
