package lzham

const (
	cBitBufSize = 64

	cArithSym       = -1
	cAlignToByteSym = -2
	cArithInit      = -3

	cNull = iota - 4
	cEncoding
	cDecoding
)

type output_symbol struct {
	bits uint32

	num_bits int16

	arith_prob0 uint16
}

type symbol_codec struct {
	pDecode_buf      []byte
	pDecode_buf_next []byte
	pDecode_buf_end  []byte
	decode_buf_size  uint64
	decode_buf_eof   bool

	pDecode_need_bytes_func func(num_bytes_consumed uint64, pPrivate_data []byte, pBuf []byte, buf_size uint64, eof_flag bool)
	pDecode_private_data    []byte

	bit_buf   uint64
	bit_count int32

	total_model_updates uint32

	output_buf       []uint8
	arith_output_buf []uint8

	output_syms []output_symbol

	total_bits_written uint32

	arith_base       uint32
	arith_value      uint32
	arith_length     uint32
	arith_total_bits uint32

	// quasi_adaptive_huffman_data_model *m_pSaved_huff_model fuck this shit
	// void                              *m_pSaved_model
	saved_node_index uint32

	mode uint32
}

func (sc *symbol_codec) reset() {
	sc.pDecode_buf = nil
	sc.pDecode_buf_next = nil
	sc.pDecode_buf_end = nil
	sc.decode_buf_size = 0

	sc.bit_buf = 0
	sc.bit_count = 0
	sc.total_model_updates = 0
	sc.mode = cNull
	sc.total_bits_written = 0

	sc.arith_base = 0
	sc.arith_value = 0
	sc.arith_length = 0
	sc.arith_total_bits = 0

	sc.output_buf = nil
	sc.arith_output_buf = nil
	sc.output_syms = nil

	sc.pDecode_need_bytes_func = nil
	sc.pDecode_private_data = nil
	// sc.pSaved_huff_model = nil
	// sc.pSaved_model = nil
	sc.saved_node_index = 0
}
