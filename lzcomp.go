package lzham

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidDictSizeLog2  = errors.New("invalid dict_size_log2")
	ErrCompressorInitFailed = errors.New("failed to initialize compressor")
	ErrNilCompressState     = errors.New("nil compress state")
)

type LZHAM_compress_state struct {
	compressor lzcompressor

	dict_size_log2 uint32

	pIn_buf       []byte
	pIn_buf_size  uint64
	pOut_buf      []byte
	pOut_buf_size uint64

	comp_data_ofs uint64

	finished_compression bool

	params LZHAM_compress_params

	status lzham_compress_status_t
}

type LZHAM_compress_params struct {
	// m_struct_size        uint32               // set to sizeof(lzham_compress_params) this isn't really needed..
	dict_size_log2     uint32               // set to the log2(dictionary_size), must range between [LZHAM_MIN_DICT_SIZE_LOG2, LZHAM_MAX_DICT_SIZE_LOG2_X86] for x86 LZHAM_MAX_DICT_SIZE_LOG2_X64 for x64
	level              lzham_compress_level // set to LZHAM_COMP_LEVEL_FASTEST, etc.
	table_update_rate  uint32               // Controls tradeoff between ratio and decompression throughput. 0=default, or [1,LZHAM_MAX_TABLE_UPDATE_RATE], higher=faster but lower ratio.
	max_helper_threads int32                // max # of additional "helper" threads to create, must range between [-1,LZHAMAX_HELPER_THREADS], where -1=max practical
	compress_flags     uint32               // optional compression flags (see lzhacompress_flags enum)
	num_seed_bytes     uint32               // for delta compression (optional) - number of seed bytes pointed to by pSeed_bytes
	pSeed_bytes        []byte               // for delta compression (optional) - pointer to seed bytes buffer, must be at least num_seed_bytes long

	// Advanced settings - set to 0 if you don't care.
	// table_max_update_interval/table_update_interval_slow_rate override table_update_rate and allow finer control over the table update settings.
	// If either are non-zero they will override whatever table_update_rate is set to. Just leave them 0 unless you are specifically customizing them for your data.

	// def=0, typical range 12-128 (LZHAM_DEFAULT_TABLE_UPDATE_RATE=64), controls the max interval between table updates, higher=longer max interval (faster decode/lower ratio). Was 16 in prev. releases.
	table_max_update_interval uint32
	// def=0, 32 or higher (LZHAM_DEFAULT_TABLE_UPDATE_RATE=64), scaled by 32, controls the slowing of the update update freq, higher=more rapid slowing (faster decode/lower ratio). Was 40 in prev. releases.
	table_update_interval_slow_rate uint32

	// If non-zero, must range between LZHAM_EXTREME_PARSING_MAX_BEST_ARRIVALS_MIN and LZHAM_EXTREME_PARSING_MAX_BEST_ARRIVALS_MAX.
	// Field added in version 0x1011
	extreme_parsing_max_best_arrivals uint32

	// If non-zero, must range between LZHAM_MIN_FAST_BYTES-LZHAM_MAX_FAST_BYTES.
	// If this is 0, the compressor will either use a fast_bytes setting controlled by level, or a for extreme parsing a fixed setting of LZHAM_EXTREME_PARSING_FAST_BYTES (96).
	// Field added in version 0x1011
	fast_bytes uint32
}

func LZHAM_lib_compress_init(pParams *LZHAM_compress_params) (*LZHAM_compress_state, error) {
	if pParams.dict_size_log2 < LZHAM_MIN_DICT_SIZE_LOG2 || pParams.dict_size_log2 > LZHAM_MAX_DICT_SIZE_LOG2_X64 {
		return nil, ErrInvalidDictSizeLog2
	}

	var internal_params init_params
	status := create_internal_init_params(&internal_params, pParams)
	if status != LZHAM_COMP_STATUS_SUCCESS {
		return nil, fmt.Errorf("create_internal_init_params failed with status %d", status)
	}

	pState := &LZHAM_compress_state{
		params: *pParams,

		pIn_buf:              nil,
		pIn_buf_size:         0,
		pOut_buf:             nil,
		pOut_buf_size:        0,
		status:               LZHAM_COMP_STATUS_NOT_FINISHED,
		comp_data_ofs:        0,
		finished_compression: false,
	}

	if internal_params.max_helper_threads > 0 {
		internal_params.max_helper_threads = 0
	}

	if !pState.compressor.init(&internal_params) {
		return nil, ErrCompressorInitFailed
	}

	return pState, nil
}

