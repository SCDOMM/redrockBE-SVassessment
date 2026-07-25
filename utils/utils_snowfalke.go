package utils

import (
	"strings"
	"time"
)

var (
	machineID = GetMachineId()
	sf        = NewSnowflake(machineID)
)

const (
	timestampBits  = 41                         // 时间戳位数
	machineIDBits  = 10                         // 机器ID位数
	sequenceBits   = 12                         // 序列号位数
	maxMachineID   = -1 ^ (-1 << machineIDBits) // 最大机器ID
	maxSequenceNum = -1 ^ (-1 << sequenceBits)  // 最大序列号
)

type Snowflake struct {
	timestamp   int64
	machineID   int64
	sequenceNum int64
}

func NewSnowflake(machineID int64) *Snowflake {
	if machineID < 0 || machineID > maxMachineID {
		panic("Invalid machine ID")
	}
	return &Snowflake{
		timestamp:   time.Now().UnixNano() / 1e6,
		machineID:   machineID,
		sequenceNum: 0,
	}
}
func (s *Snowflake) GenerateID() int64 {
	currentTimestamp := time.Now().UnixNano() / 1e6
	if currentTimestamp == s.timestamp {
		s.sequenceNum = (s.sequenceNum + 1) & maxSequenceNum
		if s.sequenceNum == 0 {
			currentTimestamp = s.waitNextMillis()
		}
	} else {
		s.sequenceNum = 0
	}
	s.timestamp = currentTimestamp
	id := (currentTimestamp << (machineIDBits + sequenceBits)) |
		(s.machineID << sequenceBits) | s.sequenceNum
	return id
}
func (s *Snowflake) waitNextMillis() int64 {
	currentTimestamp := time.Now().UnixNano() / 1e6
	for currentTimestamp <= s.timestamp {
		currentTimestamp = time.Now().UnixNano() / 1e6
	}
	return currentTimestamp
}

func DecodeID(str string) string {
	return strings.Split(str, ":")[2]
}
