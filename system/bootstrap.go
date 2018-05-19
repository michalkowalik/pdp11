package system

import (
	"pdp/pdpcpu"
)

/*
	Minimal bootstrap code -> load to memory, start executing.
*/

const (
	// BOOTBASE is a base bootstrap address
	BOOTBASE = 0140000
)

var bootcode = [...]uint16{
	0000005, 0005037, 0177776, 0005037, 0177772, 0005037, 0177564, 0005037,
	0177546, 0005067, 0000122, 0010700, 0062700, 0177750, 0010006, 0010700,
	0062700, 0001227, 0004767, 0000144, 0010700, 0062700, 0000104, 0010037,
	0000100, 0012737, 0000340, 0000102, 0052737, 0000100, 0177546, 0010700,
	0062700, 0001241, 0004767, 0000104, 0162706, 0000256, 0010600, 0004767,
	0000172, 0005000, 0000001, 0005200, 0005767, 0000014, 0001773, 0012746,
	0054000, 0016746, 0000002, 0000006, 0000000, 0000000, 0000000, 0005267,
	0177770, 0005367, 0177766, 0016737, 0177762, 0177570, 0000006, 0105737,
	0177564, 0100375, 0110037, 0177566, 0000207, 0000000, 0132737, 0000100,
	0177564, 0001374, 0010067, 0177762, 0010700, 0062700, 0000026, 0010037,
	0000064, 0012737, 0000200, 0000066, 0152737, 0000100, 0177564, 0000207,
	0105777, 0177726, 0001406, 0117737, 0177720, 0177566, 0005267, 0177712,
	0000006, 0105037, 0177564, 0000006, 0000000, 0000000, 0010067, 0177770,
	0005067, 0177766, 0010700, 0062700, 0000026, 0010037, 0000060, 0012737,
	0000200, 0000062, 0152737, 0000100, 0177560, 0000207, 0010046, 0113700,
	0177562, 0120027, 0000015, 0001456, 0120027, 0000177, 0001403, 0120027,
	0000010, 0001013, 0005767, 0177702, 0001407, 0005367, 0177674, 0010700,
	0062700, 0000534, 0004767, 0177564, 0000455, 0120027, 0000040, 0002452,
	0120027, 0000177, 0002047, 0004767, 0177524, 0120027, 0000172, 0003005,
	0120027, 0000141, 0002402, 0142700, 0000040, 0026727, 0177614, 0000254,
	0002031, 0110046, 0016700, 0177600, 0066700, 0177576, 0112610, 0005267,
	0177570, 0000420, 0016700, 0177560, 0066700, 0177556, 0105010, 0010700,
	0062700, 0000030, 0010037, 0000240, 0012737, 0000100, 0000242, 0012737,
	0002000, 0177772, 0012600, 0000006, 0005037, 0177772, 0010046, 0010146,
	0010246, 0010346, 0010446, 0010700, 0062700, 0000341, 0004767, 0177374,
	0132737, 0000100, 0177564, 0001374, 0005004, 0012703, 0141160, 0112300,
	0001435, 0016702, 0177442, 0112201, 0001436, 0120127, 0000040, 0001773,
	0120001, 0001010, 0112300, 0001412, 0112201, 0001410, 0120127, 0000040,
	0001367, 0000404, 0112300, 0001376, 0005204, 0000751, 0006304, 0010700,
	0062700, 0000060, 0060004, 0061404, 0004714, 0000405, 0010700, 0062700,
	0000307, 0004767, 0177246, 0005067, 0177340, 0010700, 0062700, 0000365,
	0004767, 0177230, 0012604, 0012603, 0012602, 0012601, 0012600, 0000006,
	0002216, 0000444, 0001454, 0000454, 0000556, 0000660, 0000464, 0000474,
	0067503, 0066555, 0067141, 0071544, 0060440, 0062562, 0041040, 0067557,
	0026164, 0044040, 0066141, 0026164, 0052040, 0071545, 0026164, 0042040,
	0060551, 0026147, 0046040, 0063551, 0072150, 0026163, 0060440, 0062156,
	0044040, 0066145, 0006560, 0041012, 0067557, 0020164, 0062544, 0064566,
	0062543, 0020163, 0071141, 0020145, 0045522, 0051040, 0020114, 0050122,
	0067440, 0020162, 0046524, 0005015, 0006400, 0000012, 0020010, 0000010,
	0047502, 0052117, 0044000, 0046105, 0000120, 0044514, 0044107, 0051524,
	0044000, 0046101, 0000124, 0042524, 0052123, 0042000, 0040511, 0000107,
	0044103, 0051501, 0051105, 0051000, 0046105, 0041517, 0052101, 0000105,
	0042117, 0000124, 0052400, 0065556, 0067556, 0067167, 0061440, 0066557,
	0060555, 0062156, 0006412, 0050000, 0072541, 0020154, 0060516, 0065556,
	0071145, 0064566, 0020163, 0020055, 0060560, 0066165, 0060556, 0065556,
	0064100, 0072157, 0060555, 0066151, 0061456, 0066557, 0006412, 0006412,
	0041000, 0047517, 0037124, 0000040, 0067125, 0067153, 0073557, 0020156,
	0067542, 0072157, 0062040, 0073145, 0061551, 0005145, 0000015, 0020040,
	0020040, 0020040, 0066143, 0061557, 0020153, 0064564, 0065543, 0005163,
	0000015, 0020040, 0020040, 0020040, 0072151, 0071145, 0072141, 0067551,
	0071556, 0006412, 0000000, 0010700, 0062700, 0177350, 0004767, 0176534,
	0000207, 0000000, 0010700, 0062700, 0177463, 0004767, 0176516, 0000207,
	0012700, 0000001, 0006100, 0000005, 0000775, 0000237, 0013703, 0177570,
	0042703, 0000001, 0020327, 0004000, 0103402, 0012703, 0004000, 0010302,
	0010700, 0062700, 0176236, 0010001, 0010700, 0062700, 0002250, 0012122,
	0020100, 0103775, 0000113, 0005067, 0176360, 0012705, 0015000, 0010504,
	0006204, 0005000, 0010501, 0071004, 0010002, 0070204, 0060301, 0020105,
	0001401, 0000000, 0077412, 0077515, 0010700, 0062700, 0177551, 0016703,
	0176310, 0005002, 0071227, 0000010, 0062703, 0000060, 0110340, 0010203,
	0001370, 0004767, 0176326, 0000207, 0000000, 0005067, 0177772, 0005067,
	0176250, 0005005, 0100404, 0102403, 0101002, 0002401, 0101401, 0000000,
	0005305, 0100003, 0001402, 0002001, 0003401, 0000000, 0006005, 0102002,
	0103001, 0001001, 0000000, 0112737, 0000017, 0177776, 0100004, 0102003,
	0002402, 0101001, 0100401, 0000000, 0012703, 0077777, 0062703, 0077777,
	0123727, 0177776, 0000005, 0001001, 0000000, 0112704, 0001700, 0100003,
	0020427, 0177700, 0001401, 0000000, 0105004, 0020427, 0177400, 0001401,
	0000000, 0012705, 0125252, 0010500, 0010001, 0010102, 0010203, 0010304,
	0010405, 0160501, 0002401, 0001401, 0000000, 0006102, 0103001, 0002401,
	0000000, 0060203, 0005203, 0005103, 0060301, 0103401, 0003401, 0000000,
	0006004, 0050403, 0060503, 0005203, 0103402, 0005301, 0002401, 0000000,
	0005100, 0101401, 0000000, 0040001, 0060101, 0003001, 0003401, 0000000,
	0000301, 0020127, 0052125, 0001004, 0030405, 0003002, 0005105, 0001001,
	0000000, 0112700, 0177401, 0100001, 0000000, 0077002, 0005001, 0005201,
	0077002, 0005700, 0001002, 0005701, 0001401, 0000000, 0010700, 0062700,
	0000010, 0010007, 0000000, 0005727, 0000000, 0010700, 0062700, 0000012,
	0010046, 0000207, 0000000, 0010700, 0062700, 0000014, 0005046, 0010046,
	0000002, 0000000, 0000167, 0000002, 0000000, 0012705, 0000003, 0010700,
	0062700, 0000020, 0005037, 0000006, 0010037, 0000004, 0005745, 0000000,
	0062706, 0000004, 0005000, 0012701, 0007777, 0010002, 0062702, 0000005,
	0022062, 0177773, 0001402, 0000000, 0000000, 0077111, 0005267, 0177270,
	0103001, 0000000, 0026727, 0175542, 0002200, 0101002, 0000167, 0177262,
	0010700, 0062700, 0177005, 0016703, 0177236, 0005002, 0071227, 0000010,
	0062703, 0000060, 0110340, 0010203, 0001370, 0004767, 0175536, 0000207,
	0005767, 0175464, 0001403, 0005067, 0175456, 0000207, 0000237, 0005037,
	0177572, 0012704, 0077406, 0012703, 0172300, 0005001, 0005002, 0004767,
	0000200, 0012737, 0177600, 0172376, 0012703, 0177600, 0005001, 0012702,
	0002000, 0004767, 0000154, 0012737, 0177600, 0177676, 0012700, 0000020,
	0012703, 0172200, 0010423, 0077002, 0012700, 0141400, 0012701, 0000067,
	0032737, 0000010, 0177570, 0001404, 0012700, 0006000, 0012701, 0000007,
	0010037, 0172244, 0010037, 0172264, 0012737, 0177600, 0172276, 0012737,
	0010000, 0177776, 0010137, 0172516, 0012737, 0000001, 0177572, 0012703,
	0040200, 0010700, 0062700, 0000074, 0010002, 0010700, 0062700, 0000322,
	0012246, 0006623, 0020200, 0103774, 0012767, 0040200, 0175232, 0000230,
	0000207, 0012700, 0000010, 0010263, 0000060, 0010163, 0000040, 0010463,
	0000020, 0010423, 0062701, 0000200, 0062702, 0000200, 0077014, 0000207,
	0012700, 0000037, 0012701, 0174000, 0032737, 0000001, 0177570, 0001414,
	0012700, 0007417, 0010001, 0005101, 0032737, 0000002, 0177570, 0001404,
	0012700, 0036163, 0012701, 0037000, 0010102, 0162702, 0000002, 0042702,
	0000001, 0010203, 0042703, 0177701, 0012763, 0000001, 0040000, 0012763,
	0000113, 0040002, 0010204, 0072427, 0177772, 0042704, 0177600, 0013703,
	0172244, 0160403, 0010204, 0072427, 0177764, 0042704, 0177761, 0010364,
	0172240, 0010105, 0072527, 0177764, 0042705, 0177761, 0020504, 0001402,
	0010365, 0172240, 0012704, 0000003, 0010703, 0062703, 0000006, 0000112,
	0077402, 0032737, 0000004, 0177570, 0001005, 0010002, 0006002, 0006101,
	0006000, 0000712, 0010002, 0006102, 0006001, 0006100, 0000705, 0005003,
	0112201, 0001560, 0120127, 0000040, 0001773, 0120127, 0000122, 0001415,
	0120127, 0000124, 0001004, 0112201, 0120127, 0000115, 0001407, 0010700,
	0062700, 0176070, 0004767, 0174724, 0000207, 0112201, 0112200, 0001416,
	0120027, 0000040, 0001413, 0120027, 0000067, 0003033, 0162700, 0000060,
	0002430, 0006303, 0006303, 0006303, 0050003, 0000760, 0010700, 0062700,
	0174442, 0010037, 0000004, 0005037, 0000006, 0120127, 0000113, 0001502,
	0120127, 0000114, 0001414, 0120127, 0000120, 0001523, 0120127, 0000115,
	0001555, 0010700, 0062700, 0175641, 0004767, 0174600, 0000207, 0000005,
	0000303, 0012701, 0174400, 0012761, 0000013, 0000004, 0052703, 0000004,
	0010311, 0105711, 0100376, 0105003, 0052703, 0000010, 0010311, 0105711,
	0100376, 0016102, 0000006, 0042702, 0000077, 0005202, 0010261, 0000004,
	0105003, 0052703, 0000006, 0010311, 0105711, 0100376, 0005061, 0000002,
	0005061, 0000004, 0012761, 0177000, 0000006, 0105003, 0052703, 0000014,
	0010311, 0105711, 0100376, 0042711, 0000377, 0005002, 0005003, 0005004,
	0005005, 0005007, 0000005, 0000303, 0006303, 0006303, 0006303, 0006303,
	0006303, 0012701, 0177412, 0010311, 0005041, 0012741, 0177000, 0012741,
	0000005, 0005002, 0005003, 0005004, 0005005, 0105711, 0100376, 0105011,
	0005007, 0000005, 0012701, 0176700, 0012761, 0000040, 0000010, 0010361,
	0000010, 0012711, 0000021, 0012761, 0010000, 0000032, 0012761, 0177000,
	0000002, 0005061, 0000004, 0005061, 0000006, 0005061, 0000034, 0012711,
	0000071, 0105711, 0100376, 0105011, 0010300, 0005007, 0000005, 0010300,
	0012701, 0172526, 0005011, 0012741, 0177777, 0010002, 0000302, 0062702,
	0060011, 0010241, 0105711, 0100376, 0010002, 0000302, 0062702, 0060003,
	0010211, 0105711, 0100376, 0005002, 0005003, 0012704, 0143744, 0005005,
	0005007}

// Boot loads bootstrap code and start emulation
func (sys *System) Boot() {
	memPointer := uint16(BOOTBASE)

	for _, c := range bootcode {
		mmunit.WriteMemoryWord(memPointer, c)
		memPointer += 2
	}

	// set SP and PC to their starting address:
	sys.CPU.Registers[7] = BOOTBASE
	sys.CPU.Registers[6] = BOOTBASE

	// start execution
	sys.console.WriteConsole("Booting..\n")
	if sys.CPU.State != pdpcpu.RUN {
		sys.CPU.State = pdpcpu.RUN
	}
	// sys.emulate()
	sys.loop()
}