func LZHAM_lib_compress_reinit(pParams *LZHAM_compress_params, ptr *LZHAM_compress_state) (*LZHAM_compress_state, error) {
	if ptr == nil {
		return nil, ErrNilCompressState
	}

	ptr.compressor.reset() // reset the compressor

	ptr.pIn_buf = nil
	ptr.pIn_buf_size = 0
	ptr.pOut_buf = nil
	ptr.pOut_buf_size = 0
	ptr.status = LZHAM_COMP_STATUS_NOT_FINISHED
	ptr.comp_data_ofs = 0
	ptr.finished_compression = false

	return ptr, nil
}

func create_internal_init_params(internal_params *init_params, pParams *LZHAM_compress_params) lzham_compress_status_t {
	if pParams.dict_size_log2 < LZHAM_MIN_DICT_SIZE_LOG2 || pParams.dict_size_log2 > LZHAM_MAX_DICT_SIZE_LOG2_X64 {
		return LZHAM_COMP_STATUS_INVALID_PARAMETER
	}

	if pParams.extreme_parsing_max_best_arrivals > cMaxParseNodeStates {
		return LZHAM_COMP_STATUS_INVALID_PARAMETER
	}

	if pParams.extreme_parsing_max_best_arrivals <= 1 {
		internal_params.extreme_parsing_max_best_arrivals = cDefaultMaxParseNodeStates
	} else {
		internal_params.extreme_parsing_max_best_arrivals = pParams.extreme_parsing_max_best_arrivals
	}

	if pParams.fast_bytes > 0 {
		internal_params.fast_bytes_override = clamp(pParams.fast_bytes, LZHAM_MIN_FAST_BYTES, LZHAM_MAX_FAST_BYTES)
	}

	internal_params.dict_size_log2 = pParams.dict_size_log2

	if pParams.max_helper_threads < 0 {
		internal_params.max_helper_threads = 0
	} else {
		internal_params.max_helper_threads = uint32(pParams.max_helper_threads)
	}
	internal_params.max_helper_threads = LZHAM_MIN(LZHAM_MAX_HELPER_THREADS, internal_params.max_helper_threads)

	internal_params.lzham_compress_flags = pParams.compress_flags

	if pParams.num_seed_bytes > 0 {
		if pParams.pSeed_bytes == nil || (pParams.num_seed_bytes > (1 << pParams.dict_size_log2)) {
			return LZHAM_COMP_STATUS_INVALID_PARAMETER
		}
		internal_params.num_seed_bytes = pParams.num_seed_bytes
		internal_params.pSeed_bytes = pParams.pSeed_bytes
	}

	switch pParams.level {
	case LZHAM_COMP_LEVEL_FASTEST:
		internal_params.compression_level = cCompressionLevelFastest
	case LZHAM_COMP_LEVEL_FASTER:
		internal_params.compression_level = cCompressionLevelFaster
	case LZHAM_COMP_LEVEL_DEFAULT:
		internal_params.compression_level = cCompressionLevelDefault
	case LZHAM_COMP_LEVEL_BETTER:
		internal_params.compression_level = cCompressionLevelBetter
	case LZHAM_COMP_LEVEL_UBER:
		internal_params.compression_level = cCompressionLevelUber
	default:
		return LZHAM_COMP_STATUS_INVALID_PARAMETER
	}

	if pParams.table_max_update_interval > 0 || pParams.table_update_interval_slow_rate > 0 {
		internal_params.table_max_update_interval = pParams.table_max_update_interval
		internal_params.table_update_interval_slow_rate = pParams.table_update_interval_slow_rate
	} else {
		var rate uint32 = pParams.table_update_rate
		if rate == 0 {
			rate = uint32(LZHAM_DEFAULT_TABLE_UPDATE_RATE)
		}
		rate = clamp(rate, 1, uint32(LZHAM_FASTEST_TABLE_UPDATE_RATE)) - 1
		internal_params.table_max_update_interval = uint32(g_table_update_settings[rate].m_max_update_interval)
		internal_params.table_update_interval_slow_rate = uint32(g_table_update_settings[rate].m_slow_rate)
	}

	return LZHAM_COMP_STATUS_SUCCESS
}
