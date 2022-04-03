package lzham

const (
	cMinMatchLen = 2
	cMaxMatchLen = 257

	cMaxHugeMatchLen = 65536

	cMinDictSizeLog2 = 15
	cMaxDictSizeLog2 = 29

	cMatchHistSize    = 4
	cMaxLen2MatchDist = 2047
)

const (
	cLZXNumSecondaryLengths = 249

	cNumHugeMatchCodes    = 1
	cMaxHugeMatchCodeBits = 16

	cLZXNumSpecialLengths = 2

	cLZXLowestUsableMatchSlot = 1
	cLZXMaxPositionSlots      = 128
)

const (
	cLZXSpecialCodeEndOfBlockCode    = 0
	cLZXSpecialCodePartialStateReset = 1
)

const (
	cLZHAMDebugSyncMarkerValue = 666
	cLZHAMDebugSyncMarkerBits  = 12
)

const (
	cBlockHeaderBits    = 2
	cBlockCheckBits     = 4
	cBlockFlushTypeBits = 2

	cSyncBlock = 0
	cCompBlock = 1
	cRawBlock  = 2
	cEOFBlock  = 3
)

const (
	cNumStates    = 12
	cNumLitStates = 7
)

type table_update_settings struct {
	m_max_update_interval uint16
	m_slow_rate           uint16
}

var (
	g_table_update_settings = []table_update_settings{
		{4, 32}, // crazy slow!
		{5, 33},
		{6, 34},
		{7, 35},
		{8, 36},

		{16, 48},
		{32, 72},
		{64, 64}, // codec's internal default
		{98, 80},
		{128, 96},

		{192, 112},
		{256, 128},
		{512, 128 + 16*2},
		{1024, 128 + 16*4},
		{2048, 128 + 16*6},

		{2048, 128 + 16*8},
		{2048, 128 + 16*10},
		{2048, 128 + 16*12},
		{2048, 128 + 16*14},
		{2048, 128 + 16*16},
	}

	s_level_settings = [cCompressionLevelCount]comp_settings{

		// cCompressionLevelFastest
		{
			8, // fast_bytes
			1, // match_accel_max_matches_per_probe
			2, // match_accel_max_probes
		},
		// cCompressionLevelFaster
		{
			24, // fast_bytes
			6,  // match_accel_max_matches_per_probe
			12, // match_accel_max_probes
		},
		// cCompressionLevelDefault
		{
			32,         // fast_bytes
			^uint32(0), // match_accel_max_matches_per_probe
			16,         // match_accel_max_probes
		},
		// cCompressionLevelBetter
		{
			48,         // fast_bytes
			^uint32(0), // match_accel_max_matches_per_probe
			32,         // match_accel_max_probes
		},
		// cCompressionLevelUber
		{
			64,         // fast_bytes
			^uint32(0), // match_accel_max_matches_per_probe
			128,        // match_accel_max_probes
		},
	}

	lzx_position_base = [cLZXMaxPositionSlots]uint32{
		0x0, 0x1, 0x2, 0x3, 0x4, 0x6, 0x8, 0xC, 0x10, 0x18, 0x20, 0x30, 0x40, 0x60, 0x80, 0xC0,
		0x100, 0x180, 0x200, 0x300, 0x400, 0x600, 0x800, 0xC00, 0x1000, 0x1800, 0x2000, 0x3000, 0x4000, 0x6000, 0x8000, 0xC000,
		0x10000, 0x18000, 0x20000, 0x30000, 0x40000, 0x60000, 0x80000, 0xC0000, 0x100000, 0x180000, 0x200000, 0x300000, 0x400000, 0x600000, 0x800000, 0xC00000,
		0x1000000, 0x1800000, 0x2000000, 0x3000000, 0x4000000, 0x6000000, 0x8000000, 0xA000000, 0xC000000, 0xE000000, 0x10000000, 0x12000000, 0x14000000, 0x16000000, 0x18000000, 0x1A000000,
		0x1C000000, 0x1E000000, 0x20000000, 0x22000000, 0x24000000, 0x26000000, 0x28000000, 0x2A000000, 0x2C000000, 0x2E000000, 0x30000000, 0x32000000, 0x34000000, 0x36000000, 0x38000000, 0x3A000000,
		0x3C000000, 0x3E000000, 0x40000000, 0x42000000, 0x44000000, 0x46000000, 0x48000000, 0x4A000000, 0x4C000000, 0x4E000000, 0x50000000, 0x52000000, 0x54000000, 0x56000000, 0x58000000, 0x5A000000,
		0x5C000000, 0x5E000000, 0x60000000, 0x62000000, 0x64000000, 0x66000000, 0x68000000, 0x6A000000, 0x6C000000, 0x6E000000, 0x70000000, 0x72000000, 0x74000000, 0x76000000, 0x78000000, 0x7A000000,
		0x7C000000, 0x7E000000, 0x80000000, 0x82000000, 0x84000000, 0x86000000, 0x88000000, 0x8A000000, 0x8C000000, 0x8E000000, 0x90000000, 0x92000000, 0x94000000, 0x96000000, 0x98000000, 0x9A000000,
	}

	lzx_position_extra_mask = [cLZXMaxPositionSlots]uint32{
		0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x3, 0x3, 0x7, 0x7, 0xF, 0xF, 0x1F, 0x1F, 0x3F, 0x3F,
		0x7F, 0x7F, 0xFF, 0xFF, 0x1FF, 0x1FF, 0x3FF, 0x3FF, 0x7FF, 0x7FF, 0xFFF, 0xFFF, 0x1FFF, 0x1FFF, 0x3FFF, 0x3FFF,
		0x7FFF, 0x7FFF, 0xFFFF, 0xFFFF, 0x1FFFF, 0x1FFFF, 0x3FFFF, 0x3FFFF, 0x7FFFF, 0x7FFFF, 0xFFFFF, 0xFFFFF, 0x1FFFFF, 0x1FFFFF, 0x3FFFFF, 0x3FFFFF,
		0x7FFFFF, 0x7FFFFF, 0xFFFFFF, 0xFFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF,
		0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF,
		0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF,
		0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF,
		0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF, 0x1FFFFFF,
	}

	lzx_position_extra_bits = [cLZXMaxPositionSlots]byte{
		0x0, 0x0, 0x0, 0x0, 0x1, 0x1, 0x2, 0x2, 0x3, 0x3, 0x4, 0x4, 0x5, 0x5, 0x6, 0x6,
		0x7, 0x7, 0x8, 0x8, 0x9, 0x9, 0xA, 0xA, 0xB, 0xB, 0xC, 0xC, 0xD, 0xD, 0xE, 0xE,
		0xF, 0xF, 0x10, 0x10, 0x11, 0x11, 0x12, 0x12, 0x13, 0x13, 0x14, 0x14, 0x15, 0x15, 0x16, 0x16,
		0x17, 0x17, 0x18, 0x18, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19,
		0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19,
		0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19,
		0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19,
		0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19, 0x19,
	}
)