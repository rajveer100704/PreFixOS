package interfaces

// Tokenizer defines token sequence encoding and decoding contracts
type Tokenizer interface {
	Encode(text string) ([]int32, error)
	Decode(tokens []int32) (string, error)
}
