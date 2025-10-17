package models

import (
	"strings"
	"time"
)

// SegmentResult represents transcription output with timestamps
type SegmentResult struct {
	Start time.Duration `json:"start"`
	End   time.Duration `json:"end"`
	Text  string        `json:"text"`
}

// Voice represents a text-to-speech voice option.
type Voice string

// Enum-like constants for voices.
const (
	// DISNEY VOICES
	GHOSTFACE    Voice = "en_us_ghostface"
	CHEWBACCA    Voice = "en_us_chewbacca"
	C3PO         Voice = "en_us_c3po"
	STITCH       Voice = "en_us_stitch"
	STORMTROOPER Voice = "en_us_stormtrooper"
	ROCKET       Voice = "en_us_rocket"
	MADAME_LEOTA Voice = "en_female_madam_leota"
	GHOST_HOST   Voice = "en_male_ghosthost"
	PIRATE       Voice = "en_male_pirate"

	// ENGLISH VOICES
	AU_FEMALE_1           Voice = "en_au_001"
	AU_MALE_1             Voice = "en_au_002"
	UK_MALE_1             Voice = "en_uk_001"
	UK_MALE_2             Voice = "en_uk_003"
	US_FEMALE_1           Voice = "en_us_001"
	US_FEMALE_2           Voice = "en_us_002"
	US_MALE_1             Voice = "en_us_006"
	US_MALE_2             Voice = "en_us_007"
	US_MALE_3             Voice = "en_us_009"
	US_MALE_4             Voice = "en_us_010"
	MALE_JOMBOY           Voice = "en_male_jomboy"
	MALE_CODY             Voice = "en_male_cody"
	FEMALE_SAMC           Voice = "en_female_samc"
	FEMALE_MAKEUP         Voice = "en_female_makeup"
	FEMALE_RICHGIRL       Voice = "en_female_richgirl"
	MALE_GRINCH           Voice = "en_male_grinch"
	MALE_DEADPOOL         Voice = "en_male_deadpool"
	MALE_JARVIS           Voice = "en_male_jarvis"
	MALE_ASHMAGIC         Voice = "en_male_ashmagic"
	MALE_OLANTERKKERS     Voice = "en_male_olantekkers"
	MALE_UKNEIGHBOR       Voice = "en_male_ukneighbor"
	MALE_UKBUTLER         Voice = "en_male_ukbutler"
	FEMALE_SHENNA         Voice = "en_female_shenna"
	FEMALE_PANSINO        Voice = "en_female_pansino"
	MALE_TREVOR           Voice = "en_male_trevor"
	FEMALE_BETTY          Voice = "en_female_betty"
	MALE_CUPID            Voice = "en_male_cupid"
	FEMALE_GRANDMA        Voice = "en_female_grandma"
	MALE_XMXS_CHRISTMAS   Voice = "en_male_m2_xhxs_m03_christmas"
	MALE_SANTA_NARRATION  Voice = "en_male_santa_narration"
	MALE_SING_DEEP_JINGLE Voice = "en_male_sing_deep_jingle"
	MALE_SANTA_EFFECT     Voice = "en_male_santa_effect"
	FEMALE_HT_NEYEAR      Voice = "en_female_ht_f08_newyear"
	MALE_WIZARD           Voice = "en_male_wizard"
	FEMALE_HT_HALLOWEEN   Voice = "en_female_ht_f08_halloween"

	// EUROPE VOICES
	FR_MALE_1 Voice = "fr_001"
	FR_MALE_2 Voice = "fr_002"
	DE_FEMALE Voice = "de_001"
	DE_MALE   Voice = "de_002"
	ES_MALE   Voice = "es_002"

	// AMERICA VOICES
	ES_MX_MALE         Voice = "es_mx_002"
	BR_FEMALE_1        Voice = "br_001"
	BR_FEMALE_2        Voice = "br_003"
	BR_FEMALE_3        Voice = "br_004"
	BR_MALE            Voice = "br_005"
	BP_FEMALE_IVETE    Voice = "bp_female_ivete"
	BP_FEMALE_LUDMILLA Voice = "bp_female_ludmilla"
	PT_FEMALE_LHAYS    Voice = "pt_female_lhays"
	PT_FEMALE_LAIZZA   Voice = "pt_female_laizza"
	PT_MALE_BUENO      Voice = "pt_male_bueno"

	// ASIA VOICES
	ID_FEMALE               Voice = "id_001"
	JP_FEMALE_1             Voice = "jp_001"
	JP_FEMALE_2             Voice = "jp_003"
	JP_FEMALE_3             Voice = "jp_005"
	JP_MALE                 Voice = "jp_006"
	KR_MALE_1               Voice = "kr_002"
	KR_FEMALE               Voice = "kr_003"
	KR_MALE_2               Voice = "kr_004"
	JP_FEMALE_FUJICOCHAN    Voice = "jp_female_fujicochan"
	JP_FEMALE_HASEGAWARIONA Voice = "jp_female_hasegawariona"
	JP_MALE_KEIICHINAKANO   Voice = "jp_male_keiichinakano"
	JP_FEMALE_OOMAEAIIKA    Voice = "jp_female_oomaeaika"
	JP_MALE_YUJINCHIGUSA    Voice = "jp_male_yujinchigusa"
	JP_FEMALE_SHIROU        Voice = "jp_female_shirou"
	JP_MALE_TAMAWAKAZUKI    Voice = "jp_male_tamawakazuki"
	JP_FEMALE_KAORISHOJI    Voice = "jp_female_kaorishoji"
	JP_FEMALE_YAGISHAKI     Voice = "jp_female_yagishaki"
	JP_MALE_HIKAKIN         Voice = "jp_male_hikakin"
	JP_FEMALE_REI           Voice = "jp_female_rei"
	JP_MALE_SHUICHIRO       Voice = "jp_male_shuichiro"
	JP_MALE_MATSUDAKE       Voice = "jp_male_matsudake"
	JP_FEMALE_MACHIKORIIITA Voice = "jp_female_machikoriiita"
	JP_MALE_MATSUO          Voice = "jp_male_matsuo"
	JP_MALE_OSADA           Voice = "jp_male_osada"

	// SINGING VOICES
	SING_FEMALE_ALTO             Voice = "en_female_f08_salut_damour"
	SING_MALE_TENOR              Voice = "en_male_m03_lobby"
	SING_FEMALE_WARMY_BREEZE     Voice = "en_female_f08_warmy_breeze"
	SING_MALE_SUNSHINE_SOON      Voice = "en_male_m03_sunshine_soon"
	SING_FEMALE_GLORIOUS         Voice = "en_female_ht_f08_glorious"
	SING_MALE_IT_GOES_UP         Voice = "en_male_sing_funny_it_goes_up"
	SING_MALE_CHIPMUNK           Voice = "en_male_m2_xhxs_m03_silly"
	SING_FEMALE_WONDERFUL_WORLD  Voice = "en_female_ht_f08_wonderful_world"
	SING_MALE_FUNNY_THANKSGIVING Voice = "en_male_sing_funny_thanksgiving"

	// OTHER
	MALE_NARRATION   Voice = "en_male_narration"
	MALE_FUNNY       Voice = "en_male_funny"
	FEMALE_EMOTIONAL Voice = "en_female_emotional"
)

// FromString returns the Voice constant matching the given name (case-insensitive).
func FromString(input string) (Voice, bool) {
	switch strings.ToUpper(input) {
	case "GHOSTFACE":
		return GHOSTFACE, true
	case "CHEWBACCA":
		return CHEWBACCA, true
	case "C3PO":
		return C3PO, true
	case "STITCH":
		return STITCH, true
	case "STORMTROOPER":
		return STORMTROOPER, true
	case "ROCKET":
		return ROCKET, true
	case "MADAME_LEOTA":
		return MADAME_LEOTA, true
	case "GHOST_HOST":
		return GHOST_HOST, true
	case "PIRATE":
		return PIRATE, true
	// ... ⚡ You would continue adding all cases here (I can auto-generate the full switch if you want)
	default:
		return "", false
	}
}
