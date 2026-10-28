package flavor

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strconv"
	"strings"

	"github.com/espresso20/ageforge/pkg/wordlist"
)

// text.go reads the authored sentences and the noun banks from the text files
// in text/. Each file opens with a comment that says how it is written, so
// whoever edits one does not have to open this code.
//
// A file is a series of lists (see pkg/wordlist). A "== words name" list is a
// noun bank, one entry per line. Any other "== name" list is a pool of
// sentences, one per line as "tags | text", held to the eras or ages its
// heading names.

//go:embed text/*.txt
var textFS embed.FS

// textFiles are the files of text/ in the order the copy export lists their
// rows: one file per Moment, named by its pool prefix, then the noun banks the
// harbinger answers share.
var textFiles = []string{
	"enc_cap", "enc_standoff", "exp_fail", "exp_success",
	"harb_appeased", "harb_arrival", "harb_braced", "harb_invited", "harb_fulfilled",
	"harb_vindicated", "harb_spared", "harb_discredited", "harb_warn",
	"run_end", "war_raid",
	"harb_banks",
}

const (
	wordsHeading = "words"
	tagSeparator = " | "
)

var (
	registerNames = map[string]register{"plain": rPlain, "wry": rWry, "joke": rJoke}
	formNames     = map[string]form{"ledger": fLedger, "complaint": fComplaint, "overheard": fOverheard, "notice": fNotice}
	toneNames     = map[string]Tone{"grim": Grim, "triumphant": Triumphant, "wry": Wry}
	needNames     = map[string]need{"subject": needSubject, "res": needRes, "mass": needMassRes, "amount": needAmount}
)

// sentenceLine is one sentence of a text file and where it is written.
type sentenceLine struct {
	skel
	line     int
	from, to int // byte offsets of the sentence text in the file
}

// poolText is a pool of sentences: the sentences that share an era or age
// gating.
type poolText struct {
	name      string // what the ids of its sentences start with: "exp_fail_any"
	bucket    string // the pool's part of its sentences' row ids
	line      int
	eras      []era    // the eras it is held to; nil for any
	ages      []string // the ages it is held to; nil for any age in eras
	sentences []sentenceLine
}

// wordsText is a noun bank.
type wordsText struct {
	name  string
	line  int
	words []wordlist.Entry
}

// fileText is one parsed text file.
type fileText struct {
	name  string // "exp_fail"
	pools []poolText
	banks []wordsText
}

// catalogText is every text file, in textFiles order.
type catalogText struct {
	files []fileText
}

// embedded is the catalog this build was made with. A file that does not
// parse stops the program at start: TestEmbeddedTextParses keeps that from
// reaching a build.
var embedded = mustReadEmbedded()

// banks holds every noun bank by name. Bank entries are lowercase mid-sentence
// noun phrases with no terminal punctuation: they drop into a slot and never
// open a sentence.
var banks = embedded.banks()

func mustReadEmbedded() *catalogText {
	folder, err := fs.Sub(textFS, "text")
	if err != nil {
		panic(err)
	}
	catalog, err := readCatalogText(folder)
	if err != nil {
		panic("flavor: " + err.Error())
	}
	return catalog
}

// readCatalogText parses the files of a text folder and checks them against
// each other: every Moment has its file, every pool and bank name is used
// once, and every slot names a bank.
func readCatalogText(folder fs.FS) (*catalogText, error) {
	if err := checkTextFileSet(folder); err != nil {
		return nil, err
	}
	catalog := &catalogText{}
	for _, name := range textFiles {
		data, err := fs.ReadFile(folder, name+".txt")
		if err != nil {
			return nil, err
		}
		file, err := parseTextFile(name, string(data))
		if err != nil {
			return nil, err
		}
		catalog.files = append(catalog.files, file)
	}
	return catalog, catalog.check()
}

// checkTextFileSet fails when the folder holds a .txt file that textFiles does
// not list: nothing would ever read it.
func checkTextFileSet(folder fs.FS) error {
	entries, err := fs.ReadDir(folder, ".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name, isText := strings.CutSuffix(entry.Name(), ".txt")
		if isText && !slices.Contains(textFiles, name) {
			return fmt.Errorf("%s is not one of the catalog files (textFiles in text.go)", entry.Name())
		}
	}
	return nil
}

// parseTextFile reads one file. The errors name the file and the line.
func parseTextFile(name, text string) (fileText, error) {
	lists, err := wordlist.Parse(text)
	if err != nil {
		return fileText{}, fmt.Errorf("%s.txt: %w", name, err)
	}
	file := fileText{name: name}
	for _, list := range lists {
		if list.Name() == wordsHeading {
			bank, err := parseBank(list)
			if err != nil {
				return fileText{}, fmt.Errorf("%s.txt:%d: %w", name, list.Line, err)
			}
			file.banks = append(file.banks, bank)
			continue
		}
		pool, err := parsePool(name, list)
		if err != nil {
			return fileText{}, err
		}
		file.pools = append(file.pools, pool)
	}
	return file, nil
}

