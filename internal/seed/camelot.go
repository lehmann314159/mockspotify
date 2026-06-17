package seed

// camelotSlots lists all 24 Camelot wheel slots in order.
var camelotSlots = []struct {
	Code   string
	KeyOf  string
	Weight float64
}{
	{"8B", "C", 1.0},
	{"9B", "G", 0.9},
	{"10B", "D", 0.8},
	{"11B", "A", 0.7},
	{"7B", "F", 0.7},
	{"1A", "Am", 0.8},
	{"9A", "Em", 0.7},
	{"7A", "Dm", 0.7},
	{"12B", "E", 0.6},
	{"6B", "Bb", 0.6},
	{"5B", "Eb", 0.5},
	{"10A", "Bm", 0.5},
	{"4B", "Ab", 0.4},
	{"2B", "F#", 0.3},
	{"1B", "B", 0.3},
	// remaining 9 slots at weight 0.2
	{"2A", "Bm", 0.2},
	{"3A", "C#m", 0.2},
	{"4A", "G#m", 0.2},
	{"5A", "Ebm", 0.2},
	{"6A", "Bbm", 0.2},
	{"8A", "Cm", 0.2},
	{"11A", "Abm", 0.2},
	{"12A", "Fm", 0.2},
	{"3B", "Db", 0.2},
}

// camelotByCode maps code → (key, weight) for fast lookup.
var camelotByCode map[string]struct {
	KeyOf  string
	Weight float64
}

func init() {
	camelotByCode = make(map[string]struct {
		KeyOf  string
		Weight float64
	}, len(camelotSlots))
	for _, s := range camelotSlots {
		camelotByCode[s.Code] = struct {
			KeyOf  string
			Weight float64
		}{s.KeyOf, s.Weight}
	}
}
