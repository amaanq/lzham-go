package lzham

type lzham_flush_t int8

const (
	LZHAM_NO_FLUSH   lzham_flush_t = iota
	LZHAM_SYNC_FLUSH lzham_flush_t = iota + 1
	LZHAM_FULL_FLUSH
	LZHAM_FINISH
	LZHAM_TABLE_FLUSH lzham_flush_t = 10
)

const (
	LZHAM_MIN_DICT_SIZE_LOG2     = 15
	LZHAM_MAX_DICT_SIZE_LOG2_X86 = 26
	LZHAM_MAX_DICT_SIZE_LOG2_X64 = 29
)

const (
	LZHAM_MAX_HELPER_THREADS = 64
)

type lzham_compress_status_t uint32

const (
	LZHAM_COMP_STATUS_NOT_FINISHED lzham_compress_status_t = iota
	LZHAM_COMP_STATUS_NEEDS_MORE_INPUT
	LZHAM_COMP_STATUS_HAS_MORE_OUTPUT

	// All the following enums must indicate failure/success.

	LZHAM_COMP_STATUS_FIRST_SUCCESS_OR_FAILURE_CODE
	LZHAM_COMP_STATUS_SUCCESS = LZHAM_COMP_STATUS_FIRST_SUCCESS_OR_FAILURE_CODE

	LZHAM_COMP_STATUS_FIRST_FAILURE_CODE  lzham_compress_status_t = iota - 1
	LZHAM_COMP_STATUS_FAILED                                      = LZHAM_COMP_STATUS_FIRST_FAILURE_CODE
	LZHAM_COMP_STATUS_FAILED_INITIALIZING                         = iota - 2
	LZHAM_COMP_STATUS_INVALID_PARAMETER
	LZHAM_COMP_STATUS_OUTPUT_BUF_TOO_SMALL
	LZHAM_COMP_STATUS_FORCE_DWORD lzham_compress_status_t = 0xFFFFFFFF
)

type lzham_compress_level uint32

const (
	LZHAM_COMP_LEVEL_FASTEST lzham_compress_level = iota
	LZHAM_COMP_LEVEL_FASTER
	LZHAM_COMP_LEVEL_DEFAULT
	LZHAM_COMP_LEVEL_BETTER
	LZHAM_COMP_LEVEL_UBER

	LZHAM_TOTAL_COMP_LEVELS

	LZHAM_COMP_LEVEL_FORCE_DWORD lzham_compress_level = 0xFFFFFFFF
)

type lzham_compress_flags uint8

const (
	LZHAM_COMP_FLAG_EXTREME_PARSING       lzham_compress_flags = 1 << (iota + 1) // Improves ratio by allowing the compressor's parse graph to grow "higher" (up to 4 parent nodes per output node), but is much slower.
	LZHAM_COMP_FLAG_DETERMINISTIC_PARSING                                        // Guarantees that the compressed output will always be the same given the same input and parameters (no variation between runs due to kernel threading scheduling).

	_
	// If enabled, the compressor is free to use any optimizations which could lower the decompression rate (such
	// as adaptively resetting the Huffman table update rate to maximum frequency, which is costly for the decompressor).
	LZHAM_COMP_FLAG_TRADEOFF_DECOMPRESSION_RATE_FOR_COMP_RATIO

	LZHAM_COMP_FLAG_WRITE_ZLIB_STREAM

	LZHAM_COMP_FLAG_FORCE_SINGLE_THREADED_PARSING

	LZHAM_COMP_FLAG_USE_LOW_MEMORY_MATCH_FINDER
)

type lzham_table_update_rate uint8

const (
	LZHAM_INSANELY_SLOW_TABLE_UPDATE_RATE lzham_table_update_rate = 1
	LZHAM_SLOW_TABLE_UPDATE_RATE          lzham_table_update_rate = 2
	LZHAM_DEFAULT_TABLE_UPDATE_RATE       lzham_table_update_rate = 8
	LZHAM_FASTEST_TABLE_UPDATE_RATE       lzham_table_update_rate = 20
)

