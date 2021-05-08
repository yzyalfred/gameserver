package processor

import (
	"gameserver/common/errors"
	"gameserver/common/utils"
	"gameserver/core/log"
	utils2 "gameserver/core/utils"
	"github.com/golang/protobuf/proto"
	"reflect"
)

type MessageInfo struct {
	msgType    reflect.Type
	msgHandler MessageHandler
}

type MessageHandler func(clientId uint64, msg proto.Message)

type PBProcessor struct {
	msgInfoList map[uint16]*MessageInfo

	littleEndian bool
}

func NewPBProcessor() *PBProcessor {
	return &PBProcessor{
		msgInfoList:  make(map[uint16]*MessageInfo),
		littleEndian: true,
	}
}

func (this *PBProcessor) SetByteOrder(littleEndian bool) {
	this.littleEndian = littleEndian
}

func (this *PBProcessor) Route(clientId uint64, msgData []byte) {
	msgId := utils.ByteToUint16(msgData, this.littleEndian)
	msgInfo, ok := this.msgInfoList[msgId]
	if !ok {
		log.Warn("msgId not found: ", msgId)
		return
	}

	msg, err := this.Unmarshal(msgId, msgData)
	if err != nil {
		return
	}

	msgInfo.msgHandler(clientId, msg)
}

func (this *PBProcessor) Unmarshal(msgId uint16, msgData []byte) (proto.Message, error) {
	msgInfo, ok := this.msgInfoList[msgId]
	if !ok {
		log.Warn("msgId not found: ", msgId)
		return nil, errors.ERROR_NOT_FOUND
	}

	msgValue := reflect.New(msgInfo.msgType.Elem()).Interface()
	msg := msgValue.(proto.Message)
	err := proto.Unmarshal(msgData[2:], msg)
	if err != nil {
		log.Warn("unmarshall error, msgId: ", msgId, "data: ", msgData)
		return nil, err
	}

	return msg, nil
}

// add head: msgId
func (this *PBProcessor) Marshal(msgId uint16, msg proto.Message) ([]byte, error) {
	msgData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, len(msgData)+utils2.MSG_ID_LEN)
	utils.PutUint16ToByte(buf, msgId, this.littleEndian)
	copy(buf[:utils2.MSG_ID_LEN], msgData)

	return buf, nil
}

// add head: clientId + msgId
func (this *PBProcessor) MarshalServerMsg(msgId uint16, clientId uint32, msg proto.Message) ([]byte, error) {
	msgData, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, len(msgData)+utils2.SERVER_MSG_HEAD_LEN)
	utils.PutUint32ToByte(buf, clientId, this.littleEndian)
	utils.PutUint16ToByte(buf[utils2.CLIENT_ID_LEN:], msgId, this.littleEndian)
	copy(buf[:utils2.SERVER_MSG_HEAD_LEN], msgData)

	return buf, nil
}

func (this *PBProcessor) Register(msgId uint16, msg proto.Message, msgHandler MessageHandler) {
	reflectType := reflect.TypeOf(msg)
	this.msgInfoList[msgId] = &MessageInfo{
		msgType:    reflectType,
		msgHandler: msgHandler,
	}
}
