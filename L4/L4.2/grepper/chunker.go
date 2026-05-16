package grepper

import (
	"bufio"
	"os"
)

type Chunk struct {
	Start    int64
	End      int64
	FileName string
}

// SplitFile — разбивает файл на примерно равные части
func SplitFile(filePath string, numChunks int) ([]Chunk, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	fileSize := info.Size()
	if fileSize == 0 {
		// Пустой файл — один пустой кусок
		return []Chunk{{Start: 0, End: 0, FileName: filePath}}, nil
	}

	chunkSize := fileSize / int64(numChunks)
	if chunkSize == 0 {
		chunkSize = fileSize
	}

	var chunks []Chunk
	for i := 0; i < numChunks; i++ {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1
		if i == numChunks-1 {
			end = fileSize - 1
		}
		chunks = append(chunks, Chunk{Start: start, End: end, FileName: filePath})
	}

	for i := 1; i < len(chunks); i++ {
		chunks[i].Start = adjustToNewline(filePath, chunks[i].Start)
	}

	return chunks, nil
}

// adjustToNewline — сдвигает позицию вперёд до ближайшего перевода строки
func adjustToNewline(filePath string, pos int64) int64 {
	if pos <= 0 {
		return 0
	}

	f, err := os.Open(filePath)
	if err != nil {
		return pos
	}
	defer f.Close()

	_, err = f.Seek(pos, 0)
	if err != nil {
		return pos
	}

	// Читаем байт за байтом, пока не встретим '\n'
	reader := bufio.NewReader(f)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			break
		}
		pos++
		if b == '\n' {
			break // можно начинать читать со следующего байта
		}
	}
	return pos
}
