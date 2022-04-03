package lzham

const (
	// Update and print high-level coding statistics if set to 1.
	// TODO: Add match distance coding statistics.
	LZHAM_UPDATE_STATS = 0

	// Only parse on the main thread, for easier debugging.
	LZHAM_FORCE_SINGLE_THREADED_PARSING = 0

	// Verify all computed match costs against the generic/slow state::get_cost() method.
	LZHAM_VERIFY_MATCH_COSTS = 0

	// Set to 1 to force all blocks to be uncompressed (raw).
	LZHAM_FORCE_ALL_RAW_BLOCKS = 0

	LZHAM_EXTREME_PARSING_FAST_BYTES = 96
)

const (
	cMaxParseGraphNodes uint32 = 3072
	cMaxParseThreads    uint32 = 8

	cMaxParseNodeStates        = LZHAM_EXTREME_PARSING_MAX_BEST_ARRIVALS_MAX
	cDefaultMaxParseNodeStates = 4
)

type compression_level uint8

const (
	cCompressionLevelFastest compression_level = iota
	cCompressionLevelFaster
	cCompressionLevelDefault
	cCompressionLevelBetter
	cCompressionLevelUber

	cCompressionLevelCount
)

type comp_settings struct {
	fast_bytes                        uint32
	match_accel_max_matches_per_probe uint32
	match_accel_max_probes            uint32
}

const (
	cDefaultBlockSize = 2 << 18
)

type init_params struct {
	max_helper_threads uint32

	compression_level compression_level
	dict_size_log2    uint32

	block_size uint32

	lzham_compress_flags uint32

	pSeed_bytes    []byte
	num_seed_bytes uint32

	table_max_update_interval       uint32
	table_update_interval_slow_rate uint32

	extreme_parsing_max_best_arrivals uint32
	fast_bytes_override               uint32
}

const (
	cLitComplexity  = 1
	cRep0Complexity = 2
	cRep3Complexity = 5

	cLongMatchComplexity          = 6
	cLongMatchComplexityLenThresh = 9

	cShortMatchComplexity = 7
)

type tracked_stat struct {
	num     uint64
	total   float64
	total2  float64
	min_val float64
	max_val float64
}

func (ts *tracked_stat) clear() {
	ts.num = 0
	ts.total = 0
	ts.total2 = 0
	ts.min_val = 0
	ts.max_val = 0
}

type coding_stats struct {
	total_bytes    uint32
	total_contexts uint32
	total_cost     float64

	context_stats tracked_stat

	total_match_bits_cost     float64
	worst_match_bits_cost     float64
	total_is_match0_bits_cost float64
	total_is_match1_bits_cost float64

	total_truncated_matches           uint32
	match_truncation_len_hist         [cMaxMatchLen + 1]uint32
	match_truncation_hist             [cMaxMatchLen + 1]uint32
	match_type_truncation_hist        [cNumStates][5]uint32
	match_type_was_not_truncated_hist [cNumStates][5]uint32

	total_nonmatches uint32
	total_matches    uint32

	lit_stats       tracked_stat
	delta_lit_stats tracked_stat

	rep_stats            [cMatchHistSize]tracked_stat
	rep0_len1_stats      tracked_stat
	rep0_len2_plus_stats tracked_stat

	full_match_stats [cMaxMatchLen + 1]tracked_stat

	total_far_len2_matches  uint32
	total_near_len2_matches uint32

	total_update_rate_resets uint32

	max_len2_dist uint32
}

func (cs *coding_stats) clear() {
	cs.total_bytes = 0
	cs.total_contexts = 0
	cs.total_match_bits_cost = 0
	cs.worst_match_bits_cost = 0
	cs.total_is_match0_bits_cost = 0
	cs.total_is_match1_bits_cost = 0
	cs.context_stats.clear()

	cs.total_nonmatches = 0
	cs.total_matches = 0
	cs.total_cost = 0.0

	cs.lit_stats.clear()
	cs.delta_lit_stats.clear()

	cs.rep0_len1_stats.clear()

	var i uint32
	for i = 0; i < cMatchHistSize; i++ {
		cs.rep_stats[i].clear()
	}
	cs.rep0_len1_stats.clear()
	cs.rep0_len2_plus_stats.clear()

	for i = 0; i <= cMaxMatchLen; i++ {
		cs.full_match_stats[i].clear()
	}

	cs.total_far_len2_matches = 0
	cs.total_near_len2_matches = 0

	cs.total_truncated_matches = 0
	cs.match_truncation_len_hist = [cMaxMatchLen + 1]uint32{}
	cs.match_truncation_hist = [cMaxMatchLen + 1]uint32{}
	cs.match_type_truncation_hist = [cNumStates][5]uint32{}
	cs.match_type_was_not_truncated_hist = [cNumStates][5]uint32{}

	cs.total_update_rate_resets = 0

	cs.max_len2_dist = 0
}

type lzcompressor struct {
	params   init_params
	settings comp_settings

	src_size    int64
	src_adler32 uint32

	accel search_accelerator

	codec symbol_codec

	stats coding_stats

	block_buf []byte
	comp_buf  []byte

	step uint32

	block_start_dict_ofs uint32

	block_index uint32

	finished            bool
	use_task_pool       bool
	use_extreme_parsing bool

	fast_bytes uint32

	num_parse_threads uint32
}

