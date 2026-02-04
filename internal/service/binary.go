package service

import (
	"encoding/binary"
	"time"
)

// MarshalBinary 序列化ThreadDTO
func (dto *ThreadDTO) MarshalBinary() ([]byte, error) {
	buf := make([]byte, 0)

	// tid (8 bytes)
	tidBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(tidBuf, uint64(dto.Tid))
	buf = append(buf, tidBuf...)

	// fid (4 bytes)
	fidBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(fidBuf, uint32(dto.Fid))
	buf = append(buf, fidBuf...)

	// uid (8 bytes)
	uidBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(uidBuf, uint64(dto.Uid))
	buf = append(buf, uidBuf...)

	// subject length (2 bytes) + subject
	subjLen := len(dto.Subject)
	subjLenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(subjLenBuf, uint16(subjLen))
	buf = append(buf, subjLenBuf...)
	buf = append(buf, []byte(dto.Subject)...)

	// views (4 bytes)
	viewsBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(viewsBuf, uint32(dto.Views))
	buf = append(buf, viewsBuf...)

	// replies (4 bytes)
	repliesBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(repliesBuf, uint32(dto.Replies))
	buf = append(buf, repliesBuf...)

	// dateline (4 bytes)
	dateBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(dateBuf, uint32(dto.Dateline))
	buf = append(buf, dateBuf...)

	// lastpost (4 bytes)
	lastBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lastBuf, uint32(dto.Lastpost))
	buf = append(buf, lastBuf...)

	// status (1 byte)
	buf = append(buf, byte(dto.Status))

	// message length (2 bytes) + message
	msgLen := len(dto.Message)
	msgLenBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(msgLenBuf, uint16(msgLen))
	buf = append(buf, msgLenBuf...)
	buf = append(buf, []byte(dto.Message)...)

	return buf, nil
}

// UnmarshalBinary 反序列化ThreadDTO
func (dto *ThreadDTO) UnmarshalBinary(data []byte) error {
	offset := 0

	// tid
	dto.Tid = int64(binary.BigEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// fid
	dto.Fid = int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// uid
	dto.Uid = int64(binary.BigEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// subject length
	subjLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// subject
	dto.Subject = string(data[offset : offset+subjLen])
	offset += subjLen

	// views
	dto.Views = int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// replies
	dto.Replies = int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// dateline
	dto.Dateline = int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// lastpost
	dto.Lastpost = int(binary.BigEndian.Uint32(data[offset : offset+4]))
	offset += 4

	// status
	dto.Status = int(data[offset])
	offset += 1

	// message length
	msgLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
	offset += 2

	// message
	if msgLen > 0 {
		dto.Message = string(data[offset : offset+msgLen])
	}

	return nil
}

// Now 获取当前Unix时间
func Now() int64 {
	return time.Now().Unix()
}