func parseBank(list wordlist.List) (wordsText, error) {
	if len(list.Head) != 2 {
		return wordsText{}, fmt.Errorf("a word list heading is \"== words name\"")
	}
	if len(list.Entries) == 0 {
		return wordsText{}, fmt.Errorf("the word list %s is empty", list.Head[1])
	}
	return wordsText{name: list.Head[1], line: list.Line, words: list.Entries}, nil
}

// parsePool reads a pool heading and its sentences.
func parsePool(file string, list wordlist.List) (poolText, error) {
	pool := poolText{name: list.Name(), line: list.Line}
	fail := func(line int, format string, args ...any) error {
		return fmt.Errorf("%s.txt:%d: %s", file, line, fmt.Sprintf(format, args...))
	}
	for _, setting := range list.Head[1:] {
		key, value, _ := strings.Cut(setting, "=")
		names := strings.Split(value, ",")
		switch key {
		case "eras":
			for _, name := range names {
				bucket, ok := eraNamed(name)
				if !ok {
					return pool, fail(list.Line, "%s is not an era (ancient, feudal, industrial, digital, cosmic)", name)
				}
				pool.eras = append(pool.eras, bucket)
			}
		case "ages":
			pool.ages = names
		default:
			return pool, fail(list.Line, "%s is not a setting of a pool (eras= or ages=)", setting)
		}
	}
	if len(pool.ages) > 0 {
		if pool.eras != nil {
			return pool, fail(list.Line, "a pool is held to eras= or to ages=, not both")
		}
		eras, err := erasOfAges(pool.ages)
		if err != nil {
			return pool, fail(list.Line, "%v", err)
		}
		pool.eras = eras
	}
	if len(list.Entries) == 0 {
		return pool, fail(list.Line, "the pool %s has no sentences", pool.name)
	}
	pool.bucket = rowBucket(pool.name, pool.ages)
	for _, entry := range list.Entries {
		sentence, err := parseSentence(entry)
		if err != nil {
			return pool, fail(entry.Line, "%v", err)
		}
		pool.sentences = append(pool.sentences, sentence)
	}
	return pool, nil
}

// rowBucket is the part of a sentence's row id that names its pool: the pool's
// name without its catalog's prefix. A pool for one age ("harb_warn_age_iron_age")
// is named for the age alone.
func rowBucket(poolName string, ages []string) string {
	_, bucket := rowParts(poolName)
	if len(ages) > 0 && bucket == "age_"+ages[0] {
		return ages[0]
	}
	return bucket
}

// eraNamed finds an era by the name text files give it.
func eraNamed(name string) (era, bool) {
	for _, bucket := range eraBuckets {
		if bucket.suffix == name {
			return bucket.era, true
		}
	}
	return eraAncient, false
}

// eraTextName is the name text files give an era.
func eraTextName(which era) string {
	for _, bucket := range eraBuckets {
		if bucket.era == which {
			return bucket.suffix
		}
	}
	return ""
}

// erasOfAges is the eras the ages fall in, each once, in order of first sight.
func erasOfAges(ages []string) ([]era, error) {
	var eras []era
	for _, age := range ages {
		bucket, known := eraOf(age)
		if !known {
			return nil, fmt.Errorf("%s is not an age", age)
		}
		if !slices.Contains(eras, bucket) {
			eras = append(eras, bucket)
		}
	}
	return eras, nil
}

// parseSentence reads a line of the form "tags | text".
func parseSentence(entry wordlist.Entry) (sentenceLine, error) {
	tags, text, found := strings.Cut(entry.Text, tagSeparator)
	if !found {
		return sentenceLine{}, fmt.Errorf("a sentence is written \"tags | text\"")
	}
	sentence := sentenceLine{line: entry.Line, skel: skel{Text: strings.TrimSpace(text)}}
	sentence.from = entry.End - len(sentence.Text)
	sentence.to = entry.End
	given := map[string]bool{}
	for _, tag := range strings.Fields(tags) {
		if err := sentence.skel.addTag(tag, given); err != nil {
			return sentence, err
		}
	}
	switch {
	case sentence.Text == "":
		return sentence, fmt.Errorf("the sentence is empty")
	case sentence.Reg == regUnset:
		return sentence, fmt.Errorf("no register: the tags need plain, wry or joke")
	case sentence.Topic == "":
		return sentence, fmt.Errorf("no topic in the tags")
	case strings.Count(sentence.Text, slotMark) != boolToInt(sentence.Slot != ""):
		return sentence, fmt.Errorf("a sentence has a %s exactly when its tags name a slot=", slotMark)
	}
	return sentence, nil
}

func boolToInt(yes bool) int {
	if yes {
		return 1
	}
	return 0
}

