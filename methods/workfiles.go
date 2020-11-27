package methods

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type FileTable struct {
	FileName   string
	IsDir      uint32
	Offset     uint32
	HeadOffset int64
	Size       uint32
	ArcNum     uint32
	NameOff    uint32
	FileID     []byte
}

func Repack(table []FileTable, FilePath string, header []byte) {
	var Size uint32
	var FileOffset uint32
	var ArcNum int
	ArcNum = 0
	FileOffset = 0
	Size = 0

	file, err := os.Create(strings.ReplaceAll(FilePath, ".toc", ".M"+fmt.Sprintf("%02d", ArcNum)))

	if err != nil {
		panic(err)
	}

	defer file.Close()

	for i := 0; i < len(table); i++ {
		table[i].FileName = strings.ReplaceAll(table[i].FileName, "//", "/")

		if Size+table[i].Size > 4294967295 {
			ArcNum++
			FileOffset = 0
			Size = 0

			file.Close()

			file, err = os.Create(strings.ReplaceAll(FilePath, ".toc", ".M"+fmt.Sprintf("%02d", ArcNum)))
			if err != nil {
				panic(err)
			}

			defer file.Close()
		}

		var fileInfo os.FileInfo
		fileInfo, err = os.Stat(strings.ReplaceAll(FilePath, ".toc", "") + "/" + table[i].FileName)
		if err != nil {
			panic(err)
		}

		tmp := make([]byte, 8)
		binary.LittleEndian.PutUint64(tmp, uint64(fileInfo.Size()))

		table[i].Size = binary.LittleEndian.Uint32(tmp[:4])

		table[i].Offset = FileOffset
		table[i].ArcNum = uint32(ArcNum)
		Size += table[i].Size
		FileOffset += table[i].Size

		tmp = make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, table[i].ArcNum)
		copy(header[table[i].HeadOffset+8:], tmp)

		tmp = make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, table[i].Size)
		copy(header[table[i].HeadOffset+16:], tmp)

		tmp = make([]byte, 4)
		binary.LittleEndian.PutUint32(tmp, table[i].Offset)
		copy(header[table[i].HeadOffset+12:], tmp)

		tmp = nil

		read, err := ioutil.ReadFile(strings.ReplaceAll(FilePath, ".toc", "") + "/" + table[i].FileName)

		if err != nil {
			panic(err)
		}

		file.Write(read)
		fmt.Printf("HeadOff: %d\tOff: %d\tSize: %d\tFileName: %s\n", table[i].HeadOffset, table[i].Offset, table[i].Size, table[i].FileName)
	}

	header = EncHeader(header)

	file, err = os.Create(FilePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Write(header)
}

func Unpack(table []FileTable, FilePath string) {
	ArcFilePath := strings.ReplaceAll(FilePath, ".toc", ".M")

	err := os.MkdirAll(strings.ReplaceAll(ArcFilePath, ".M", ""), 0666)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(table); i++ {
		file, err := os.Open(ArcFilePath + fmt.Sprintf("%02d", table[i].ArcNum))
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()

		block := make([]byte, table[i].Size)

		off, err := file.Seek(int64(table[i].Offset), 0)
		//fmt.Printf("File.Seek: %d\n", off)
		_, err = file.ReadAt(block, off)

		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()

		Dir := filepath.Dir(strings.ReplaceAll(ArcFilePath, ".M", "") + "/" + table[i].FileName)
		_, err = os.Stat(Dir)

		if os.IsNotExist(err) {
			os.MkdirAll(Dir, 0666)
		}

		file, err = os.Create(strings.ReplaceAll(ArcFilePath, ".M", "") + "/" + table[i].FileName)

		if err != nil {
			panic(err)
		}

		defer file.Close()

		_, err = file.Write(block)

		if err != nil {
			panic(err)
		}

		defer file.Close()

		table[i].FileName = strings.ReplaceAll(table[i].FileName, "//", "/")
		fmt.Printf("HeadOff: %d\tOff: %d\tSize: %d\tFileName: %s\n", table[i].HeadOffset, table[i].Offset, table[i].Size, table[i].FileName)

		block = nil
	}
}