package wal

import (
	"os"
	"encoding/binary"
    "encoding/json"
    "fmt"
    "hash/crc32"
	"bufio"
	"time"
	"sync"
)

const MaxFileSize int64 = 16 * 1024 * 1024 // 16 MB
const MaxBatchSize int64 = 4 * 1024 * 1024 // 4 MB

type iPayload struct {
	id    		string
	payload  	[]byte
}

type wResult struct {
	Value      int
	err        error
}

type wal struct {
	filePath     string
	Segment      int64
	batchExec    *time.Timer
	batch        []iPayload
	batchSize    int64
	mu 			 sync.Mutex
	wLock		 sync.Mutex
	pAck         map[string]wResult
	mapUpdate    sync.Mutex
}


func (wal *wal) CanInsert(dataSize int64) (bool, error) {

	filePath := wal.filePath

	info, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		err :=  wal.CreateNextFile()
		return false, err
	}

    if err != nil {
        return false, err
    }

    var currentSize int64
    if info != nil {
        currentSize = info.Size()
    }


	newFileSize := currentSize + dataSize

	if newFileSize >= MaxFileSize {
		err = wal.CreateNextFile()
	}

    return newFileSize < MaxFileSize, err
}

func (wal *wal) CreateNextFile() error {

	path := "../data/wal/" + fmt.Sprintf("%010d", wal.Segment + 1)

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)

	if err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	wal.filePath = path
	wal.Segment++

	return nil
}

func (wal *wal) RefreshBatch() error {

	wal.mu.Lock()
	
	if wal.batchExec != nil {
		wal.batchExec.Stop()
	}

	currBatch := wal.batch 
	wal.batch = []iPayload{}
	wal.batchSize = 0

	wal.batchExec = time.AfterFunc(5*time.Second, func() { wal.RefreshBatch() })

	wal.wLock.Lock()
	defer wal.wLock.Unlock()

	wal.mu.Unlock()

	return wal.BulkWrite(currBatch)
}



func (wal *wal) BulkWrite(batch []iPayload) error {

	file, err := os.OpenFile(wal.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644 )

	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	sizeTillNow := 0
	flushTill := len(batch)


	for i, data := range batch {

		ok, ierr := wal.CanInsert(int64(sizeTillNow + len(data.payload)))
		sizeTillNow += len(data.payload)

		if ierr != nil {
			return ierr
		}

		if !ok {
			flushTill = i
			break
		}
		
	}

	writeErrs := []error{}

	for i := 0; i < flushTill; i++ {

		data := batch[i]

		_, err := writer.Write(data.payload)

		if err != nil {
			writeErrs = append(writeErrs, err)
		} else {
			writeErrs = append(writeErrs, nil)
		}

	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return err
	}

	wal.mapUpdate.Lock()
	for i := 0; i < flushTill; i++ {
		data := batch[i]

		result := wal.pAck[data.id]

		if writeErrs[i] != nil {
			result.Value = 1
			result.err = writeErrs[i]
		} else {
			result.Value = 2
		}

		wal.pAck[data.id] = result
	}
	wal.mapUpdate.Unlock()

	if flushTill < len(batch) {
		return wal.BulkWrite(batch[flushTill : len(batch)])
	}

	return nil

}

func (wal *wal) Append(id string, payload any) error {

	bytes, err := wal.prepare(payload)

	if err != nil {
		return err
	}

	wal.mu.Lock()

	wal.batch = append(wal.batch, iPayload{id: id, payload: bytes})
	wal.batchSize += int64(len(bytes))

	wal.mapUpdate.Lock()
	wal.pAck[id] = wResult{
		Value: 0,
		err:   nil,
	}
	wal.mapUpdate.Unlock()

	if wal.batchSize >= MaxBatchSize {
		defer wal.RefreshBatch()
	} 
	
	wal.mu.Unlock()

	return nil
}


func (wal *wal) prepare(payload any) ([]byte, error) {

	data, err := json.Marshal(payload)

	if err != nil {
		return []byte{}, err
	}

	length := uint32(len(data))

	if uint64(length)+8 >= uint64(MaxFileSize) {
		return []byte{}, fmt.Errorf("payload exceeds maximum file size")
	}

	lengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBytes, length)


	checksumData := make([]byte, 0, 4+len(data))
	checksumData = append(checksumData, lengthBytes...)
	checksumData = append(checksumData, data...)

	checksum := crc32.Checksum(
		checksumData,
		crc32.MakeTable(crc32.Castagnoli),
	)

	checksumBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBytes, checksum)

	record := make([]byte, 0, 8+len(data))
    record = append(record, lengthBytes...)
    record = append(record, checksumBytes...)
    record = append(record, data...)

	return record, nil

}





func (wal *wal) Init() string {

	if wal.pAck == nil {
		wal.pAck = map[string]wResult{}
	}

	dir := "../data/wal"

	_, err := os.Stat(dir)

	if os.IsNotExist(err) {

		err := os.MkdirAll(dir, 0755)

		if err != nil {
			panic(err)
		}

		return "SUCCESSFUL"

	}

	if wal.batchExec == nil {
		wal.batchExec = time.AfterFunc(5*time.Second, func() { wal.RefreshBatch() })
	}

	if err != nil {
		panic(err)
	}

	return "ALREADY_EXISTS"

}