// addTag applies one tag of a sentence. given records the kinds of tag already
// seen, so a second register or a second tone= is an error.
func (s *skel) addTag(tag string, given map[string]bool) error {
	kind, value := tagKind(tag)
	if kind == "" {
		return fmt.Errorf("%q is not a register, a form, a topic or a tone=, kind=, need= or slot= tag", tag)
	}
	if given[kind] {
		return fmt.Errorf("two %s tags", kind)
	}
	given[kind] = true
	values := strings.Split(value, ",")
	switch kind {
	case "register":
		s.Reg = registerNames[tag]
	case "form":
		s.Form = formNames[tag]
	case "topic":
		s.Topic = tag
	case "slot":
		s.Slot = value
	case "kind":
		s.Kinds = values
	case "tone":
		for _, name := range values {
			tone, ok := toneNames[name]
			if !ok {
				return fmt.Errorf("%s is not a tone (grim, triumphant, wry)", name)
			}
			s.Tones = append(s.Tones, tone)
		}
	case "need":
		for _, name := range values {
			which, ok := needNames[name]
			if !ok {
				return fmt.Errorf("%s is not a need (subject, res, mass, amount)", name)
			}
			s.Needs |= which
		}
	}
	return nil
}

// tagKind says what a tag is. A key=values tag is its key (tone, kind, need or
// slot) and its values; a bare word is a "register", a "form" or a "topic".
// The kind is "" for a tag that is none of them.
func tagKind(tag string) (kind, value string) {
	key, value, keyed := strings.Cut(tag, "=")
	if keyed {
		if slices.Contains([]string{"tone", "kind", "need", "slot"}, key) {
			return key, value
		}
		return "", ""
	}
	_, isRegister := registerNames[tag]
	_, isForm := formNames[tag]
	switch {
	case isRegister:
		return "register", ""
	case isForm:
		return "form", ""
	case knownTopics[tag]:
		return "topic", ""
	}
	return "", ""
}

// check holds the files to each other.
func (c *catalogText) check() error {
	poolNames := map[string]bool{}
	bankNames := map[string]bool{}
	for _, file := range c.files {
		isMomentFile := file.name != "harb_banks"
		if isMomentFile && len(file.pools) == 0 {
			return fmt.Errorf("%s.txt has no pools", file.name)
		}
		if !isMomentFile && len(file.pools) > 0 {
			return fmt.Errorf("%s.txt holds word lists only", file.name)
		}
		for _, pool := range file.pools {
			if key, _ := rowParts(pool.name); key == "" || !strings.HasPrefix(pool.name, file.name+"_") {
				return fmt.Errorf("%s.txt:%d: the pool %s does not start with %s_", file.name, pool.line, pool.name, file.name)
			}
			if poolNames[pool.name] {
				return fmt.Errorf("%s.txt:%d: the pool %s is already in another file", file.name, pool.line, pool.name)
			}
			poolNames[pool.name] = true
		}
		for _, bank := range file.banks {
			if bankNames[bank.name] {
				return fmt.Errorf("%s.txt:%d: the word list %s is already in another file", file.name, bank.line, bank.name)
			}
			bankNames[bank.name] = true
		}
	}
	for _, file := range c.files {
		for _, pool := range file.pools {
			for _, sentence := range pool.sentences {
				if sentence.Slot != "" && !bankNames[sentence.Slot] {
					return fmt.Errorf("%s.txt:%d: slot=%s names no word list", file.name, sentence.line, sentence.Slot)
				}
			}
		}
	}
	return nil
}

// banks gives every word list's entries by name.
func (c *catalogText) banks() map[string][]string {
	out := map[string][]string{}
	for _, file := range c.files {
		for _, bank := range file.banks {
			for _, word := range bank.words {
				out[bank.name] = append(out[bank.name], word.Text)
			}
		}
	}
	return out
}

// templates is a Moment's skeleton set: its pools in file order, which is the
// order ids and draws count by.
func (c *catalogText) templates(prefix string) []tmpl {
	var out []tmpl
	for _, file := range c.files {
		if file.name != prefix {
			continue
		}
		for _, pool := range file.pools {
			out = append(out, pool.templates()...)
		}
	}
	return out
}

// templates converts a pool's sentences into the renderer's form, numbering
// their ids from the pool's name.
//
// Ids are POSITIONAL within a pool. Nothing persists them but a harbinger's
// saved lines, so an insert in the middle renumbers its neighbours and the
// copy export's row ids; new sentences go at the end of a pool.
func (p poolText) templates() []tmpl {
	key, _ := rowParts(p.name)
	out := make([]tmpl, 0, len(p.sentences))
	for i, sentence := range p.sentences {
		t := sentence.skel.template(p.name+"_"+strconv.Itoa(i), p.eras)
		t.Ages = p.ages
		t.Row = sentenceRow(key, p.bucket, i)
		out = append(out, t)
	}
	return out
}
