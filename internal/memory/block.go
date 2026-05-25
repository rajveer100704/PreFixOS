package memory

// BlockSize defines the fixed number of tokens managed per logical block.
const BlockSize = 16

// Block represents a Metadata Pointer to shared GPU/IPC memory.
// RadixKV acts strictly as the Control Plane: it does NOT store the heavy
// actual [num_heads, head_dim, block_size]float16 attention tensors.
// Instead, it manages the complex Radix tree layout and provides metadata pointers
// to the inference engine (vLLM, C++ backend) acting as the Data Plane.
type Block struct {
	ID int

	// Tokens stores the raw token IDs used strictly for prefix matching 
	// and divergence calculations within the Control Plane.
	Tokens [BlockSize]int32

	// MemoryOffset simulates the exact byte offset in a C++ mmap or CUDA 
	// shared memory space where the physical attention tensor resides.
	// This offset is passed via gRPC to the Data Plane engine so it 
	// can directly access and compute upon the physical VRAM.
	MemoryOffset int64
}
