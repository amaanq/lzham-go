package lzham

const (
	cHashSize24 = 0x1000000
	cHashSize16 = 0x10000
)

const (
	cMatchAccelMaxSupportedProbes  uint32 = 128
	cMatchAccelMaxSupportedThreads uint32 = 32
)

// If all_matches is true, the match finder returns all found matches with no filtering.
// Otherwise, the finder will tend to return lists of matches with mostly unique lengths.
// For each length, it will discard matches with worse distances (in the coding sense).
const (
	cFlagDeterministic = 1 << 0
	cFlagLen2Matches   = 1 << 1
	cFlagHash24        = 1 << 2
)

type node struct {
	left  uint32
	right uint32
}

type dict_match struct {
	dist uint32
	len  uint16
}

func (d dict_match) get_dist() uint32 {
	return d.dist & 0x7FFFFFFF
}

func (d dict_match) get_len() uint16 {
	return d.len + 2
}

func (d dict_match) is_last() bool {
	return d.dist < 0
}

type search_accelerator struct {
	// CLZBase            *m_pLZBase figure out later
	max_helper_threads uint32

	max_dict_size      uint32
	max_dict_size_mask uint32

	lookahead_pos  uint32
	lookahead_size uint32

	cur_dict_size uint32

	dict []byte

	hash  []uint32
	nodes []node

	matches    []dict_match
	match_refs []int32

	digram_hash []uint32
	digram_next []uint32

	thread_dict_offsets [cMatchAccelMaxSupportedThreads]uint32

	fill_lookahead_pos  uint32
	fill_lookahead_size uint32
	fill_dict_size      uint32

	max_probes  uint32
	max_matches uint32

	all_matches bool

	deterministic bool
	len2_matches  bool
	hash24        bool

	next_match_ref int32

	num_completed_helper_threads int32
}

const (
	cDigramHashSize = 4096
)

func (sa *search_accelerator) init(max_helper_threads, max_dict_size, max_matches uint32, all_matches bool, max_probes, flags uint32) bool {
	if !is_power_of_2(uint64(max_dict_size)) {
		return false
	}
	if max_probes == 0 {
		max_probes = 1
	}

	sa.max_probes = LZHAM_MIN(cMatchAccelMaxSupportedProbes, max_probes)
	sa.deterministic = (flags & cFlagDeterministic) != 0
	sa.len2_matches = (flags & cFlagLen2Matches) != 0
	sa.hash24 = (flags & cFlagHash24) != 0
	sa.max_helper_threads = 0
	sa.max_matches = LZHAM_MIN(sa.max_probes, max_matches)
	sa.all_matches = all_matches

	sa.max_dict_size = max_dict_size
	sa.max_dict_size_mask = max_dict_size - 1
	sa.cur_dict_size = 0
	sa.lookahead_size = 0
	sa.lookahead_pos = 0
	sa.fill_lookahead_pos = 0
	sa.fill_lookahead_size = 0
	sa.fill_dict_size = 0
	sa.num_completed_helper_threads = 0

	sa.dict = make([]byte, max_dict_size+LZHAM_MIN(sa.max_dict_size, cMaxHugeMatchLen))

	if sa.hash24 {
		sa.hash = make([]uint32, cHashSize24)
	} else {
		sa.hash = make([]uint32, cHashSize16)
	}

	sa.nodes = make([]node, max_dict_size)

	var i uint32
	for i = 0; i < max_helper_threads; i++ {
		sa.thread_dict_offsets[i] = 256 * 1024
	}

	return true
}

func (sa *search_accelerator) reset() {
	sa.cur_dict_size = 0
	sa.lookahead_size = 0
	sa.lookahead_pos = 0
	sa.fill_lookahead_pos = 0
	sa.fill_lookahead_size = 0
	sa.fill_dict_size = 0
	sa.num_completed_helper_threads = 0

	if len(sa.hash) > 0 {
		sa.hash[0] = 0
		for bp := 1; bp < len(sa.hash); bp *= 2 {
			copy(sa.hash[bp:], sa.hash[:bp])
		}
	}

	if len(sa.digram_hash) > 0 {
		sa.digram_hash[0] = 0
		for bp := 1; bp < len(sa.digram_hash); bp *= 2 {
			copy(sa.digram_hash[bp:], sa.digram_hash[:bp])
		}
	}
}

func (sa *search_accelerator) add_bytes_begin(num_bytes uint32, pBytes []byte) {
	var add_pos uint32 = sa.lookahead_pos & sa.max_dict_size_mask

	n := copy(sa.dict[add_pos:], pBytes[:num_bytes])
	if uint32(n) != num_bytes {
		panic("copy failed")
	}

	var dict_bytes_to_mirror uint32 = LZHAM_MIN(cMaxHugeMatchLen, sa.max_dict_size)
	if add_pos < dict_bytes_to_mirror {
		copy(sa.dict[sa.max_dict_size:], sa.dict[0:dict_bytes_to_mirror])
	}

	sa.lookahead_size = num_bytes

	var max_possible_dict_size uint32 = sa.max_dict_size - num_bytes
	sa.cur_dict_size = LZHAM_MIN(sa.cur_dict_size, max_possible_dict_size)

	sa.next_match_ref = 0

	//return sa.find_all_matches(num_bytes)
}