func (lz *lzcompressor) init(params *init_params) bool {
	if params.dict_size_log2 < cMinDictSizeLog2 || params.dict_size_log2 > cMaxDictSizeLog2 {
		return false
	}

	if params.compression_level < 0 || params.compression_level > cCompressionLevelCount {
		return false
	}

	use_extreme_parsing := params.lzham_compress_flags&uint32(LZHAM_COMP_FLAG_EXTREME_PARSING) > 0 && params.compression_level == cCompressionLevelUber

	settings := s_level_settings[params.compression_level]

	if use_extreme_parsing {
		lz.fast_bytes = LZHAM_EXTREME_PARSING_FAST_BYTES
	} else {
		lz.fast_bytes = settings.fast_bytes
	}
	if params.fast_bytes_override > 0 {
		lz.fast_bytes = clamp(params.fast_bytes_override, 8, cMaxMatchLen+1)
	}

	var dict_size uint32 = 1 << params.dict_size_log2

	if params.num_seed_bytes > 0 {
		if params.pSeed_bytes != nil {
			return false
		}
		if params.num_seed_bytes > uint32(dict_size) {
			return false
		}
	}

	var max_block_size uint32 = dict_size / 8
	if params.block_size > max_block_size {
		params.block_size = max_block_size
	}

	var num_parse_threads uint32 = 1

	num_parse_jobs := num_parse_threads - 1
	match_accel_helper_threads := LZHAM_MAX(0, params.max_helper_threads-num_parse_jobs)
	match_accel_helper_threads = LZHAM_MIN(match_accel_helper_threads, cMatchAccelMaxSupportedThreads)

	var accel_flags uint32 = 0
	if params.lzham_compress_flags&uint32(LZHAM_COMP_FLAG_DETERMINISTIC_PARSING) > 0 {
		accel_flags |= cFlagDeterministic
	}

	if params.compression_level > cCompressionLevelFastest {
		if (params.lzham_compress_flags & uint32(LZHAM_COMP_FLAG_USE_LOW_MEMORY_MATCH_FINDER)) == 0 {
			accel_flags |= cFlagHash24
		}
		accel_flags |= cFlagLen2Matches
	}

	if !lz.accel.init(match_accel_helper_threads, dict_size, settings.match_accel_max_matches_per_probe, false, settings.match_accel_max_probes, accel_flags) {
		return false
	}

	lz.block_buf = make([]byte, params.block_size)
	lz.comp_buf = make([]byte, params.block_size*2)

	if params.num_seed_bytes > 0 {
		if !lz.init_seed_bytes() {
			logger.Errorf("init_seed_bytes failed, code %d", 7010)
			return false
		}
	}

	lz.src_size = 0

	return true
}

func (lz *lzcompressor) reset() bool {
	lz.accel.reset()
	lz.codec.reset()
	lz.stats.clear()
	lz.src_size = 0
	lz.src_adler32 = 1
	lz.block_buf = nil
	lz.comp_buf = nil

	lz.step = 0
	lz.finished = false
	lz.block_start_dict_ofs = 0
	lz.block_index = 0

	if lz.params.num_seed_bytes > 0 {
		if !lz.init_seed_bytes() {
			return false
		}
	}

	return lz.send_zlib_header()
}

func (lz *lzcompressor) init_seed_bytes() bool {
	// var cur_seed_ofs uint32 = 0

	// for cur_seed_ofs < lz.params.num_seed_bytes {
	// 	var total_bytes_remaining uint32 = lz.params.num_seed_bytes - cur_seed_ofs
	// 	var num_bytes_to_add uint32 = minimum(total_bytes_remaining, lz.params.block_size)

	// 	if !lz.accel.add_bytes_begin(num_bytes_to_add, (lz.params.pSeed_bytes)+cur_seed_ofs) {
	// 		return false
	// 	}
	// 	lz.accel.add_bytes_end()

	// 	lz.accel.advance_bytes(num_bytes_to_add)

	// 	cur_seed_ofs += num_bytes_to_add
	// }

	return true
}

func (lz *lzcompressor) send_zlib_header() bool {
	if (lz.params.lzham_compress_flags & uint32(LZHAM_COMP_FLAG_WRITE_ZLIB_STREAM)) == 0 {
		return true
	}

	cmf := LZHAM_Z_LZHAM | ((lz.params.dict_size_log2 - 15) << 4)
	var flg uint32 = 0

	switch lz.params.compression_level {
	case compression_level(LZHAM_COMP_LEVEL_FASTEST):
		flg = 0 << 6
	case compression_level(LZHAM_COMP_LEVEL_FASTER):
		flg = 1 << 6
	case compression_level(LZHAM_COMP_LEVEL_DEFAULT), compression_level(LZHAM_COMP_LEVEL_BETTER):
		flg = 2 << 6
	default:
		flg = 3 << 6
	}

	if lz.params.pSeed_bytes != nil {
		flg |= 32
	}

	check := ((cmf << uint32(8)) + flg) % 31
	if check > 0 {
		flg += 31 - check
	}

	// i dont even know anymore

	return true
}
