package utils

import (
	"strings"
	"unicode/utf8"
)

type TextChunker struct {
	ChunkSize     int // 每块字符数
	ChunkOverlap  int // 重叠字符数
	LineChunkSize int //按照行数分割
}

func NewTextChunker(chunkSize, overlap, lineChunSize int) *TextChunker {
	return &TextChunker{
		ChunkSize:     chunkSize,
		ChunkOverlap:  overlap,
		LineChunkSize: lineChunSize,
	}
}

// SplitByLines 按行数切割代码
func (tc *TextChunker) SplitByLines(text string) ([]string, []int64, []int64) {
	lines := strings.Split(text, "\n")
	var chunks []string
	var startLines []int64
	var endLines []int64
	if tc.LineChunkSize <= 0 {
		tc.LineChunkSize = 50 // 默认每块50行
	}
	for i := 0; i < len(lines); i += tc.LineChunkSize {
		end := i + tc.LineChunkSize
		if end > len(lines) {
			end = len(lines)
		}
		chunkText := strings.Join(lines[i:end], "\n")
		chunks = append(chunks, chunkText)
		startLines = append(startLines, int64(i)) // 起始行号
		endLines = append(endLines, int64(end))   // 结束行号
	}
	return chunks, startLines, endLines
}

// SplitByParagraph 按照段落切割文档
func (tc *TextChunker) SplitByParagraph(text string) []string {
	paragraphs := strings.Split(text, "\n\n")
	var chunks []string
	currentChunk := ""
	for _, para := range paragraphs {
		//去除首尾空白
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		//计算字符个数
		if utf8.RuneCountInString(currentChunk+para) <= tc.ChunkSize {
			//分段
			if currentChunk != "" {
				currentChunk += "\n\n"
			}
			currentChunk += para
		} else {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			// 如果单段太长，按照大小切割
			if utf8.RuneCountInString(para) > tc.ChunkSize {
				chunks = append(chunks, tc.splitBySize(para)...)
			} else {
				currentChunk = para
			}
		}
	}
	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}
	return chunks
}

// splitBySize 纯按大小切
func (tc *TextChunker) splitBySize(text string) []string {
	runes := []rune(text)
	var chunks []string
	start := 0
	for start < len(runes) {
		end := start + tc.ChunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
		start = end - tc.ChunkOverlap
		if start < 0 {
			start = 0
		}
	}
	return chunks
}