const (
	LZHAM_EXTREME_PARSING_MAX_BEST_ARRIVALS_MIN = 2
	LZHAM_EXTREME_PARSING_MAX_BEST_ARRIVALS_MAX = 8
)

const (
	LZHAM_MIN_FAST_BYTES = 8
	LZHAM_MAX_FAST_BYTES = 258
)

type lzham_decompress_status_t uint32

const (
	// LZHAM_DECOMP_STATUS_NOT_FINISHED indicates that the decompressor is flushing its internal buffer to the caller's output buffer.
	// There may be more bytes available to decompress on the next call, but there is no guarantee.
	LZHAM_DECOMP_STATUS_NOT_FINISHED lzham_decompress_status_t = iota

	// LZHAM_DECOMP_STATUS_HAS_MORE_OUTPUT indicates that the decompressor is trying to flush its internal buffer to the caller's output buffer,
	// but the caller hasn't provided any space to copy this data to the caller's output buffer. Call the lzham_decompress() again with a non-empty sized output buffer.
	LZHAM_DECOMP_STATUS_HAS_MORE_OUTPUT

	// LZHAM_DECOMP_STATUS_NEEDS_MORE_INPUT indicates that the decompressor has consumed all input bytes, has not encountered an "end of stream" code,
	// and the caller hasn't set no_more_input_bytes_flag to true, so it's expecting more input to proceed.
	LZHAM_DECOMP_STATUS_NEEDS_MORE_INPUT

	// All the following enums always (and MUST) indicate failure/success.
	LZHAM_DECOMP_STATUS_FIRST_SUCCESS_OR_FAILURE_CODE

	// LZHAM_DECOMP_STATUS_SUCCESS indicates decompression has successfully completed.
	LZHAM_DECOMP_STATUS_SUCCESS = LZHAM_DECOMP_STATUS_FIRST_SUCCESS_OR_FAILURE_CODE

	// The remaining status codes indicate a failure of some sort. Most failures are unrecoverable. TODO: Document which codes are recoverable.
	LZHAM_DECOMP_STATUS_FIRST_FAILURE_CODE lzham_decompress_status_t = iota - 1

	LZHAM_DECOMP_STATUS_FAILED_INITIALIZING                                 = LZHAM_DECOMP_STATUS_FIRST_FAILURE_CODE
	LZHAM_DECOMP_STATUS_FAILED_DEST_BUF_TOO_SMALL lzham_decompress_status_t = iota - 2
	LZHAM_DECOMP_STATUS_FAILED_EXPECTED_MORE_RAW_BYTES
	LZHAM_DECOMP_STATUS_FAILED_BAD_CODE
	LZHAM_DECOMP_STATUS_FAILED_ADLER32
	LZHAM_DECOMP_STATUS_FAILED_BAD_RAW_BLOCK
	LZHAM_DECOMP_STATUS_FAILED_BAD_COMP_BLOCK_SYNC_CHECK
	LZHAM_DECOMP_STATUS_FAILED_BAD_ZLIB_HEADER
	LZHAM_DECOMP_STATUS_FAILED_NEED_SEED_BYTES
	LZHAM_DECOMP_STATUS_FAILED_BAD_SEED_BYTES
	LZHAM_DECOMP_STATUS_FAILED_BAD_SYNC_BLOCK
	LZHAM_DECOMP_STATUS_INVALID_PARAMETER
)

type lzham_decompress_flags uint8

const (
	LZHAM_DECOMP_FLAG_OUTPUT_UNBUFFERED lzham_decompress_flags = 1 << iota
	LZHAM_DECOMP_FLAG_COMPUTE_ADLER32
	LZHAM_DECOMP_FLAG_READ_ZLIB_STREAM
)

