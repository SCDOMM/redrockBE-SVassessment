package model

type MilvusData struct {
	Id        string
	Vectors   []float32
	Source    string
	Path      string
	Language  string
	StartLine int
	EndLine   int
}
