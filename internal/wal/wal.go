package wal

import (
	"os"
	"encoding/binary"
    "encoding/json"
    "fmt"
    "hash/crc32"
	"encoding/binary"
	"time"
	"sync"
)

const MaxFileSize int64 = 16 * 1024 * 1024 // 16 MB
const MaxBatchSize int64 = 4 * 1024 * 1024 // 4 MB

type wal struct {
	filePath     string
	Segment      int64
	batchExec    *time.Timer
	batch        [][]bytes
	batchSize    int64
	mu 			 sync.Mutex
	wLock		 sync.Mutex
}


func (* wal) CanInsert(dataSize int64) bool, error {

	filePath := wal.filePath

	info, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		err :=  CreateNextFile()
		return false, err
	}

    if err != nil {
        return false, err
    }

    var currentSize int64
    if info != nil {
        currentSize = info.Size()
    }


	newFileSize := currentSize + dataSize + 8

	if newFileSize >= MaxFileSize {
		err = CreateNextFile()
	}

    return newFileSize >= MaxFileSize, err
}

func (* wal) CreateNextFile() error {

	path := "../data/wal/" + fmt.Sprintf("%10d", wal.Segment + 1)

	newFile := file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)

	wal.filePath = path
	wal.Segment++

}

func (* wal) RefreshBatch() error {

	wal.mu.Lock()
	wal.RefreshBatchIn()

}

func (* wal) RefreshBatchIn() error {

	wal.batchExec = nil
	currBatch := wal.batch 
	wal.batch = [][]bytes{}
	wal.batchSize = 0
	wal.wLock.Lock()
	defer wal.wLock.Unlock()
	wal.mu.Unlock()
	BulkWrite(currBatch)

}

func (* wal) BulkWrite(batch [][]bytes) []error, error {

	var wError []error

	file, err := os.OpenFile(wal.filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644 )

	if err != nil {
		return []error{}, err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)


	for _, data := range batch {

		ok, ierr := CanInsert(int64(len(data)))

		if ierr != nil {
			wError = append(wError, err)
		}

		if !ok {
			
		}

		_, err := writer.Write(data)
		wError = append(wError, err)

	}

	if err := writer.Flush(); err != nil {
		return []error{}, err
	}

	return wError, nil

}

func (* wal) Append(payload any) error {

	bytes, err := prepare(payload)

	wal.mu.Lock()
	if wal.batchExec == nil {
		wal.batchExec = time.AfterFunc(5*time.Second, wal.RefreshBatch)
	}

	wal.batch = append(wal.batch, bytes)
	wal.batchSize += int64(len(bytes))

	if wal.batchSize >= MaxFileSize {
		wal.batchExec.Stop()
		wal.RefreshBatchIn()
	} else {
		wal.mu.Unlock()
	}

}


func (* wal) prepare(payload any) []byte, error {

	data, err := json.Marshal(payload)

	if err != nil {
		return []byte{}, err
	}

	length := uint32(len(data))

	if uint64(length) > uint64(MaxFileSize) {
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


	record := make([]byte, 0, 8+len(data))
    record = append(record, lengthBytes...)
    record = append(record, checksumBytes...)
    record = append(record, data...)

	return record, nil

}





func Init() string {

	dir := "../data/wal"

	info, err := os.Stat(dir)

	if os.IsNotExist(err) {

		err := os.MkdirAll(dir, 0755)

		if err != nil {
			panic(err)
		}

		return "SUCCESSFUL"

	}

	if err != nil {
		panic(err)
	}

	return "ALREADY_EXISTS"

}