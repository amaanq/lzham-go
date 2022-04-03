package lzham

// lzbase.cpp

var (
	slot_tab0 = [4096]uint32{}
	slot_tab1 = [512]uint32{}
	slot_tab2 = [256]uint32{}
)

func compute_lzx_position_slot(dist uint32) (uint32, uint32) {
	var s uint32
	if dist < 0x1000 {
		s = slot_tab0[dist]
	} else if dist < 0x100000 {
		s = slot_tab1[dist>>11]
	} else if dist < 0x1000000 {
		s = slot_tab2[dist>>16]
	} else if dist < 0x2000000 {
		s = 48 + ((dist - 0x1000000) >> 23)
	} else if dist < 0x4000000 {
		s = 50 + ((dist - 0x2000000) >> 24)
	} else {
		s = 52 + ((dist - 0x4000000) >> 25)
	}

	ofs := (dist - lzx_position_base[s]) & lzx_position_extra_mask[s]
	slot := s

	return slot, ofs
}
